/* ========================================================================
 * KuraNAS Stream Grabber — Background Service Worker (MV3)
 * ======================================================================== */

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

const DEFAULT_KURANAS_API_BASE = "http://localhost:8000/api/v1";
const DISCOVERY_HEALTH_SUFFIX = "/health";
const DISCOVERY_REQUEST_TIMEOUT_MS = 1500;
const DISCOVERY_TAB_HOST_LIMIT = 10;

const MEDIA_PATTERNS = [
  { regex: /\.m3u8(\?|$)/i, type: "hls" },
  { regex: /\.mpd(\?|$)/i, type: "dash" },
  { regex: /\.ts(\?|$)/i, type: "ts" },
  { regex: /\.mp4(\?|$)/i, type: "mp4" },
  { regex: /\.m4s(\?|$)/i, type: "m4s" },
  { regex: /\.aac(\?|$)/i, type: "aac" },
  { regex: /\.webm(\?|$)/i, type: "webm" },
];

const MEDIA_CONTENT_TYPES = [
  { pattern: /mpegurl/i, type: "hls" },
  { pattern: /dash\+xml/i, type: "dash" },
  { pattern: /^video\//i, type: "video" },
  { pattern: /^audio\//i, type: "audio" },
];

const HYBRID_STABILITY_MS = 200;
const HYBRID_STOP_GRACE_MS = 5000;

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

const detectedMedia = new Map();
const hybridStates = new Map();
const detectedTitles = new Map();

// ---------------------------------------------------------------------------
// 1. Media Detection via Network
// ---------------------------------------------------------------------------

function classifyByUrl(url) {
  for (const { regex, type } of MEDIA_PATTERNS) {
    if (regex.test(url)) return type;
  }
  return null;
}

function classifyByContentType(contentType) {
  for (const { pattern, type } of MEDIA_CONTENT_TYPES) {
    if (pattern.test(contentType)) return type;
  }
  return null;
}

function addMedia(tabId, item) {
  if (!detectedMedia.has(tabId)) {
    detectedMedia.set(tabId, []);
  }

  const list = detectedMedia.get(tabId);
  const isDuplicate = list.some(
    (m) => m.url === item.url && m.type === item.type
  );
  if (isDuplicate) return;

  list.push(item);
  updateBadge(tabId);

  chrome.runtime.sendMessage({ action: "media_detected", tabId, item }).catch(
    () => {}
  );
}

function updateBadge(tabId) {
  const list = detectedMedia.get(tabId) || [];
  const text = list.length > 0 ? String(list.length) : "";
  chrome.action.setBadgeText({ text, tabId }).catch(() => {});
  chrome.action.setBadgeBackgroundColor({ color: "#4CAF50", tabId }).catch(
    () => {}
  );
}

chrome.webRequest.onBeforeRequest.addListener(
  (details) => {
    if (details.tabId < 0) return;
    const type = classifyByUrl(details.url);
    if (type) {
      addMedia(details.tabId, {
        url: details.url,
        type,
        source: "network",
        timestamp: Date.now(),
      });
    }
  },
  { urls: ["<all_urls>"] }
);

chrome.webRequest.onHeadersReceived.addListener(
  (details) => {
    if (details.tabId < 0) return;
    const ctHeader = (details.responseHeaders || []).find(
      (h) => h.name.toLowerCase() === "content-type"
    );
    if (!ctHeader) return;

    const type = classifyByContentType(ctHeader.value);
    if (type) {
      addMedia(details.tabId, {
        url: details.url,
        type,
        source: "network",
        timestamp: Date.now(),
      });
    }
  },
  { urls: ["<all_urls>"] },
  ["responseHeaders"]
);

// ---------------------------------------------------------------------------
// 2. Message Router
// ---------------------------------------------------------------------------

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  const tabId = sender.tab ? sender.tab.id : msg.tabId;

  switch (msg.action) {
    case "blob_detected":
      handleBlobDetected(tabId, msg);
      break;

    case "title_detected":
      handleTitleDetected(tabId, msg);
      break;

    case "get_title":
      sendResponse(getTitleForTab(msg.tabId));
      return false;

    case "hybrid_video_state":
      handleHybridVideoState(tabId, msg.snapshot);
      break;

    case "get_media":
      sendResponse({ media: detectedMedia.get(msg.tabId) || [] });
      return false;

    case "hybrid_arm":
      armHybrid(msg.tabId);
      sendResponse({ ok: true });
      return false;

    case "hybrid_disarm":
      disarmHybrid(msg.tabId);
      sendResponse({ ok: true });
      return false;

    case "hybrid_stop_now":
      stopHybridRecording(msg.tabId);
      sendResponse({ ok: true });
      return false;

    case "hybrid_offscreen_started":
      handleOffscreenStarted(tabId, msg);
      break;

    case "hybrid_offscreen_stopped":
      handleOffscreenStopped(tabId, msg);
      break;

    case "hybrid_offscreen_error":
      handleOffscreenError(tabId, msg);
      break;

    case "hybrid_save_recording_blob":
      handleSaveRecordingBlob(msg);
      break;

    case "download_hls":
      downloadHLS(msg.url, msg.name).then(sendResponse);
      return true;

    case "download_dash":
      downloadDASH(msg.url, msg.name).then(sendResponse);
      return true;

    case "download_direct":
      downloadDirect(msg.url, msg.name).then(sendResponse);
      return true;

    case "upload_blob_capture":
      uploadBlobCapture(msg.tabId, msg.blobUrl, msg.name).then(sendResponse);
      return true;

    case "get_hybrid_status":
      sendResponse(getHybridStatus(msg.tabId));
      return false;
  }
});

// ---------------------------------------------------------------------------
// 3. Blob Detection
// ---------------------------------------------------------------------------

function handleBlobDetected(tabId, msg) {
  addMedia(tabId, {
    url: msg.blobUrl || "blob",
    type: "blob",
    mimeType: msg.mimeType,
    size: msg.size,
    source: "blob",
    timestamp: Date.now(),
  });
}

// ---------------------------------------------------------------------------
// 3b. Title Detection
// ---------------------------------------------------------------------------

function handleTitleDetected(tabId, msg) {
  if (!tabId || !msg.title) return;
  detectedTitles.set(tabId, {
    title: msg.title,
    source: msg.source,
    url: msg.url,
    hostname: msg.hostname,
    timestamp: Date.now(),
  });
}

function getTitleForTab(tabId) {
  const entry = detectedTitles.get(tabId);
  if (entry) return { title: entry.title, source: entry.source };
  return { title: null, source: null };
}

// ---------------------------------------------------------------------------
// 4. Hybrid State Machine
// ---------------------------------------------------------------------------

function getHybridStatus(tabId) {
  const state = hybridStates.get(tabId);
  if (!state) return { armed: false, state: "IDLE" };
  return {
    armed: state.armed,
    state: state.recordingState,
    monitorEnabled: state.monitorEnabled,
  };
}

function armHybrid(tabId) {
  let state = hybridStates.get(tabId);
  if (!state) {
    state = {
      armed: true,
      monitorEnabled: true,
      recordingState: "ARMED",
      stabilityTimer: null,
      graceTimer: null,
      lastSnapshot: null,
    };
    hybridStates.set(tabId, state);
  } else {
    state.armed = true;
    state.monitorEnabled = true;
    state.recordingState = "ARMED";
  }

  chrome.tabs
    .sendMessage(tabId, { action: "hybrid_monitor_start" })
    .catch(() => {});
  broadcastHybridStatus(tabId);
}

function disarmHybrid(tabId) {
  const state = hybridStates.get(tabId);
  if (!state) return;

  clearTimeout(state.stabilityTimer);
  clearTimeout(state.graceTimer);
  state.armed = false;
  state.monitorEnabled = false;
  state.recordingState = "IDLE";

  if (state.recording) {
    stopOffscreenRecording(tabId);
  }

  chrome.tabs
    .sendMessage(tabId, { action: "hybrid_monitor_stop" })
    .catch(() => {});
  broadcastHybridStatus(tabId);
}

function handleHybridVideoState(tabId, snapshot) {
  const state = hybridStates.get(tabId);
  if (!state || !state.armed) return;

  state.lastSnapshot = snapshot;

  const shouldRecord =
    snapshot.hasVideo &&
    snapshot.isPlaying &&
    snapshot.isFullscreen &&
    !snapshot.isEnded;

  if (state.recordingState === "ARMED") {
    if (shouldRecord) {
      clearTimeout(state.stabilityTimer);
      state.stabilityTimer = setTimeout(() => {
        if (state.armed && state.recordingState === "ARMED") {
          startHybridRecording(tabId);
        }
      }, HYBRID_STABILITY_MS);
    } else {
      clearTimeout(state.stabilityTimer);
    }
  } else if (state.recordingState === "RECORDING") {
    if (snapshot.isEnded) {
      clearTimeout(state.graceTimer);
      stopHybridRecording(tabId);
    } else if (!shouldRecord) {
      if (!state.graceTimer) {
        state.graceTimer = setTimeout(() => {
          if (state.recordingState === "RECORDING") {
            stopHybridRecording(tabId);
          }
        }, HYBRID_STOP_GRACE_MS);
      }
    } else {
      clearTimeout(state.graceTimer);
      state.graceTimer = null;
    }
  }

  if (state.lastUrl && snapshot.url !== state.lastUrl) {
    if (state.recordingState === "RECORDING") {
      stopHybridRecording(tabId);
    }
  }
  state.lastUrl = snapshot.url;
}

async function startHybridRecording(tabId) {
  const state = hybridStates.get(tabId);
  if (!state) return;
  state.recordingState = "RECORDING";
  broadcastHybridStatus(tabId);

  try {
    const streamId = await chrome.tabCapture.getMediaStreamId({
      targetTabId: tabId,
    });
    await ensureOffscreen();
    chrome.runtime.sendMessage({
      action: "offscreen_start_recording",
      tabId,
      streamId,
    });
  } catch (err) {
    state.recordingState = "ARMED";
    broadcastHybridStatus(tabId);
  }
}

function stopHybridRecording(tabId) {
  const state = hybridStates.get(tabId);
  if (!state) return;

  clearTimeout(state.stabilityTimer);
  clearTimeout(state.graceTimer);
  state.graceTimer = null;
  state.recordingState = "STOPPED";
  broadcastHybridStatus(tabId);

  stopOffscreenRecording(tabId);

  setTimeout(() => {
    if (state.armed) {
      state.recordingState = "ARMED";
      broadcastHybridStatus(tabId);
    }
  }, 1000);
}

function stopOffscreenRecording(tabId) {
  chrome.runtime
    .sendMessage({ action: "offscreen_stop_recording", tabId })
    .catch(() => {});
}

function handleOffscreenStarted(tabId) {
  const state = hybridStates.get(tabId);
  if (state) {
    state.recording = true;
  }
}

function handleOffscreenStopped(tabId) {
  const state = hybridStates.get(tabId);
  if (state) {
    state.recording = false;
  }
}

function handleOffscreenError(tabId, msg) {
  const state = hybridStates.get(tabId);
  if (state) {
    state.recording = false;
    state.recordingState = "ARMED";
    broadcastHybridStatus(tabId);
  }
}

function broadcastHybridStatus(tabId) {
  const status = getHybridStatus(tabId);
  chrome.runtime
    .sendMessage({ action: "hybrid_status", tabId, status })
    .catch(() => {});
}

// ---------------------------------------------------------------------------
// 5. Offscreen Document Management
// ---------------------------------------------------------------------------

let creatingOffscreen = null;

async function ensureOffscreen() {
  const contexts = await chrome.runtime.getContexts({
    contextTypes: ["OFFSCREEN_DOCUMENT"],
  });
  if (contexts.length > 0) return;

  if (creatingOffscreen) {
    await creatingOffscreen;
    return;
  }

  creatingOffscreen = chrome.offscreen.createDocument({
    url: "offscreen/recorder.html",
    reasons: ["USER_MEDIA"],
    justification: "Tab capture recording for media grabbing",
  });

  await creatingOffscreen;
  creatingOffscreen = null;
}

// ---------------------------------------------------------------------------
// 6. Save Recording Blob → Upload to KuraNAS
// ---------------------------------------------------------------------------

async function handleSaveRecordingBlob(msg) {
  try {
    const response = await fetch(msg.blobUrl);
    const blob = await response.blob();

    const name = msg.name || `recording_${Date.now()}`;
    await uploadToKuraNAS(blob, name, "recording");
  } catch (err) {
    // Blob URL may have been revoked
  }
}

// ---------------------------------------------------------------------------
// 7. HLS Download & Upload
// ---------------------------------------------------------------------------

async function downloadHLS(manifestUrl, name) {
  try {
    const resp = await fetch(manifestUrl);
    const text = await resp.text();

    if (text.includes("#EXT-X-STREAM-INF")) {
      return parseHLSMasterPlaylist(text, manifestUrl);
    }

    return await downloadHLSMediaPlaylist(text, manifestUrl, name);
  } catch (err) {
    return { error: err.message };
  }
}

function parseHLSMasterPlaylist(text, baseUrl) {
  const lines = text.split("\n");
  const variants = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line.startsWith("#EXT-X-STREAM-INF:")) continue;

    const attrs = line.substring(18);
    const bandwidthMatch = attrs.match(/BANDWIDTH=(\d+)/);
    const resolutionMatch = attrs.match(/RESOLUTION=([^\s,]+)/);
    const codecsMatch = attrs.match(/CODECS="([^"]+)"/);

    const nextLine = (lines[i + 1] || "").trim();
    if (!nextLine || nextLine.startsWith("#")) continue;

    variants.push({
      url: resolveUrl(baseUrl, nextLine),
      bandwidth: bandwidthMatch ? parseInt(bandwidthMatch[1], 10) : 0,
      resolution: resolutionMatch ? resolutionMatch[1] : "",
      codecs: codecsMatch ? codecsMatch[1] : "",
    });
    i++;
  }

  return { type: "master", variants };
}

async function downloadHLSMediaPlaylist(text, baseUrl, name) {
  const lines = text.split("\n");
  const segmentUrls = [];

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    segmentUrls.push(resolveUrl(baseUrl, trimmed));
  }

  const chunks = [];
  let totalSize = 0;

  for (const url of segmentUrls) {
    const resp = await fetch(url);
    const buf = await resp.arrayBuffer();
    chunks.push(new Uint8Array(buf));
    totalSize += buf.byteLength;
  }

  const merged = new Uint8Array(totalSize);
  let offset = 0;
  for (const chunk of chunks) {
    merged.set(chunk, offset);
    offset += chunk.byteLength;
  }

  const blob = new Blob([merged], { type: "video/mp2t" });
  const captureName = name || `stream_hls_${Date.now()}`;

  await uploadToKuraNAS(blob, captureName, "hls");
  return { ok: true, name: captureName };
}

// ---------------------------------------------------------------------------
// 8. DASH Download & Upload
// ---------------------------------------------------------------------------

async function downloadDASH(manifestUrl, name) {
  try {
    const resp = await fetch(manifestUrl);
    const text = await resp.text();

    const parser = new DOMParser();
    const doc = parser.parseFromString(text, "application/xml");
    const representations = [];

    doc.querySelectorAll("Representation").forEach((rep) => {
      const adaptationSet = rep.closest("AdaptationSet");
      const mimeType = rep.getAttribute("mimeType") ||
        (adaptationSet ? adaptationSet.getAttribute("mimeType") : "") || "";

      representations.push({
        id: rep.getAttribute("id") || "",
        bandwidth: parseInt(rep.getAttribute("bandwidth") || "0", 10),
        width: parseInt(rep.getAttribute("width") || "0", 10),
        height: parseInt(rep.getAttribute("height") || "0", 10),
        codecs: rep.getAttribute("codecs") || "",
        mimeType,
        manifestUrl,
      });
    });

    if (representations.length > 1) {
      return { type: "dash_manifest", representations };
    }

    if (representations.length === 1) {
      return await downloadDASHRepresentation(
        manifestUrl,
        text,
        representations[0].id,
        name
      );
    }

    return { error: "No representations found" };
  } catch (err) {
    return { error: err.message };
  }
}

async function downloadDASHRepresentation(manifestUrl, manifestText, repId, name) {
  const parser = new DOMParser();
  const doc = parser.parseFromString(manifestText, "application/xml");
  const rep = repId
    ? doc.querySelector(`Representation[id="${repId}"]`)
    : doc.querySelector("Representation");

  if (!rep) return { error: "Representation not found" };

  const segmentUrls = collectDASHSegmentUrls(rep, manifestUrl);

  const chunks = [];
  let totalSize = 0;

  for (const url of segmentUrls) {
    const resp = await fetch(url);
    const buf = await resp.arrayBuffer();
    chunks.push(new Uint8Array(buf));
    totalSize += buf.byteLength;
  }

  const merged = new Uint8Array(totalSize);
  let offset = 0;
  for (const chunk of chunks) {
    merged.set(chunk, offset);
    offset += chunk.byteLength;
  }

  const mimeType = rep.getAttribute("mimeType") || "video/mp4";
  const ext = mimeType.includes("audio") ? "m4a" : "mp4";
  const blob = new Blob([merged], { type: mimeType });
  const captureName = name || `stream_dash_${Date.now()}`;

  await uploadToKuraNAS(blob, captureName, "dash");
  return { ok: true, name: captureName, ext };
}

function collectDASHSegmentUrls(rep, manifestUrl) {
  const urls = [];
  const adaptationSet = rep.closest("AdaptationSet");
  const period = rep.closest("Period");

  const segTemplate =
    rep.querySelector("SegmentTemplate") ||
    (adaptationSet ? adaptationSet.querySelector("SegmentTemplate") : null);

  if (segTemplate) {
    const timeline = segTemplate.querySelector("SegmentTimeline");
    const initTemplate = segTemplate.getAttribute("initialization") || "";
    const mediaTemplate = segTemplate.getAttribute("media") || "";
    const startNumber = parseInt(
      segTemplate.getAttribute("startNumber") || "1",
      10
    );
    const timescale = parseInt(
      segTemplate.getAttribute("timescale") || "1",
      10
    );
    const repId = rep.getAttribute("id") || "";
    const bandwidth = rep.getAttribute("bandwidth") || "";

    if (initTemplate) {
      urls.push(
        resolveUrl(
          manifestUrl,
          expandDASHTemplate(initTemplate, repId, bandwidth, 0, 0)
        )
      );
    }

    if (timeline) {
      let number = startNumber;
      let time = 0;
      const entries = timeline.querySelectorAll("S");

      for (const s of entries) {
        const t = s.getAttribute("t");
        if (t !== null) time = parseInt(t, 10);
        const d = parseInt(s.getAttribute("d") || "0", 10);
        const r = parseInt(s.getAttribute("r") || "0", 10);

        for (let i = 0; i <= r; i++) {
          urls.push(
            resolveUrl(
              manifestUrl,
              expandDASHTemplate(mediaTemplate, repId, bandwidth, number, time)
            )
          );
          number++;
          time += d;
        }
      }
    } else {
      const duration = parseFloat(
        segTemplate.getAttribute("duration") || "0"
      );
      const periodDuration = parseDuration(
        (period ? period.getAttribute("duration") : null) || ""
      );
      if (duration > 0 && periodDuration > 0) {
        const segCount = Math.ceil(
          (periodDuration * timescale) / duration
        );
        for (let i = 0; i < segCount; i++) {
          urls.push(
            resolveUrl(
              manifestUrl,
              expandDASHTemplate(
                mediaTemplate,
                repId,
                bandwidth,
                startNumber + i,
                i * duration
              )
            )
          );
        }
      }
    }
  } else {
    const segList =
      rep.querySelector("SegmentList") ||
      (adaptationSet ? adaptationSet.querySelector("SegmentList") : null);

    if (segList) {
      const init = segList.querySelector("Initialization");
      if (init) {
        urls.push(resolveUrl(manifestUrl, init.getAttribute("sourceURL")));
      }
      segList.querySelectorAll("SegmentURL").forEach((seg) => {
        urls.push(resolveUrl(manifestUrl, seg.getAttribute("media")));
      });
    } else {
      const baseUrl = rep.querySelector("BaseURL") ||
        (adaptationSet ? adaptationSet.querySelector("BaseURL") : null);
      if (baseUrl) {
        urls.push(resolveUrl(manifestUrl, baseUrl.textContent.trim()));
      }
    }
  }

  return urls;
}

function expandDASHTemplate(template, repId, bandwidth, number, time) {
  let result = template;
  result = result.replace(/\$RepresentationID\$/g, repId);
  result = result.replace(/\$Bandwidth\$/g, bandwidth);
  result = result.replace(/\$Time\$/g, String(time));

  result = result.replace(/\$Number(%(\d+)d)?\$/g, (_, fmt, width) => {
    if (width) return String(number).padStart(parseInt(width, 10), "0");
    return String(number);
  });

  return result;
}

function parseDuration(iso) {
  if (!iso) return 0;
  const m = iso.match(
    /PT(?:(\d+(?:\.\d+)?)H)?(?:(\d+(?:\.\d+)?)M)?(?:(\d+(?:\.\d+)?)S)?/
  );
  if (!m) return 0;
  return (
    (parseFloat(m[1] || "0") * 3600) +
    (parseFloat(m[2] || "0") * 60) +
    parseFloat(m[3] || "0")
  );
}

// ---------------------------------------------------------------------------
// 9. Direct Download & Upload
// ---------------------------------------------------------------------------

async function downloadDirect(url, name) {
  try {
    const resp = await fetch(url);
    const blob = await resp.blob();
    const captureName = name || `direct_${Date.now()}`;
    await uploadToKuraNAS(blob, captureName, "direct");
    return { ok: true, name: captureName };
  } catch (err) {
    return { error: err.message };
  }
}

// ---------------------------------------------------------------------------
// 10. Blob Capture Upload
// ---------------------------------------------------------------------------

async function uploadBlobCapture(tabId, blobUrl, name) {
  try {
    const resp = await fetch(blobUrl);
    const blob = await resp.blob();
    const captureName = name || `blob_${Date.now()}`;
    await uploadToKuraNAS(blob, captureName, "blob");
    return { ok: true, name: captureName };
  } catch (err) {
    return { error: err.message };
  }
}

// ---------------------------------------------------------------------------
// 11. Upload to KuraNAS Backend
// ---------------------------------------------------------------------------

async function uploadToKuraNAS(blob, name, mediaType) {
  const formData = new FormData();
  const ext = guessExtension(blob.type, mediaType);
  const fileName = `${sanitizeFileName(name)}.${ext}`;

  formData.append("file", blob, fileName);
  formData.append("name", name);
  formData.append("media_type", mediaType);
  formData.append("mime_type", blob.type || "application/octet-stream");
  formData.append("size", String(blob.size));

  const apiUrl = await getApiBaseUrl();

  const resp = await fetch(`${apiUrl}/captures/upload`, {
    method: "POST",
    body: formData,
  });

  if (!resp.ok) {
    const body = await resp.text();
    throw new Error(`Upload failed (${resp.status}): ${body}`);
  }

  return resp.json();
}

async function getApiBaseUrl() {
  const configured = await getConfiguredApiBaseUrl();
  if (configured) {
    if (await isKuraNASApiReachable(configured)) {
      return configured;
    }

    const discoveredFromConfigured = await discoverApiBaseUrl(configured);
    if (discoveredFromConfigured) {
      await persistApiBaseUrl(discoveredFromConfigured);
      return discoveredFromConfigured;
    }

    return configured;
  }

  const discovered = await discoverApiBaseUrl();
  if (discovered) {
    await persistApiBaseUrl(discovered);
    return discovered;
  }

  return DEFAULT_KURANAS_API_BASE;
}

async function getConfiguredApiBaseUrl() {
  try {
    const result = await chrome.storage.sync.get("apiBaseUrl");
    return normalizeApiBaseUrl(result.apiBaseUrl || "");
  } catch {
    return "";
  }
}

async function persistApiBaseUrl(apiBaseUrl) {
  try {
    await chrome.storage.sync.set({ apiBaseUrl });
  } catch {
    // Keep runtime behavior even if sync storage is unavailable.
  }
}

async function discoverApiBaseUrl(configuredApiBaseUrl = "") {
  const candidates = await buildApiBaseCandidates(configuredApiBaseUrl);
  if (candidates.length === 0) return null;

  const attempts = candidates.map(async (candidate) => {
    const reachable = await isKuraNASApiReachable(candidate);
    if (!reachable) {
      throw new Error("unreachable");
    }
    return candidate;
  });

  try {
    return await Promise.any(attempts);
  } catch {
    return null;
  }
}

async function buildApiBaseCandidates(configuredApiBaseUrl) {
  const candidates = new Set();
  const addCandidate = (url) => {
    const normalized = normalizeApiBaseUrl(url);
    if (normalized) {
      candidates.add(normalized);
    }
  };

  addCandidate(configuredApiBaseUrl);
  addCandidate(DEFAULT_KURANAS_API_BASE);
  addCandidate("http://127.0.0.1:8000/api/v1");
  addCandidate("http://kuranas.local:8000/api/v1");

  const tabHosts = await getLikelyBackendHostsFromTabs();
  for (const host of tabHosts) {
    addCandidate(`http://${host}:8000/api/v1`);
  }

  return Array.from(candidates);
}

async function getLikelyBackendHostsFromTabs() {
  try {
    const tabs = await chrome.tabs.query({});
    const hosts = new Set();

    for (const tab of tabs) {
      if (!tab.url) continue;

      let parsed;
      try {
        parsed = new URL(tab.url);
      } catch {
        continue;
      }

      if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
        continue;
      }

      const hostname = parsed.hostname || "";
      if (!isLikelyLanHost(hostname)) {
        continue;
      }

      hosts.add(hostname);
      if (hosts.size >= DISCOVERY_TAB_HOST_LIMIT) {
        break;
      }
    }

    return Array.from(hosts);
  } catch {
    return [];
  }
}

function isLikelyLanHost(hostname) {
  if (!hostname) return false;
  if (hostname === "localhost") return true;
  if (hostname.endsWith(".local")) return true;

  const ipv4Parts = hostname.split(".");
  if (ipv4Parts.length !== 4) return false;

  const nums = ipv4Parts.map((part) => Number(part));
  if (nums.some((n) => Number.isNaN(n) || n < 0 || n > 255)) {
    return false;
  }

  if (nums[0] === 10) return true;
  if (nums[0] === 172 && nums[1] >= 16 && nums[1] <= 31) return true;
  if (nums[0] === 192 && nums[1] === 168) return true;
  if (nums[0] === 127) return true;

  return false;
}

function normalizeApiBaseUrl(url) {
  if (!url || typeof url !== "string") return "";

  const trimmed = url.trim().replace(/\/+$/g, "");
  if (!trimmed) return "";

  if (/\/api\/v1$/i.test(trimmed)) {
    return trimmed;
  }

  return `${trimmed}/api/v1`;
}

async function isKuraNASApiReachable(apiBaseUrl) {
  const healthUrl = `${apiBaseUrl}${DISCOVERY_HEALTH_SUFFIX}`;
  try {
    const response = await fetchWithTimeout(
      healthUrl,
      { method: "GET", cache: "no-store" },
      DISCOVERY_REQUEST_TIMEOUT_MS
    );

    if (!response.ok) {
      return false;
    }

    const body = (await response.text()).toLowerCase();
    return body.includes("kuranas");
  } catch {
    return false;
  }
}

async function fetchWithTimeout(url, options, timeoutMs) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  try {
    return await fetch(url, { ...options, signal: controller.signal });
  } finally {
    clearTimeout(timeoutId);
  }
}

function sanitizeFileName(name) {
  return name
    .replace(/[<>:"/\\|?*\x00-\x1f]/g, "_")
    .replace(/\s+/g, "_")
    .substring(0, 200);
}

function guessExtension(mimeType, mediaType) {
  if (mimeType) {
    if (mimeType.includes("mp2t")) return "ts";
    if (mimeType.includes("mp4")) return "mp4";
    if (mimeType.includes("webm")) return "webm";
    if (mimeType.includes("m4a") || mimeType.includes("x-m4a")) return "m4a";
    if (mimeType.includes("aac")) return "aac";
    if (mimeType.includes("mpeg") && mimeType.includes("audio")) return "mp3";
  }
  if (mediaType === "hls") return "ts";
  if (mediaType === "dash") return "mp4";
  return "bin";
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

function resolveUrl(base, relative) {
  if (!relative) return base;
  try {
    return new URL(relative, base).href;
  } catch {
    const basePath = base.substring(0, base.lastIndexOf("/") + 1);
    return basePath + relative;
  }
}

// ---------------------------------------------------------------------------
// Tab Cleanup
// ---------------------------------------------------------------------------

chrome.tabs.onRemoved.addListener((tabId) => {
  detectedMedia.delete(tabId);
  detectedTitles.delete(tabId);
  const state = hybridStates.get(tabId);
  if (state) {
    clearTimeout(state.stabilityTimer);
    clearTimeout(state.graceTimer);
    if (state.recording) {
      stopOffscreenRecording(tabId);
    }
    hybridStates.delete(tabId);
  }
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo) => {
  if (changeInfo.url) {
    detectedMedia.delete(tabId);
    detectedTitles.delete(tabId);
    updateBadge(tabId);
    // Ask content script to re-detect title for new page
    chrome.tabs.sendMessage(tabId, { action: "request_title" }).catch(() => {});
  }
});

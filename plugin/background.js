/* ========================================================================
 * KuraNAS Stream Grabber — Background Service Worker (MV3)
 * ======================================================================== */

import {
  DEFAULT_KURANAS_API_BASE,
  DISCOVERY_HEALTH_SUFFIX,
  DISCOVERY_REQUEST_TIMEOUT_MS,
  DISCOVERY_TAB_HOST_LIMIT,
  HYBRID_PREPARE_SETTLE_MS,
  HYBRID_STABILITY_MS,
  HYBRID_STOP_GRACE_MS,
  MEDIA_CONTENT_TYPES,
  MEDIA_PATTERNS,
} from "./src/shared/constants.js";
import {
  guessExtension,
  resolveUrl,
  sanitizeFileName,
  wait,
} from "./src/shared/utils.js";
import { createMediaDetectionManager } from "./src/background/media-detection.js";
import { routeRuntimeMessage } from "./src/background/message-router.js";
import { createUploader } from "./src/background/uploader.js";
import { createDownloader } from "./src/background/downloader.js";
import { createFetcher } from "./src/background/fetcher.js";
import { createHybridStateMachine } from "./src/background/hybrid-state.js";

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

const detectedMedia = new Map();
const hybridStates = new Map();
const detectedTitles = new Map();
const detectedMetadata = new Map();

// Load probe: this prints the moment the service worker starts. If you reload
// the extension (chrome://extensions -> reload) and open the "service worker"
// console, seeing this line proves the new code is the one running.
console.log(
  `[KuraNAS bg] service worker carregado — ${
    chrome.runtime.getManifest().version_name || chrome.runtime.getManifest().version
  }`,
  new Date().toISOString()
);

const uploader = createUploader({
  getApiBaseUrl,
  guessExtension,
  sanitizeFileName,
  waitFn: wait,
});
const {
  handleSaveRecordingBlob,
  uploadBlobCapture,
  uploadChunkWithRetry,
  uploadToKuraNAS,
} = uploader;

const downloader = createDownloader({
  resolveUrl,
  uploadToKuraNAS,
});
const {
  downloadDASH,
  downloadDirect,
  downloadHLS,
} = downloader;

const fetcher = createFetcher({ getApiBaseUrl });
const {
  submitFetch,
  listTargets: listIngestTargets,
  listPresets: listIngestPresets,
} = fetcher;

const hybridStateMachine = createHybridStateMachine({
  hybridStates,
  broadcastHybridStatus,
  initHybridUploadSession,
  ensureOffscreen,
  getMediaStreamId: (tabId) => chrome.tabCapture.getMediaStreamId({
    targetTabId: tabId,
  }),
  sendRuntimeMessage: (message) => {
    chrome.runtime.sendMessage(message).catch(() => {});
  },
  sendTabMessage: (tabId, message) => {
    chrome.tabs.sendMessage(tabId, message).catch(() => {});
  },
  stopOffscreenRecording,
  hybridStabilityMs: HYBRID_STABILITY_MS,
  hybridStopGraceMs: HYBRID_STOP_GRACE_MS,
  hybridPrepareSettleMs: HYBRID_PREPARE_SETTLE_MS,
});
const {
  armHybrid,
  cleanupTab,
  disarmHybrid,
  getHybridStatus,
  handleHybridPrepared,
  handleHybridVideoState,
  handleOffscreenError,
  handleOffscreenStarted,
  handleOffscreenStopped,
  stopHybridRecording,
} = hybridStateMachine;

// ---------------------------------------------------------------------------
// 1. Media Detection via Network
// ---------------------------------------------------------------------------

const mediaDetectionManager = createMediaDetectionManager({
  chromeApi: chrome,
  detectedMedia,
  mediaPatterns: MEDIA_PATTERNS,
  mediaContentTypes: MEDIA_CONTENT_TYPES,
});
const addMedia = mediaDetectionManager.addMedia;
const updateBadge = mediaDetectionManager.updateBadge;
mediaDetectionManager.registerNetworkListeners();

// ---------------------------------------------------------------------------
// 2. Message Router
// ---------------------------------------------------------------------------

// Central log relay: the popup and the (hidden) offscreen document forward
// their console output here via { action: "kuranas_log" }, so EVERY plugin log
// also shows up in the one console that is trivial to open — the service
// worker's. This is purely additive; the relay never answers the message.
chrome.runtime.onMessage.addListener((msg) => {
  if (msg && msg.action === "kuranas_log") {
    console.log(msg.source || "[KuraNAS]", ...(Array.isArray(msg.args) ? msg.args : []));
  }
  return false;
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => routeRuntimeMessage(
  msg,
  sender,
  sendResponse,
  {
    armHybrid,
    disarmHybrid,
    downloadDASH,
    downloadDirect,
    downloadHLS,
    getDetectedMedia: (tabId) => detectedMedia.get(tabId) || [],
    getHybridStatus,
    getMetadataForTab,
    getTitleForTab,
    handleBlobDetected,
    handleHybridPrepared,
    handleHybridRecordingChunk,
    handleHybridRecordingComplete,
    handleHybridVideoState,
    handleMetadataDetected,
    handleOffscreenError,
    handleOffscreenStarted,
    handleOffscreenStopped,
    handleSaveRecordingBlob,
    handleTitleDetected,
    listIngestPresets,
    listIngestTargets,
    stopHybridRecording,
    submitFetch,
    uploadBlobCapture,
  }
));

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
// 3c. Metadata Detection
// ---------------------------------------------------------------------------

function handleMetadataDetected(tabId, msg) {
  if (!tabId || !msg.metadata) return;
  detectedMetadata.set(tabId, { metadata: msg.metadata, timestamp: Date.now() });
}

function getMetadataForTab(tabId) {
  const entry = detectedMetadata.get(tabId);
  return entry ? entry.metadata : null;
}

// ---------------------------------------------------------------------------
// 4. Hybrid State Machine
// ---------------------------------------------------------------------------

function stopOffscreenRecording(tabId) {
  chrome.runtime
    .sendMessage({ action: "offscreen_stop_recording", tabId })
    .catch(() => {});
}

// Resolve the capture name from the page title. Title detection is async (and
// the title is cleared on SPA navigation), so if it is not ready yet, ask the
// page to re-detect and wait briefly for it before falling back to a timestamp.
async function resolveCaptureName(tabId) {
  const current = getTitleForTab(tabId);
  if (current && current.title && current.title.trim()) {
    return current.title.trim();
  }

  chrome.tabs.sendMessage(tabId, { action: "request_title" }).catch(() => {});
  for (let i = 0; i < 15; i++) {
    await wait(100);
    const t = getTitleForTab(tabId);
    if (t && t.title && t.title.trim()) {
      logBg(tabId, `título resolvido após ${(i + 1) * 100}ms`);
      return t.title.trim();
    }
  }

  logBg(tabId, "nenhum título detectado — usando nome com timestamp");
  return `recording_${tabId}_${Date.now()}`;
}

// Resolve the standardized metadata for the capture. Like the title, it is
// detected asynchronously and reset on SPA navigation, so ask the page to
// re-detect and wait briefly for a fresh object before giving up (null is fine —
// a capture may legitimately carry no metadata).
async function resolveCaptureMetadata(tabId) {
  const current = getMetadataForTab(tabId);
  if (current) return current;

  chrome.tabs.sendMessage(tabId, { action: "request_metadata" }).catch(() => {});
  for (let i = 0; i < 15; i++) {
    await wait(100);
    const meta = getMetadataForTab(tabId);
    if (meta) {
      logBg(tabId, `metadados resolvidos após ${(i + 1) * 100}ms`);
      return meta;
    }
  }
  return null;
}

async function initHybridUploadSession(tabId) {
  const state = hybridStates.get(tabId);
  if (!state) {
    throw new Error("Hybrid state not initialized");
  }
  if (state.uploadSession && state.uploadSession.uploadID) {
    return state.uploadSession;
  }

  const apiUrl = await getApiBaseUrl();
  // Name the capture after the detected page title (show + episode, e.g.
  // "Anime - S1 E2 - Título") so the file says which episode it is and episode 2
  // does not overwrite episode 1. Fall back to a timestamped name when no title
  // could be resolved.
  const name = await resolveCaptureName(tabId);
  const mimeType = "video/webm";
  const fileName = `${sanitizeFileName(name)}.webm`;
  logBg(tabId, `nome da captura: "${name}"`);

  // Standardized metadata (title, episode, duration, origin, …) is persisted by
  // the server as metadata.json beside the recording. The episode_key (when the
  // page yields a stable per-episode id) also drives the server's idempotency, so
  // re-arming on the same episode resumes instead of recording a duplicate.
  const metadata = await resolveCaptureMetadata(tabId);
  const episodeKey = (metadata && metadata.episode_key) || "";
  if (metadata) {
    logBg(tabId, `metadados: episode_key="${episodeKey}", plataforma="${metadata.platform || "?"}"`);
  }

  const initResp = await fetch(`${apiUrl}/captures/upload/init`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      name,
      media_type: "recording",
      mime_type: mimeType,
      size: 0,
      file_name: fileName,
      episode_key: episodeKey,
      metadata: metadata || undefined,
    }),
  });

  if (!initResp.ok) {
    const body = await initResp.text();
    throw new Error(`Init hybrid upload failed (${initResp.status}): ${body}`);
  }

  const payload = await initResp.json();

  // This episode is already archived (episode_key idempotency): there is no
  // upload_id and nothing to record. Throw so the state machine reverts to ARMED
  // without starting the offscreen recorder (it will move on at the next episode).
  if (payload.already_complete) {
    logBg(tabId, `episódio já capturado (episode_key="${episodeKey}") — pulando gravação`);
    throw new Error("capture already complete for this episode");
  }

  if (!payload.upload_id) {
    throw new Error("Invalid hybrid upload init response: upload_id is required");
  }

  // On a resumed session the server already holds received_size bytes; chunks
  // must continue from that offset or the very first chunk fails on a mismatch.
  const startOffset = Number(payload.received_size || 0);

  state.uploadSession = {
    apiUrl,
    uploadID: payload.upload_id,
    offset: startOffset,
    chunkIndex: 0,
    pending: Promise.resolve(),
    failed: false,
    completed: false,
  };

  logBg(tabId, `upload session HYBRID criada: uploadID=${payload.upload_id}`);
  return state.uploadSession;
}

function logBg(tabId, ...args) {
  console.log("[KuraNAS bg]", `tab=${tabId}`, ...args);
}

// Logged-once guard so the per-chunk handler (called ~hundreds of times) emits a
// single line for a dropped-chunk reason instead of flooding the console.
const chunkDropLogged = new Set();

// Streamed chunks arrive as a blob: URL string (a Blob cannot cross
// chrome.runtime messaging). Fetch it here to recover the real bytes, then
// revoke so a long recording does not leak blob URLs.
async function fetchChunkBlob(chunkUrl) {
  const resp = await fetch(chunkUrl);
  const blob = await resp.blob();
  // URL.revokeObjectURL does NOT exist in the MV3 service worker (it threw
  // "is not a function" and failed the whole upload). The bytes are already read
  // above, so ask the offscreen document — which created the blob URL and where
  // URL.revokeObjectURL does exist — to revoke it.
  chrome.runtime
    .sendMessage({ action: "offscreen_revoke_url", url: chunkUrl })
    .catch(() => {});
  return blob;
}

function handleHybridRecordingChunk(tabId, msg) {
  const state = hybridStates.get(tabId);
  if (!state || !state.uploadSession || state.uploadSession.failed) {
    if (!chunkDropLogged.has(tabId)) {
      chunkDropLogged.add(tabId);
      logBg(
        tabId,
        "chunk DESCARTADO (hybrid): ",
        !state ? "sem state" : !state.uploadSession ? "sem uploadSession (init não rodou?)" : "session.failed"
      );
    }
    return;
  }
  if (!msg.chunkUrl) return;

  const session = state.uploadSession;
  const { chunkUrl } = msg;
  session.pending = session.pending
    .then(async () => {
      if (session.failed || session.completed) return;
      const chunkBlob = await fetchChunkBlob(chunkUrl);
      if (!chunkBlob.size) {
        logBg(tabId, "chunk vazio após fetch da blob URL (SW não leu a blob?)");
        return;
      }
      await uploadChunkWithRetry(
        session.apiUrl,
        session.uploadID,
        chunkBlob,
        session.offset,
        session.chunkIndex
      );
      session.offset += chunkBlob.size;
      session.chunkIndex += 1;
      if (session.chunkIndex === 1 || session.chunkIndex % 25 === 0) {
        logBg(tabId, `chunk #${session.chunkIndex} enviado (offset ${session.offset} bytes)`);
      }
    })
    .catch((err) => {
      session.failed = true;
      console.error("[KuraNAS bg]", `tab=${tabId}`, "falha no chunk (hybrid):", err && err.message);
      chrome.runtime
        .sendMessage({
          action: "hybrid_upload_error",
          tabId,
          error: err && err.message ? err.message : "Chunk upload failed",
        })
        .catch(() => {});
    });
}

async function handleHybridRecordingComplete(tabId) {
  chunkDropLogged.delete(tabId);
  const state = hybridStates.get(tabId);
  if (!state || !state.uploadSession) return;

  const session = state.uploadSession;
  try {
    await session.pending;
    logBg(tabId, `complete HYBRID: ${session.chunkIndex} chunks, ${session.offset} bytes, failed=${session.failed}`);
    if (session.failed || session.completed) {
      state.uploadSession = null;
      return;
    }

    const completeResp = await fetch(`${session.apiUrl}/captures/upload/complete`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ upload_id: session.uploadID }),
    });

    if (!completeResp.ok) {
      const body = await completeResp.text();
      throw new Error(`Complete hybrid upload failed (${completeResp.status}): ${body}`);
    }

    logBg(tabId, "complete HYBRID OK (captura salva)");
    session.completed = true;
  } catch (err) {
    chrome.runtime
      .sendMessage({
        action: "hybrid_upload_error",
        tabId,
        error: err && err.message ? err.message : "Complete upload failed",
      })
      .catch(() => {});
  } finally {
    state.uploadSession = null;
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
    // USER_MEDIA: tabCapture recording. AUDIO_PLAYBACK: re-play the captured tab
    // audio back to the speakers so the owner can still hear while recording
    // (tabCapture otherwise swallows the tab's audio output).
    reasons: ["USER_MEDIA", "AUDIO_PLAYBACK"],
    justification: "Tab capture recording and audio passthrough",
  });

  await creatingOffscreen;
  creatingOffscreen = null;
}

// ---------------------------------------------------------------------------
// 6. Save Recording Blob → Upload to KuraNAS
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Tab Cleanup
// ---------------------------------------------------------------------------

chrome.tabs.onRemoved.addListener((tabId) => {
  detectedMedia.delete(tabId);
  detectedTitles.delete(tabId);
  detectedMetadata.delete(tabId);
  cleanupTab(tabId);
  chunkDropLogged.delete(tabId);
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo) => {
  if (changeInfo.url) {
    detectedMedia.delete(tabId);
    detectedTitles.delete(tabId);
    detectedMetadata.delete(tabId);
    updateBadge(tabId);
    // Ask content scripts to re-detect title + metadata for the new page
    chrome.tabs.sendMessage(tabId, { action: "request_title" }).catch(() => {});
    chrome.tabs.sendMessage(tabId, { action: "request_metadata" }).catch(() => {});
  }
});

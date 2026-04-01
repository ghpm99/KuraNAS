/* ========================================================================
 * KuraNAS Stream Grabber — Background Service Worker (MV3)
 * ======================================================================== */

import {
  DEFAULT_KURANAS_API_BASE,
  DISCOVERY_HEALTH_SUFFIX,
  DISCOVERY_REQUEST_TIMEOUT_MS,
  DISCOVERY_TAB_HOST_LIMIT,
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
import { createHybridStateMachine } from "./src/background/hybrid-state.js";

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

const detectedMedia = new Map();
const hybridStates = new Map();
const detectedTitles = new Map();

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
});
const {
  armHybrid,
  cleanupTab,
  disarmHybrid,
  getHybridStatus,
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
    getTitleForTab,
    handleBlobDetected,
    handleHybridRecordingChunk,
    handleHybridRecordingComplete,
    handleHybridVideoState,
    handleOffscreenError,
    handleOffscreenStarted,
    handleOffscreenStopped,
    handleSaveRecordingBlob,
    handleTitleDetected,
    stopHybridRecording,
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
// 4. Hybrid State Machine
// ---------------------------------------------------------------------------

function stopOffscreenRecording(tabId) {
  chrome.runtime
    .sendMessage({ action: "offscreen_stop_recording", tabId })
    .catch(() => {});
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
  const name = `recording_${tabId}_${Date.now()}`;
  const mimeType = "video/webm";
  const fileName = `${sanitizeFileName(name)}.webm`;

  const initResp = await fetch(`${apiUrl}/captures/upload/init`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      name,
      media_type: "recording",
      mime_type: mimeType,
      size: 0,
      file_name: fileName,
    }),
  });

  if (!initResp.ok) {
    const body = await initResp.text();
    throw new Error(`Init hybrid upload failed (${initResp.status}): ${body}`);
  }

  const payload = await initResp.json();
  if (!payload.upload_id) {
    throw new Error("Invalid hybrid upload init response: upload_id is required");
  }

  state.uploadSession = {
    apiUrl,
    uploadID: payload.upload_id,
    offset: 0,
    chunkIndex: 0,
    pending: Promise.resolve(),
    failed: false,
    completed: false,
  };

  return state.uploadSession;
}

function handleHybridRecordingChunk(tabId, msg) {
  const state = hybridStates.get(tabId);
  if (!state || !state.uploadSession || state.uploadSession.failed) return;

  const chunkBlob = msg.chunk;
  if (!chunkBlob || !chunkBlob.size) return;

  const session = state.uploadSession;
  session.pending = session.pending
    .then(async () => {
      if (session.failed || session.completed) return;
      await uploadChunkWithRetry(
        session.apiUrl,
        session.uploadID,
        chunkBlob,
        session.offset,
        session.chunkIndex
      );
      session.offset += chunkBlob.size;
      session.chunkIndex += 1;
    })
    .catch((err) => {
      session.failed = true;
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
  const state = hybridStates.get(tabId);
  if (!state || !state.uploadSession) return;

  const session = state.uploadSession;
  try {
    await session.pending;
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
    reasons: ["USER_MEDIA"],
    justification: "Tab capture recording for media grabbing",
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
  cleanupTab(tabId);
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

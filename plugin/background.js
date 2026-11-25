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
import {
  createConnectionStatusChecker,
  createReachabilityChecker,
  createServerAddressSaver,
  normalizeApiBaseUrl,
} from "./src/background/server-address.js";
import { createUploader } from "./src/background/uploader.js";
import { createHybridUploadSession } from "./src/background/hybrid-upload-session.js";
import { createDownloader } from "./src/background/downloader.js";
import { createFetcher } from "./src/background/fetcher.js";
import { createHybridStateMachine } from "./src/background/hybrid-state.js";

const detectedMedia = new Map();
const hybridStates = new Map();
const detectedTitles = new Map();
const detectedMetadata = new Map();

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
  resolveCaptureContext: (tabId) => resolveCaptureContext(tabId),
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
  resolveCaptureContext: (tabId) => resolveCaptureContext(tabId),
});
const {
  downloadDASH,
  downloadDirect,
  downloadHLS,
} = downloader;

const fetcher = createFetcher({ getApiBaseUrl });
const reachabilityOptions = {
  healthSuffix: DISCOVERY_HEALTH_SUFFIX,
  timeoutMs: DISCOVERY_REQUEST_TIMEOUT_MS,
};
const isKuraNASApiReachable = createReachabilityChecker(reachabilityOptions);
const checkConnectionStatus = createConnectionStatusChecker(reachabilityOptions);
const saveVerifiedServerAddress = createServerAddressSaver({
  isKuraNASApiReachable,
  saveApiBaseUrl: (apiBaseUrl) => chrome.storage.sync.set({ apiBaseUrl }),
});
const {
  submitFetch,
  listTargets: listIngestTargets,
  listPresets: listIngestPresets,
  getJobStatus: getIngestJobStatus,
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

const mediaDetectionManager = createMediaDetectionManager({
  chromeApi: chrome,
  detectedMedia,
  mediaPatterns: MEDIA_PATTERNS,
  mediaContentTypes: MEDIA_CONTENT_TYPES,
});
const addMedia = mediaDetectionManager.addMedia;
const updateBadge = mediaDetectionManager.updateBadge;
mediaDetectionManager.registerNetworkListeners();

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
    getIngestJobStatus,
    listIngestPresets,
    listIngestTargets,
    stopHybridRecording,
    submitFetch,
    uploadBlobCapture,
    saveVerifiedServerAddress,
    getConnectionStatus,
  }
));

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

function handleMetadataDetected(tabId, msg) {
  if (!tabId || !msg.metadata) return;
  detectedMetadata.set(tabId, { metadata: msg.metadata, timestamp: Date.now() });
}

function getMetadataForTab(tabId) {
  const entry = detectedMetadata.get(tabId);
  return entry ? entry.metadata : null;
}

function stopOffscreenRecording(tabId) {
  chrome.runtime
    .sendMessage({ action: "offscreen_stop_recording", tabId })
    .catch(() => {});
}

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

async function resolveCaptureMetadata(tabId) {
  chrome.tabs.sendMessage(tabId, { action: "request_metadata" }).catch(() => {});

  let latest = getMetadataForTab(tabId);
  for (let i = 0; i < 25; i++) {
    const meta = getMetadataForTab(tabId);
    if (meta) {
      latest = meta;
      if (meta.title) {
        logBg(tabId, `metadados (com título) resolvidos após ${(i + 1) * 100}ms`);
        return meta;
      }
    }
    await wait(100);
  }

  if (latest) {
    logBg(tabId, "metadados parciais (sem título estruturado) — usando o que há");
  }
  return latest || null;
}

async function resolveCaptureContext(tabId) {
  if (tabId === undefined || tabId === null) return {};
  const captureMetadata = await resolveCaptureMetadata(tabId);
  return { captureMetadata: captureMetadata || undefined, episodeKey: extractEpisodeKey(captureMetadata) };
}

function extractEpisodeKey(captureMetadata) {
  return (captureMetadata && captureMetadata.episode_key) || "";
}

async function probeOffscreenFormat() {
  try {
    await ensureOffscreen();
    const res = await chrome.runtime.sendMessage({ action: "offscreen_probe_format" });
    if (res && res.mimeType && res.fileExt) return res;
  } catch (_e) {
  }
  return { mimeType: "video/webm", fileExt: "webm", container: "webm" };
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
  const name = await resolveCaptureName(tabId);
  const fmt = await probeOffscreenFormat();
  const mimeType = fmt.mimeType;
  const fileName = `${sanitizeFileName(name)}.${fmt.fileExt}`;
  logBg(tabId, `nome da captura: "${name}" (${mimeType})`);

  const metadata = await resolveCaptureMetadata(tabId);
  const episodeKey = extractEpisodeKey(metadata);
  if (metadata) {
    logBg(tabId, `metadados: episode_key="${episodeKey}", plataforma="${metadata.platform || "?"}"`);
  }

  const buildInitBody = () => ({
    name,
    media_type: "recording",
    mime_type: mimeType,
    size: 0,
    file_name: fileName,
    episode_key: episodeKey,
    metadata: metadata || undefined,
  });

  const initResp = await fetch(`${apiUrl}/captures/upload/init`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(buildInitBody()),
  });

  if (!initResp.ok) {
    const body = await initResp.text();
    throw new Error(`Init hybrid upload failed (${initResp.status}): ${body}`);
  }

  const payload = await initResp.json();

  if (payload.already_complete) {
    logBg(tabId, `episódio já capturado (episode_key="${episodeKey}") — pulando gravação`);
    throw new Error("capture already complete for this episode");
  }

  if (!payload.upload_id) {
    throw new Error("Invalid hybrid upload init response: upload_id is required");
  }

  const startOffset = Number(payload.received_size || 0);

  state.uploadSession = createHybridUploadSession({
    apiUrl,
    uploadID: payload.upload_id,
    startOffset,
    buildInitBody,
    uploadChunkWithRetry,
    wait,
    onStatusChange: ({ status, error }) => {
      logBg(tabId, `upload HYBRID: ${status}${error ? ` (${error})` : ""}`);
      broadcastHybridStatus(tabId);
      if (status !== "failed") return;
      chrome.runtime
        .sendMessage({ action: "hybrid_upload_error", tabId, error })
        .catch(() => {});
    },
  });

  logBg(tabId, `upload session HYBRID criada: uploadID=${payload.upload_id}`);
  return state.uploadSession;
}

function logBg(tabId, ...args) {
  console.log("[KuraNAS bg]", `tab=${tabId}`, ...args);
}

const chunkDropLogged = new Set();
const chunkFetchChains = new Map();

async function fetchChunkBlob(chunkUrl) {
  const resp = await fetch(chunkUrl);
  const blob = await resp.blob();
  chrome.runtime
    .sendMessage({ action: "offscreen_revoke_url", url: chunkUrl })
    .catch(() => {});
  return blob;
}

function handleHybridRecordingChunk(tabId, msg) {
  const state = hybridStates.get(tabId);
  if (!state || !state.uploadSession || state.uploadSession.hasFailed()) {
    if (!chunkDropLogged.has(tabId)) {
      chunkDropLogged.add(tabId);
      logBg(
        tabId,
        "chunk DESCARTADO (hybrid): ",
        !state ? "sem state" : !state.uploadSession ? "sem uploadSession (init não rodou?)" : "upload falhou"
      );
    }
    return;
  }
  if (!msg.chunkUrl) return;

  const session = state.uploadSession;
  const previousFetch = chunkFetchChains.get(tabId) || Promise.resolve();
  const currentFetch = previousFetch
    .then(() => fetchChunkBlob(msg.chunkUrl))
    .then((chunkBlob) => {
      if (!chunkBlob.size) {
        logBg(tabId, "chunk vazio após fetch da blob URL (SW não leu a blob?)");
        return;
      }
      session.enqueueChunk(chunkBlob);
    })
    .catch((err) => {
      session.abort(err && err.message ? err.message : "Chunk read failed");
    });
  chunkFetchChains.set(tabId, currentFetch);
}

async function handleHybridRecordingComplete(tabId) {
  chunkDropLogged.delete(tabId);
  const state = hybridStates.get(tabId);
  if (!state || !state.uploadSession) return;

  const session = state.uploadSession;
  try {
    await (chunkFetchChains.get(tabId) || Promise.resolve());
    chunkFetchChains.delete(tabId);
    await session.waitUntilDrained();
    logBg(tabId, `complete HYBRID: ${session.getChunkIndex()} chunks, ${session.getUploadedOffset()} bytes, failed=${session.hasFailed()}`);
    if (session.hasFailed() || session.completed) {
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
    reasons: ["USER_MEDIA", "AUDIO_PLAYBACK"],
    justification: "Tab capture recording and audio passthrough",
  });

  await creatingOffscreen;
  creatingOffscreen = null;
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

async function getConnectionStatus() {
  const apiBaseUrl = await getApiBaseUrl();
  const status = await checkConnectionStatus(apiBaseUrl);
  return { status, apiBaseUrl };
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
    chrome.tabs.sendMessage(tabId, { action: "request_title" }).catch(() => {});
    chrome.tabs.sendMessage(tabId, { action: "request_metadata" }).catch(() => {});
  }
});

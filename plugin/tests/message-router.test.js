const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const messageRouterModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/background/message-router.js')
).href;

function createHandlers() {
  const calls = [];

  return {
    calls,
    handlers: {
      armHybrid: (tabId) => calls.push(['armHybrid', tabId]),
      disarmHybrid: (tabId) => calls.push(['disarmHybrid', tabId]),
      downloadDASH: async (url, name) => ({ kind: 'dash', url, name }),
      downloadDirect: async (url, name) => ({ kind: 'direct', url, name }),
      downloadHLS: async (url, name) => ({ kind: 'hls', url, name }),
      getDetectedMedia: (tabId) => [{ tabId, type: 'video' }],
      getIngestJobStatus: async (jobId) => ({ ok: true, jobId }),
      getHybridStatus: (tabId) => ({ tabId, state: 'ARMED' }),
      getTitleForTab: (tabId) => ({ title: `tab-${tabId}`, source: 'test' }),
      handleBlobDetected: (tabId, msg) => calls.push(['blob', tabId, msg.blobUrl]),
      handleHybridRecordingChunk: (tabId) => calls.push(['chunk', tabId]),
      handleHybridRecordingComplete: (tabId) => calls.push(['complete', tabId]),
      handleHybridVideoState: (tabId, snapshot) => calls.push(['video_state', tabId, snapshot]),
      handleMetadataDetected: (tabId, msg) => calls.push(['metadata', tabId, msg.metadata]),
      getMetadataForTab: (tabId) => ({ platform: 'test', tabId }),
      handleOffscreenError: (tabId) => calls.push(['offscreen_error', tabId]),
      handleOffscreenStarted: (tabId) => calls.push(['offscreen_started', tabId]),
      handleOffscreenStopped: (tabId) => calls.push(['offscreen_stopped', tabId]),
      handleSaveRecordingBlob: (msg) => calls.push(['save_blob', msg.name]),
      handleTitleDetected: (tabId, msg) => calls.push(['title', tabId, msg.title]),
      stopHybridRecording: (tabId) => calls.push(['stop', tabId]),
      uploadBlobCapture: async (tabId, blobUrl, name) => ({ tabId, blobUrl, name }),
    },
  };
}

test('routeRuntimeMessage dispatches sync actions', async () => {
  const { routeRuntimeMessage } = await import(messageRouterModuleUrl);
  const { calls, handlers } = createHandlers();
  const responses = [];
  const sendResponse = (payload) => responses.push(payload);

  const keepAlive = routeRuntimeMessage(
    { action: 'blob_detected', blobUrl: 'blob:test' },
    { tab: { id: 9 } },
    sendResponse,
    handlers
  );
  assert.equal(keepAlive, false);
  assert.deepEqual(calls[0], ['blob', 9, 'blob:test']);

  routeRuntimeMessage({ action: 'get_media', tabId: 9 }, {}, sendResponse, handlers);
  assert.deepEqual(responses[0], { media: [{ tabId: 9, type: 'video' }] });

  const metaAlive = routeRuntimeMessage(
    { action: 'metadata_detected', metadata: { platform: 'netflix' } },
    { tab: { id: 9 } },
    sendResponse,
    handlers
  );
  assert.equal(metaAlive, false);
  assert.deepEqual(calls[calls.length - 1], ['metadata', 9, { platform: 'netflix' }]);

  routeRuntimeMessage({ action: 'get_metadata', tabId: 7 }, {}, sendResponse, handlers);
  assert.deepEqual(responses[responses.length - 1], { metadata: { platform: 'test', tabId: 7 } });
});

test('routeRuntimeMessage keeps channel alive for async downloads', async () => {
  const { routeRuntimeMessage } = await import(messageRouterModuleUrl);
  const { handlers } = createHandlers();
  const responses = [];
  const sendResponse = (payload) => responses.push(payload);

  const keepAlive = routeRuntimeMessage(
    { action: 'download_hls', url: 'https://x/stream.m3u8', name: 'demo' },
    {},
    sendResponse,
    handlers
  );

  assert.equal(keepAlive, true);
  await new Promise((resolve) => setTimeout(resolve, 0));
  assert.deepEqual(responses[0], {
    kind: 'hls',
    url: 'https://x/stream.m3u8',
    name: 'demo',
  });
});

test('routeRuntimeMessage returns false for unknown action', async () => {
  const { routeRuntimeMessage } = await import(messageRouterModuleUrl);
  const { handlers } = createHandlers();
  const keepAlive = routeRuntimeMessage({ action: 'unknown' }, {}, () => {}, handlers);
  assert.equal(keepAlive, false);
});

test('routeRuntimeMessage routes ingest_job_status to the job status handler', async () => {
  const { routeRuntimeMessage } = await import(messageRouterModuleUrl);
  const { handlers } = createHandlers();
  const responses = [];

  const keepAlive = routeRuntimeMessage(
    { action: 'ingest_job_status', jobId: 12 },
    {},
    (payload) => responses.push(payload),
    handlers
  );
  await Promise.resolve();
  await Promise.resolve();

  assert.equal(keepAlive, true);
  assert.deepEqual(responses[0], { ok: true, jobId: 12 });
});

test('routeRuntimeMessage forwards tabId to download and blob upload handlers', async () => {
  const { routeRuntimeMessage } = await import(messageRouterModuleUrl);
  const received = [];
  const handlers = {
    downloadHLS: async (...args) => { received.push(['hls', ...args]); return {}; },
    downloadDASH: async (...args) => { received.push(['dash', ...args]); return {}; },
    downloadDirect: async (...args) => { received.push(['direct', ...args]); return {}; },
    uploadBlobCapture: async (...args) => { received.push(['blob', ...args]); return {}; },
  };

  routeRuntimeMessage({ action: 'download_hls', url: 'u1', name: 'n', tabId: 5 }, {}, () => {}, handlers);
  routeRuntimeMessage({ action: 'download_dash', url: 'u2', name: 'n', tabId: 5 }, {}, () => {}, handlers);
  routeRuntimeMessage({ action: 'download_direct', url: 'u3', name: 'n', tabId: 5 }, {}, () => {}, handlers);
  routeRuntimeMessage({ action: 'upload_blob_capture', blobUrl: 'b', name: 'n', tabId: 5 }, {}, () => {}, handlers);

  assert.deepEqual(received, [
    ['hls', 'u1', 'n', 5],
    ['dash', 'u2', 'n', 5],
    ['direct', 'u3', 'n', 5],
    ['blob', 5, 'b', 'n'],
  ]);
});

test('routeRuntimeMessage routes save_server_address to the address validation handler', async () => {
  const { routeRuntimeMessage } = await import(messageRouterModuleUrl);
  const { handlers } = createHandlers();
  const responses = [];
  handlers.saveVerifiedServerAddress = async (address) => ({ ok: true, apiBaseUrl: address });

  const keepAlive = routeRuntimeMessage(
    { action: 'save_server_address', address: 'http://nas:8000' },
    {},
    (payload) => responses.push(payload),
    handlers
  );
  await Promise.resolve();
  await Promise.resolve();

  assert.equal(keepAlive, true);
  assert.deepEqual(responses[0], { ok: true, apiBaseUrl: 'http://nas:8000' });
});

test('routeRuntimeMessage routes get_connection_status to the connection status handler', async () => {
  const { routeRuntimeMessage } = await import(messageRouterModuleUrl);
  const { handlers } = createHandlers();
  const responses = [];
  handlers.getConnectionStatus = async () => ({ status: 'forbidden', apiBaseUrl: 'http://nas:8000/api/v1' });

  const keepAlive = routeRuntimeMessage(
    { action: 'get_connection_status' },
    {},
    (payload) => responses.push(payload),
    handlers
  );
  await Promise.resolve();
  await Promise.resolve();

  assert.equal(keepAlive, true);
  assert.deepEqual(responses[0], { status: 'forbidden', apiBaseUrl: 'http://nas:8000/api/v1' });
});

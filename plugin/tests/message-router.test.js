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
      getHybridStatus: (tabId) => ({ tabId, state: 'ARMED' }),
      getTitleForTab: (tabId) => ({ title: `tab-${tabId}`, source: 'test' }),
      handleBlobDetected: (tabId, msg) => calls.push(['blob', tabId, msg.blobUrl]),
      handleHybridRecordingChunk: (tabId) => calls.push(['chunk', tabId]),
      handleHybridRecordingComplete: (tabId) => calls.push(['complete', tabId]),
      handleHybridVideoState: (tabId, snapshot) => calls.push(['video_state', tabId, snapshot]),
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

const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const mediaDetectionModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/background/media-detection.js')
).href;

test('classify helpers detect media by url and content-type', async () => {
  const { classifyByUrl, classifyByContentType } = await import(mediaDetectionModuleUrl);
  const mediaPatterns = [{ regex: /\.m3u8$/i, type: 'hls' }];
  const mediaContentTypes = [{ pattern: /^video\//i, type: 'video' }];

  assert.equal(classifyByUrl('https://example.com/stream.m3u8', mediaPatterns), 'hls');
  assert.equal(classifyByUrl('https://example.com/index.html', mediaPatterns), null);
  assert.equal(classifyByContentType('video/mp4', mediaContentTypes), 'video');
  assert.equal(classifyByContentType('text/plain', mediaContentTypes), null);
});

test('manager registers listeners and deduplicates detected media', async () => {
  const { createMediaDetectionManager } = await import(mediaDetectionModuleUrl);
  const detectedMedia = new Map();
  const sentMessages = [];
  const listeners = {};

  const chromeApi = {
    action: {
      setBadgeText: () => Promise.resolve(),
      setBadgeBackgroundColor: () => Promise.resolve(),
    },
    runtime: {
      sendMessage: (payload) => {
        sentMessages.push(payload);
        return Promise.resolve();
      },
    },
    webRequest: {
      onBeforeRequest: {
        addListener: (listener) => {
          listeners.beforeRequest = listener;
        },
      },
      onHeadersReceived: {
        addListener: (listener) => {
          listeners.headersReceived = listener;
        },
      },
    },
  };

  const manager = createMediaDetectionManager({
    chromeApi,
    detectedMedia,
    mediaPatterns: [{ regex: /\.m3u8$/i, type: 'hls' }],
    mediaContentTypes: [{ pattern: /^video\//i, type: 'video' }],
    now: () => 123,
  });

  manager.registerNetworkListeners();
  listeners.beforeRequest({ tabId: 7, url: 'https://example.com/live.m3u8' });
  listeners.beforeRequest({ tabId: 7, url: 'https://example.com/live.m3u8' });
  listeners.headersReceived({
    tabId: 7,
    url: 'https://example.com/clip',
    responseHeaders: [{ name: 'content-type', value: 'video/mp4' }],
  });

  const tabMedia = detectedMedia.get(7);
  assert.equal(tabMedia.length, 2);
  assert.equal(tabMedia[0].timestamp, 123);
  assert.equal(tabMedia[0].type, 'hls');
  assert.equal(tabMedia[1].type, 'video');
  assert.equal(sentMessages.length, 2);
});

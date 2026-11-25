const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const downloaderModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/background/downloader.js')
).href;

test('downloadDirect fetches blob and delegates upload', async () => {
  const { createDownloader } = await import(downloaderModuleUrl);
  const uploads = [];

  const downloader = createDownloader({
    resolveUrl: (base, relative) => new URL(relative, base).href,
    uploadToKuraNAS: async (blob, name, mediaType) => {
      uploads.push({ blob, name, mediaType });
      return { ok: true };
    },
    fetchImpl: async () => ({
      blob: async () => new Blob(['data'], { type: 'video/mp4' }),
    }),
  });

  const result = await downloader.downloadDirect('https://example.com/video.mp4', 'video_name');
  assert.equal(result.ok, true);
  assert.equal(result.name, 'video_name');
  assert.equal(uploads.length, 1);
  assert.equal(uploads[0].mediaType, 'direct');
});

test('downloadDirect resolves the tab capture context and passes metadata and episode key to the upload', async () => {
  const { createDownloader } = await import(downloaderModuleUrl);
  const uploads = [];
  const resolvedTabIds = [];

  const downloader = createDownloader({
    resolveUrl: (base, relative) => new URL(relative, base).href,
    resolveCaptureContext: async (tabId) => {
      resolvedTabIds.push(tabId);
      return { captureMetadata: { platform: 'crunchyroll' }, episodeKey: 'cr:ep-1' };
    },
    uploadToKuraNAS: async (blob, name, mediaType, captureMetadata, episodeKey) => {
      uploads.push({ name, mediaType, captureMetadata, episodeKey });
      return { ok: true };
    },
    fetchImpl: async () => ({
      blob: async () => new Blob(['data'], { type: 'video/mp4' }),
    }),
  });

  await downloader.downloadDirect('https://example.com/video.mp4', 'video_name', 42);
  assert.deepEqual(resolvedTabIds, [42]);
  assert.deepEqual(uploads, [
    { name: 'video_name', mediaType: 'direct', captureMetadata: { platform: 'crunchyroll' }, episodeKey: 'cr:ep-1' },
  ]);
});

test('downloadHLS media playlist passes the tab capture context to the upload', async () => {
  const { createDownloader } = await import(downloaderModuleUrl);
  const uploads = [];

  const downloader = createDownloader({
    resolveUrl: (base, relative) => new URL(relative, base).href,
    resolveCaptureContext: async () => ({ captureMetadata: { title: 'Ep' }, episodeKey: 'k1' }),
    uploadToKuraNAS: async (blob, name, mediaType, captureMetadata, episodeKey) => {
      uploads.push({ mediaType, captureMetadata, episodeKey });
    },
    fetchImpl: async (url) => ({
      text: async () => '#EXTM3U\nseg1.ts\n',
      arrayBuffer: async () => new Uint8Array([1, 2, 3]).buffer,
    }),
  });

  const result = await downloader.downloadHLS('https://example.com/p.m3u8', 'hls_name', 9);
  assert.equal(result.ok, true);
  assert.deepEqual(uploads, [{ mediaType: 'hls', captureMetadata: { title: 'Ep' }, episodeKey: 'k1' }]);
});

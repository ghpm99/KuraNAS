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

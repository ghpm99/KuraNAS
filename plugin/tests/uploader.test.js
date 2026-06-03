const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const uploaderModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/background/uploader.js')
).href;

test('uploadToKuraNAS performs init, chunk and complete flow', async () => {
  const { createUploader } = await import(uploaderModuleUrl);
  const calls = [];

  const fetchImpl = async (url, options = {}) => {
    calls.push({ url, method: options.method || 'GET' });

    if (url.endsWith('/captures/upload/init')) {
      return {
        ok: true,
        json: async () => ({ upload_id: 'upload-1', chunk_size: 2 }),
      };
    }

    if (url.endsWith('/captures/upload/chunk')) {
      return { ok: true, text: async () => '' };
    }

    if (url.endsWith('/captures/upload/complete')) {
      return { ok: true, json: async () => ({ ok: true }) };
    }

    throw new Error(`Unexpected URL ${url}`);
  };

  const uploader = createUploader({
    getApiBaseUrl: async () => 'http://api.local/api/v1',
    guessExtension: () => 'mp4',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    fetchImpl,
  });

  const blob = new Blob(['abcde'], { type: 'video/mp4' });
  const result = await uploader.uploadToKuraNAS(blob, 'video_name', 'direct');

  const chunkCalls = calls.filter((call) => call.url.endsWith('/captures/upload/chunk'));
  assert.equal(chunkCalls.length, 3);
  assert.deepEqual(result, { ok: true });
});

test('uploadBlobCapture uploads fetched blob', async () => {
  const { createUploader } = await import(uploaderModuleUrl);
  const calls = [];

  const fetchImpl = async (url, options = {}) => {
    calls.push({ url, method: options.method || 'GET' });

    if (url === 'blob:123') {
      return { blob: async () => new Blob(['abc'], { type: 'video/webm' }) };
    }
    if (url.endsWith('/captures/upload/init')) {
      return {
        ok: true,
        json: async () => ({ upload_id: 'upload-2', chunk_size: 10 }),
      };
    }
    if (url.endsWith('/captures/upload/chunk')) {
      return { ok: true, text: async () => '' };
    }
    if (url.endsWith('/captures/upload/complete')) {
      return { ok: true, json: async () => ({ ok: true }) };
    }
    throw new Error(`Unexpected URL ${url}`);
  };

  const uploader = createUploader({
    getApiBaseUrl: async () => 'http://api.local/api/v1',
    guessExtension: () => 'webm',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    fetchImpl,
  });

  const result = await uploader.uploadBlobCapture(1, 'blob:123', 'capture_name');
  assert.equal(result.ok, true);
  assert.equal(result.name, 'capture_name');
});

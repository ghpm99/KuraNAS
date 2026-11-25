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

test('uploadToKuraNAS rejects an empty blob before any network call', async () => {
  const { createUploader } = await import(uploaderModuleUrl);
  let fetchCalled = false;

  const uploader = createUploader({
    getApiBaseUrl: async () => 'http://api.local/api/v1',
    guessExtension: () => 'webm',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    fetchImpl: async () => {
      fetchCalled = true;
      return { ok: true, json: async () => ({}) };
    },
  });

  const emptyBlob = new Blob([], { type: 'video/webm' });
  await assert.rejects(
    () => uploader.uploadToKuraNAS(emptyBlob, 'recording_empty', 'recording'),
    /empty/i
  );
  assert.equal(fetchCalled, false);
});

test('uploadToKuraNAS realigns to received_size on chunk 409 without duplicating or skipping bytes', async () => {
  const { createUploader } = await import(uploaderModuleUrl);
  const chunkOffsets = [];
  let serverReceivedSize = 2;
  let hasRejectedFirstChunk = false;

  const fetchImpl = async (url, options = {}) => {
    if (url.endsWith('/captures/upload/init')) {
      return { ok: true, json: async () => ({ upload_id: 'upload-3', chunk_size: 2 }) };
    }

    if (url.endsWith('/captures/upload/chunk')) {
      const offset = Number(options.body.get('offset'));
      chunkOffsets.push(offset);
      if (offset !== serverReceivedSize) {
        hasRejectedFirstChunk = true;
        return {
          ok: false,
          status: 409,
          json: async () => ({ error: 'mismatch', received_size: serverReceivedSize }),
          text: async () => '',
        };
      }
      serverReceivedSize += options.body.get('chunk').size;
      return { ok: true, status: 200, text: async () => '' };
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

  await uploader.uploadToKuraNAS(new Blob(['abcdef'], { type: 'video/mp4' }), 'video_name', 'direct');

  assert.equal(hasRejectedFirstChunk, true);
  assert.deepEqual(chunkOffsets, [0, 2, 4]);
  assert.equal(serverReceivedSize, 6);
});

test('uploadToKuraNAS fails on 409 without a valid received_size', async () => {
  const { createUploader } = await import(uploaderModuleUrl);

  const fetchImpl = async (url) => {
    if (url.endsWith('/captures/upload/init')) {
      return { ok: true, json: async () => ({ upload_id: 'upload-4', chunk_size: 2 }) };
    }
    return { ok: false, status: 409, json: async () => ({}), text: async () => 'conflict' };
  };

  const uploader = createUploader({
    getApiBaseUrl: async () => 'http://api.local/api/v1',
    guessExtension: () => 'mp4',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    fetchImpl,
  });

  await assert.rejects(
    uploader.uploadToKuraNAS(new Blob(['abcdef']), 'video_name', 'direct'),
    /Chunk upload failed \(409\)/
  );
});

function createInitRecordingUploader() {
  const initBodies = [];
  const fetchImpl = async (url, options = {}) => {
    if (url.endsWith('/captures/upload/init')) {
      initBodies.push(JSON.parse(options.body));
      return { ok: true, json: async () => ({ upload_id: 'upload-1', chunk_size: 10 }) };
    }
    if (url.endsWith('/captures/upload/chunk')) {
      return { ok: true, text: async () => '' };
    }
    return { ok: true, json: async () => ({ ok: true }) };
  };
  return { initBodies, fetchImpl };
}

test('uploadToKuraNAS sends metadata and episode_key in the init body when provided', async () => {
  const { createUploader } = await import(uploaderModuleUrl);
  const { initBodies, fetchImpl } = createInitRecordingUploader();
  const uploader = createUploader({
    getApiBaseUrl: async () => 'http://api.local/api/v1',
    guessExtension: () => 'mp4',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    fetchImpl,
  });

  const captureMetadata = { title: 'Show', episode: 3 };
  await uploader.uploadToKuraNAS(
    new Blob(['abc'], { type: 'video/mp4' }),
    'video_name',
    'direct',
    captureMetadata,
    'show-s01e03'
  );

  assert.deepEqual(initBodies[0].metadata, captureMetadata);
  assert.equal(initBodies[0].episode_key, 'show-s01e03');
});

test('uploadToKuraNAS omits metadata and episode_key from the init body when absent', async () => {
  const { createUploader } = await import(uploaderModuleUrl);
  const { initBodies, fetchImpl } = createInitRecordingUploader();
  const uploader = createUploader({
    getApiBaseUrl: async () => 'http://api.local/api/v1',
    guessExtension: () => 'mp4',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    fetchImpl,
  });

  await uploader.uploadToKuraNAS(new Blob(['abc'], { type: 'video/mp4' }), 'video_name', 'direct');

  assert.equal('episode_key' in initBodies[0], false);
  assert.equal('metadata' in initBodies[0], false);
});

test('uploadBlobCapture sends the resolved metadata and episode key in the init body', async () => {
  const { createUploader } = await import(uploaderModuleUrl);
  const initBodies = [];

  const fetchImpl = async (url, options = {}) => {
    if (url === 'blob:abc') {
      return { blob: async () => new Blob(['abc'], { type: 'video/webm' }) };
    }
    if (url.endsWith('/captures/upload/init')) {
      initBodies.push(JSON.parse(options.body));
      return { ok: true, json: async () => ({ upload_id: 'u1' }) };
    }
    return { ok: true, json: async () => ({}), text: async () => '' };
  };

  const uploader = createUploader({
    getApiBaseUrl: async () => 'http://api.local/api/v1',
    guessExtension: () => 'webm',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    resolveCaptureContext: async (tabId) => ({ captureMetadata: { tabId }, episodeKey: 'ep-9' }),
    fetchImpl,
  });

  await uploader.uploadBlobCapture(3, 'blob:abc', 'cap');
  assert.deepEqual(initBodies[0].metadata, { tabId: 3 });
  assert.equal(initBodies[0].episode_key, 'ep-9');
});

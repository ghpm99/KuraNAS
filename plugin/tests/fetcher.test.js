const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const fetcherModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/background/fetcher.js')
).href;

const apiBase = 'http://nas.local/api/v1';
const getApiBaseUrl = async () => apiBase;

test('submitFetch posts the ingest payload and returns the job id', async () => {
  const { createFetcher } = await import(fetcherModuleUrl);
  const calls = [];

  const fetcher = createFetcher({
    getApiBaseUrl,
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      return { ok: true, status: 202, json: async () => ({ job_id: 99 }) };
    },
  });

  const result = await fetcher.submitFetch({
    url: 'https://youtu.be/abc',
    preset: 'audio_mp3',
    targetRoot: '/srv/midia',
    subfolder: 'musicas',
  });

  assert.equal(result.ok, true);
  assert.equal(result.jobId, 99);
  assert.equal(calls[0].url, `${apiBase}/ingest/fetch`);
  const body = JSON.parse(calls[0].options.body);
  assert.equal(body.url, 'https://youtu.be/abc');
  assert.equal(body.preset, 'audio_mp3');
  assert.equal(body.target_root, '/srv/midia');
  assert.equal(body.subfolder, 'musicas');
});

test('submitFetch surfaces a server error body', async () => {
  const { createFetcher } = await import(fetcherModuleUrl);
  const fetcher = createFetcher({
    getApiBaseUrl,
    fetchImpl: async () => ({
      ok: false,
      status: 400,
      json: async () => ({ error: 'URL inválida' }),
    }),
  });

  const result = await fetcher.submitFetch({ url: 'bad', preset: 'audio_mp3', targetRoot: '/srv' });
  assert.equal(result.ok, false);
  assert.equal(result.error, 'URL inválida');
});

test('submitFetch returns the thrown error message on network failure', async () => {
  const { createFetcher } = await import(fetcherModuleUrl);
  const fetcher = createFetcher({
    getApiBaseUrl,
    fetchImpl: async () => {
      throw new Error('offline');
    },
  });

  const result = await fetcher.submitFetch({ url: 'https://x/v', preset: 'audio_mp3', targetRoot: '/srv' });
  assert.equal(result.ok, false);
  assert.equal(result.error, 'offline');
});

test('listTargets and listPresets unwrap the JSON arrays', async () => {
  const { createFetcher } = await import(fetcherModuleUrl);
  const fetcher = createFetcher({
    getApiBaseUrl,
    fetchImpl: async (url) => {
      if (url.endsWith('/ingest/targets')) {
        return { ok: true, json: async () => [{ label: 'Midia', path: '/srv/midia' }] };
      }
      return { ok: true, json: async () => [{ key: 'audio_mp3', label: 'Áudio (MP3)' }] };
    },
  });

  const targets = await fetcher.listTargets();
  assert.equal(targets.ok, true);
  assert.equal(targets.targets[0].path, '/srv/midia');

  const presets = await fetcher.listPresets();
  assert.equal(presets.ok, true);
  assert.equal(presets.presets[0].key, 'audio_mp3');
});

test('listTargets reports failure without throwing', async () => {
  const { createFetcher } = await import(fetcherModuleUrl);
  const fetcher = createFetcher({
    getApiBaseUrl,
    fetchImpl: async () => ({ ok: false, status: 503, json: async () => ({}) }),
  });

  const result = await fetcher.listTargets();
  assert.equal(result.ok, false);
  assert.deepEqual(result.targets, []);
});

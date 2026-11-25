const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');
const { TextDecoder, TextEncoder } = require('node:util');

const pluginRoot = path.resolve(__dirname, '..');
const importModule = (relativePath) =>
  import(pathToFileURL(path.join(pluginRoot, relativePath)).href);

const forwardToServer = (server) => (...args) => server.fetchImpl(...args);

const API_URL = 'http://api.local/api/v1';
const UPLOAD_ID = 'upload-1';

function buildFakeServer() {
  const server = {
    receivedBytes: [],
    isOnline: true,
    initCalls: 0,
    chunkAttempts: 0,
    dropResponseOnce: false,
  };

  server.fetchImpl = async (url, options = {}) => {
    if (!server.isOnline) {
      throw new TypeError('Failed to fetch');
    }
    if (url.endsWith('/captures/upload/init')) {
      server.initCalls += 1;
      return {
        ok: true,
        status: 200,
        json: async () => ({
          upload_id: UPLOAD_ID,
          resumed: true,
          received_size: server.receivedBytes.length,
        }),
      };
    }
    if (url.endsWith('/captures/upload/chunk')) {
      server.chunkAttempts += 1;
      const offset = Number(options.body.get('offset'));
      if (offset !== server.receivedBytes.length) {
        return {
          ok: false,
          status: 409,
          json: async () => ({ received_size: server.receivedBytes.length }),
          text: async () => 'offset mismatch',
        };
      }
      const bytes = new Uint8Array(await options.body.get('chunk').arrayBuffer());
      server.receivedBytes.push(...bytes);
      if (server.dropResponseOnce) {
        server.dropResponseOnce = false;
        throw new TypeError('Failed to fetch');
      }
      return { ok: true, status: 200, text: async () => '' };
    }
    throw new Error(`Unexpected URL ${url}`);
  };

  return server;
}

function buildClock() {
  const clock = { currentMs: 0, waits: [] };
  clock.now = () => clock.currentMs;
  clock.wait = async (ms) => {
    clock.waits.push(ms);
    clock.currentMs += ms;
    if (clock.onWait) clock.onWait(clock.waits.length);
  };
  return clock;
}

async function buildSession({ server, clock, startOffset = 0, statuses = [], giveUpMs }) {
  const { createUploader } = await importModule('src/background/uploader.js');
  const { createHybridUploadSession } = await importModule(
    'src/background/hybrid-upload-session.js'
  );
  const uploader = createUploader({
    getApiBaseUrl: async () => API_URL,
    guessExtension: () => 'mp4',
    sanitizeFileName: (name) => name,
    waitFn: async () => {},
    fetchImpl: forwardToServer(server),
  });
  return createHybridUploadSession({
    apiUrl: API_URL,
    uploadID: UPLOAD_ID,
    startOffset,
    buildInitBody: () => ({ name: 'rec', episode_key: 'ep-1' }),
    uploadChunkWithRetry: uploader.uploadChunkWithRetry,
    fetchImpl: forwardToServer(server),
    wait: clock.wait,
    now: clock.now,
    onStatusChange: (change) => statuses.push(change),
    ...(giveUpMs ? { reconnectGiveUpMs: giveUpMs } : {}),
  });
}

const chunkOf = (text) => new Blob([text]);
const receivedText = (server) => new TextDecoder().decode(new Uint8Array(server.receivedBytes));

test('network drop then recovery delivers every byte in order without duplication', async () => {
  const server = buildFakeServer();
  const clock = buildClock();
  const statuses = [];
  const session = await buildSession({ server, clock, statuses });

  session.enqueueChunk(chunkOf('AAAA'));
  await session.waitUntilDrained();

  server.isOnline = false;
  clock.onWait = (waitCount) => {
    if (waitCount > 20) server.isOnline = true;
  };
  session.enqueueChunk(chunkOf('BBBB'));
  session.enqueueChunk(chunkOf('CCCC'));
  session.enqueueChunk(chunkOf('DDDD'));
  await session.waitUntilDrained();

  assert.equal(receivedText(server), 'AAAABBBBCCCCDDDD');
  assert.equal(session.hasFailed(), false);
  assert.equal(session.getUploadedOffset(), 16);
  assert.deepEqual(
    statuses.map((change) => change.status),
    ['reconnecting', 'uploading']
  );
  assert.deepEqual(clock.waits.slice(0, 5), [2000, 4000, 8000, 16000, 30000]);
});

test('chunk that reached the server but lost its response is not sent twice', async () => {
  const server = buildFakeServer();
  const clock = buildClock();
  const session = await buildSession({ server, clock });

  server.dropResponseOnce = true;
  session.enqueueChunk(chunkOf('AAAA'));
  session.enqueueChunk(chunkOf('BBBB'));
  await session.waitUntilDrained();

  assert.equal(receivedText(server), 'AAAABBBB');
  assert.equal(session.hasFailed(), false);
});

test('409 with received_size realigns the offset and drops the bytes the server already has', async () => {
  const server = buildFakeServer();
  const clock = buildClock();
  const session = await buildSession({ server, clock });

  server.receivedBytes.push(...new TextEncoder().encode('AA'));
  session.enqueueChunk(chunkOf('AAAA'));
  session.enqueueChunk(chunkOf('BBBB'));
  await session.waitUntilDrained();

  assert.equal(receivedText(server), 'AAAABBBB');
  assert.equal(session.hasFailed(), false);
  assert.equal(clock.waits.length, 0);
});

test('server holding less than what was released from memory fails clearly', async () => {
  const server = buildFakeServer();
  const clock = buildClock();
  const session = await buildSession({ server, clock });

  session.enqueueChunk(chunkOf('AAAA'));
  await session.waitUntilDrained();

  server.receivedBytes.length = 1;
  session.enqueueChunk(chunkOf('BBBB'));
  await session.waitUntilDrained();

  assert.equal(session.hasFailed(), true);
  assert.match(session.getStatus().error, /already released from memory/);
});

test('staying offline past the give-up time fails the session and reports it', async () => {
  const server = buildFakeServer();
  const clock = buildClock();
  const statuses = [];
  const session = await buildSession({ server, clock, statuses, giveUpMs: 60000 });

  server.isOnline = false;
  session.enqueueChunk(chunkOf('AAAA'));
  await session.waitUntilDrained();

  assert.equal(session.hasFailed(), true);
  assert.equal(session.getStatus().status, 'failed');
  assert.match(session.getStatus().error, /not restored in time/);
  assert.deepEqual(
    statuses.map((change) => change.status),
    ['reconnecting', 'failed']
  );
  assert.ok(clock.currentMs >= 60000);

  session.enqueueChunk(chunkOf('BBBB'));
  await session.waitUntilDrained();
  assert.equal(server.receivedBytes.length, 0);
});

test('server that lost the open session fails instead of silently restarting', async () => {
  const server = buildFakeServer();
  const clock = buildClock();
  const session = await buildSession({ server, clock });
  const originalFetch = server.fetchImpl;

  server.isOnline = false;
  clock.onWait = () => {
    server.isOnline = true;
    server.fetchImpl = async (url, options) =>
      url.endsWith('/captures/upload/init')
        ? { ok: true, status: 200, json: async () => ({ upload_id: 'other', resumed: false }) }
        : originalFetch(url, options);
  };
  session.enqueueChunk(chunkOf('AAAA'));
  await session.waitUntilDrained();

  assert.equal(session.hasFailed(), true);
  assert.match(session.getStatus().error, /no longer has the open upload session/);
});

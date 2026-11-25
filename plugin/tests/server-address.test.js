const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const serverAddressModuleUrl = pathToFileURL(
  path.resolve(__dirname, '..', 'src/background/server-address.js')
).href;

function createSaver(isReachable) {
  return import(serverAddressModuleUrl).then(({ createServerAddressSaver }) => {
    const savedAddresses = [];
    const reachabilityChecks = [];
    const saveVerifiedServerAddress = createServerAddressSaver({
      isKuraNASApiReachable: async (apiBaseUrl) => {
        reachabilityChecks.push(apiBaseUrl);
        return isReachable;
      },
      saveApiBaseUrl: async (apiBaseUrl) => savedAddresses.push(apiBaseUrl),
    });
    return { saveVerifiedServerAddress, savedAddresses, reachabilityChecks };
  });
}

test('reachable address is saved normalized to the api base', async () => {
  const { saveVerifiedServerAddress, savedAddresses, reachabilityChecks } = await createSaver(true);

  const outcome = await saveVerifiedServerAddress('  http://nas.local:8000/ ');

  assert.deepEqual(outcome, { ok: true, apiBaseUrl: 'http://nas.local:8000/api/v1' });
  assert.deepEqual(reachabilityChecks, ['http://nas.local:8000/api/v1']);
  assert.deepEqual(savedAddresses, ['http://nas.local:8000/api/v1']);
});

test('unreachable address is refused and not saved', async () => {
  const { saveVerifiedServerAddress, savedAddresses } = await createSaver(false);

  const outcome = await saveVerifiedServerAddress('http://nas.local:8000');

  assert.deepEqual(outcome, { ok: false, reason: 'unreachable' });
  assert.deepEqual(savedAddresses, []);
});

test('address that is not http(s) is refused without probing the network', async () => {
  const { saveVerifiedServerAddress, savedAddresses, reachabilityChecks } = await createSaver(true);

  const outcome = await saveVerifiedServerAddress('not a url');

  assert.deepEqual(outcome, { ok: false, reason: 'invalid_address' });
  assert.deepEqual(reachabilityChecks, []);
  assert.deepEqual(savedAddresses, []);
});

test('empty address clears the saved address to restore auto-detection', async () => {
  const { saveVerifiedServerAddress, savedAddresses } = await createSaver(false);

  const outcome = await saveVerifiedServerAddress('   ');

  assert.deepEqual(outcome, { ok: true, apiBaseUrl: '' });
  assert.deepEqual(savedAddresses, ['']);
});

test('reachability checker accepts only a healthy response that mentions KuraNAS', async () => {
  const { createReachabilityChecker } = await import(serverAddressModuleUrl);
  const responseFor = (ok, body) => async () => ({ ok, text: async () => body });
  const check = (fetchImpl) =>
    createReachabilityChecker({ healthSuffix: '/health', timeoutMs: 100, fetchImpl });

  assert.equal(await check(responseFor(true, '{"app":"KuraNAS"}'))('http://a/api/v1'), true);
  assert.equal(await check(responseFor(true, 'ok'))('http://a/api/v1'), false);
  assert.equal(await check(responseFor(false, 'kuranas'))('http://a/api/v1'), false);
  assert.equal(await check(async () => { throw new Error('down'); })('http://a/api/v1'), false);
});

const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const pluginRoot = path.resolve(__dirname, '..');

function readManifest() {
  const manifestPath = path.join(pluginRoot, 'manifest.json');
  const manifestRaw = fs.readFileSync(manifestPath, 'utf8');
  return JSON.parse(manifestRaw);
}

test('manifest points to existing runtime files', () => {
  const manifest = readManifest();

  const runtimeFiles = [
    manifest.background && manifest.background.service_worker,
    ...(manifest.content_scripts || []).flatMap((entry) => entry.js || []),
    manifest.action && manifest.action.default_popup
  ].filter(Boolean);

  runtimeFiles.forEach((relativeFilePath) => {
    const absolutePath = path.join(pluginRoot, relativeFilePath);
    assert.ok(
      fs.existsSync(absolutePath),
      `Expected manifest file to exist: ${relativeFilePath}`
    );
  });
});

test('manifest keeps mandatory MV3 identity fields', () => {
  const manifest = readManifest();

  assert.equal(manifest.manifest_version, 3);
  assert.equal(typeof manifest.name, 'string');
  assert.equal(typeof manifest.version, 'string');
  assert.ok(manifest.background && manifest.background.service_worker);
});

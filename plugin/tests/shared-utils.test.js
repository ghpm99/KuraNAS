const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const utilsModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/shared/utils.js')
).href;
const constantsModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/shared/constants.js')
).href;

test('sanitizeFileName normalizes reserved characters and spaces', async () => {
  const { sanitizeFileName } = await import(utilsModuleUrl);
  const input = 'my:invalid/file*name?.mp4';
  assert.equal(sanitizeFileName(input), 'my_invalid_file_name_.mp4');
});

test('guessExtension prefers mime type over media type fallback', async () => {
  const { guessExtension } = await import(utilsModuleUrl);
  assert.equal(guessExtension('video/mp4', 'hls'), 'mp4');
  assert.equal(guessExtension('', 'dash'), 'mp4');
  assert.equal(guessExtension('', ''), 'bin');
});

test('resolveUrl resolves relative paths against base URL', async () => {
  const { resolveUrl } = await import(utilsModuleUrl);
  assert.equal(
    resolveUrl('https://example.com/path/master.m3u8', 'segment.ts'),
    'https://example.com/path/segment.ts'
  );
});

test('constants expose media classification patterns', async () => {
  const { MEDIA_PATTERNS, MEDIA_CONTENT_TYPES, HYBRID_STABILITY_MS } = await import(constantsModuleUrl);
  assert.ok(Array.isArray(MEDIA_PATTERNS));
  assert.ok(Array.isArray(MEDIA_CONTENT_TYPES));
  assert.ok(MEDIA_PATTERNS.length > 0);
  assert.ok(MEDIA_CONTENT_TYPES.length > 0);
  assert.equal(HYBRID_STABILITY_MS, 200);
});

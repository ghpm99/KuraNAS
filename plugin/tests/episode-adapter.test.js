const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const registryUrl = pathToFileURL(
  path.join(pluginRoot, 'content/adapters/registry.js')
).href;
const crunchyrollUrl = pathToFileURL(
  path.join(pluginRoot, 'content/adapters/crunchyroll.js')
).href;

// registry.js / crunchyroll.js are side-effectful scripts; ES module caching
// means they run once per process, registering the adapter on globalThis. The
// registry guards against re-defining itself, so importing again is a no-op.
async function loadRegistry() {
  await import(registryUrl);
  await import(crunchyrollUrl);
  return globalThis.__kuraEpisodeAdapters;
}

// Minimal document fixture: maps the adapter's two selectors to text nodes.
function fakeDocument({ show, episode }) {
  return {
    querySelector(selector) {
      if (selector.includes('h1.hero-heading-line')) {
        return show ? { textContent: show } : null;
      }
      if (selector === '[class*="CurrentMediaInfo"] h1') {
        return episode ? { textContent: episode } : null;
      }
      return null;
    },
  };
}

test('registry resolves the crunchyroll adapter by host and ignores others', async () => {
  const registry = await loadRegistry();

  assert.equal(registry.resolve('www.crunchyroll.com').service, 'crunchyroll');
  assert.equal(registry.resolve('crunchyroll.com').service, 'crunchyroll');
  assert.equal(registry.resolve('beta.crunchyroll.com').service, 'crunchyroll');
  assert.equal(registry.resolve('www.netflix.com'), null);
});

test('crunchyroll adapter extracts a stable episodeId from the watch URL', async () => {
  const registry = await loadRegistry();
  const adapter = registry.resolve('www.crunchyroll.com');

  const ctx = {
    location: { href: 'https://www.crunchyroll.com/watch/GREP01ABC/the-first-episode' },
    document: fakeDocument({ show: 'My Show', episode: 'Episode 1' }),
  };

  const { episodeId, title } = adapter.getEpisode(ctx);
  assert.equal(episodeId, 'GREP01ABC');
  assert.equal(title, 'My Show - Episode 1');
});

test('crunchyroll adapter falls back to a single title and null id', async () => {
  const registry = await loadRegistry();
  const adapter = registry.resolve('www.crunchyroll.com');

  const ctx = {
    location: { href: 'https://www.crunchyroll.com/series/GSHOW/my-show' },
    document: fakeDocument({ show: 'My Show', episode: null }),
  };

  const { episodeId, title } = adapter.getEpisode(ctx);
  assert.equal(episodeId, null); // not a /watch/ URL
  assert.equal(title, 'My Show');
});

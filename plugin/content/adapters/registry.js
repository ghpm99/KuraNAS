/* ========================================================================
 * KuraNAS Stream Grabber — Episode Adapter Registry (ISOLATED world)
 *
 * Maps hostname -> page adapter for the smart per-episode capture. Each adapter
 * is the only site-specific piece (it breaks when the site changes), so it stays
 * small and isolated; this registry just resolves one for the current host. A
 * host with no adapter resolves to null and arms nothing (the manual hybrid mode
 * still applies).
 *
 * Written as a side-effectful script (no import/export) so it loads both as a
 * content script and, for tests, via a plain `import()` of the file.
 * ======================================================================== */

(function () {
  "use strict";

  if (globalThis.__kuraEpisodeAdapters) return;

  const adapters = [];

  globalThis.__kuraEpisodeAdapters = {
    register(adapter) {
      if (adapter && typeof adapter.matches === "function") {
        adapters.push(adapter);
      }
    },
    resolve(hostname) {
      for (const adapter of adapters) {
        try {
          if (adapter.matches(hostname)) return adapter;
        } catch {
          // a broken adapter must not poison resolution for the others
        }
      }
      return null;
    },
    list() {
      return adapters.slice();
    },
  };
})();

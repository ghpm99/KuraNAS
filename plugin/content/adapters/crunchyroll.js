/* ========================================================================
 * KuraNAS Stream Grabber — Crunchyroll Episode Adapter (ISOLATED world)
 *
 * The cobaia (pilot) adapter for the smart per-episode capture. It normalizes
 * the bit that is site-specific: a stable episodeId (the content id in the
 * /watch/<id> URL) and a readable title from the player DOM. Everything else of
 * the player state (isPlaying/currentTime/duration/fullscreen) is read
 * generically from the <video> by the monitor — keep this file tiny so only it
 * needs maintenance when Crunchyroll changes its markup.
 *
 * Side-effectful registration (no import/export): loads as a content script and,
 * for tests, via a plain `import()` after registry.js has defined the registry.
 * ======================================================================== */

(function () {
  "use strict";

  const registry = globalThis.__kuraEpisodeAdapters;
  if (!registry) return;

  function matches(hostname) {
    return (
      typeof hostname === "string" &&
      (hostname === "crunchyroll.com" || hostname.endsWith(".crunchyroll.com"))
    );
  }

  // Crunchyroll watch URLs look like /watch/GREP01ABC/episode-slug — the content
  // id right after /watch/ is stable per episode and is the capture key.
  function getEpisodeId(ctx) {
    const href = ctx && ctx.location && ctx.location.href;
    if (!href) return null;
    const match = /\/watch\/([A-Za-z0-9]+)/.exec(href);
    return match ? match[1] : null;
  }

  function getTitle(ctx) {
    const doc = ctx && ctx.document;
    if (!doc) return null;

    const show = doc.querySelector(
      'h1.hero-heading-line, [class*="CurrentMediaInfo"] h4, .erc-current-media-info h4'
    );
    const episode = doc.querySelector('[class*="CurrentMediaInfo"] h1');

    const showText = show && show.textContent ? show.textContent.trim() : "";
    const episodeText =
      episode && episode.textContent ? episode.textContent.trim() : "";

    if (showText && episodeText) return `${showText} - ${episodeText}`;
    return showText || episodeText || null;
  }

  registry.register({
    service: "crunchyroll",
    matches,
    getEpisode(ctx) {
      return { episodeId: getEpisodeId(ctx), title: getTitle(ctx) };
    },
  });
})();

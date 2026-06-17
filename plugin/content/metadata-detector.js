/* ========================================================================
 * KuraNAS Stream Grabber — Metadata Detector (MAIN world)
 *
 * Builds a STANDARDIZED metadata object for the media currently playing, so the
 * server can persist it (metadata.json) beside the recording. The canonical keys
 * below are platform-agnostic; each streaming site does its own "de→para",
 * mapping its internal/DOM/structured data onto these keys. A generic layer
 * (schema.org JSON-LD + OpenGraph + the <video> element) fills whatever a site
 * strategy does not, so unknown sites still yield useful metadata.
 *
 * Canonical keys (all optional — only what we could resolve is sent):
 *   platform, source_url, content_type ("movie"|"episode"|"series"|"clip"|"live"),
 *   title, episode_title, season, episode, episode_key,
 *   content_id, series_id,
 *   duration_seconds, description, release_year,
 *   genres[], content_rating, cast[], directors[], studio,
 *   audio_language, subtitle_language,
 *   thumbnail_url, poster_url,
 *   next_episode_title, next_episode_url,
 *   captured_at
 * ======================================================================== */

(function () {
  "use strict";

  // -----------------------------------------------------------------------
  // Small DOM / parsing helpers
  // -----------------------------------------------------------------------

  function text(el) {
    return el && el.textContent ? el.textContent.trim() : null;
  }

  function firstText(selectors) {
    for (const sel of selectors) {
      const el = document.querySelector(sel);
      const t = text(el);
      if (t) return t;
    }
    return null;
  }

  function metaContent(selectors) {
    for (const sel of selectors) {
      const el = document.querySelector(sel);
      if (el && el.content) return el.content;
    }
    return null;
  }

  // Parse an ISO 8601 duration ("PT1H23M9S") into seconds. Returns null on miss.
  function parseIsoDuration(value) {
    if (!value || typeof value !== "string") return null;
    const m = value.match(/^P(?:\d+D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?$/);
    if (!m) return null;
    const h = Number(m[1] || 0);
    const min = Number(m[2] || 0);
    const s = Number(m[3] || 0);
    const total = h * 3600 + min * 60 + s;
    return total > 0 ? Math.round(total) : null;
  }

  function getLargestVideo() {
    let largest = null;
    let largestArea = 0;
    for (const v of document.querySelectorAll("video")) {
      const rect = v.getBoundingClientRect();
      const area = rect.width * rect.height;
      if (area > largestArea) {
        largestArea = area;
        largest = v;
      }
    }
    return largest;
  }

  function names(value) {
    if (!value) return null;
    const arr = Array.isArray(value) ? value : [value];
    const out = [];
    for (const item of arr) {
      const name = typeof item === "string" ? item : item && item.name;
      if (name) out.push(String(name).trim());
    }
    return out.length ? out : null;
  }

  function yearOf(value) {
    if (!value) return null;
    const match = String(value).match(/(\d{4})/);
    return match ? Number(match[1]) : null;
  }

  // Merge `src` onto `dst` without letting empty values clobber filled ones.
  function mergeDefined(dst, src) {
    if (!src) return dst;
    for (const key of Object.keys(src)) {
      const v = src[key];
      if (v === null || v === undefined || v === "") continue;
      if (Array.isArray(v) && v.length === 0) continue;
      dst[key] = v;
    }
    return dst;
  }

  // -----------------------------------------------------------------------
  // Generic schema.org JSON-LD layer (covers Crunchyroll, Vimeo, many sites)
  // -----------------------------------------------------------------------

  const VIDEO_TYPES = new Set(["TVEpisode", "Movie", "VideoObject", "TVSeries", "Episode"]);

  function collectJsonLdNodes() {
    const nodes = [];
    for (const script of document.querySelectorAll('script[type="application/ld+json"]')) {
      let data;
      try {
        data = JSON.parse(script.textContent);
      } catch {
        continue;
      }
      const items = Array.isArray(data)
        ? data
        : Array.isArray(data["@graph"])
          ? data["@graph"]
          : [data];
      for (const item of items) {
        if (item && typeof item === "object") nodes.push(item);
      }
    }
    return nodes;
  }

  function fromJsonLd() {
    const nodes = collectJsonLdNodes();
    const node = nodes.find((n) => VIDEO_TYPES.has(n["@type"]));
    if (!node) return null;

    const out = {};
    const isEpisode = node["@type"] === "TVEpisode" || node["@type"] === "Episode";
    const series = node.partOfSeries && node.partOfSeries.name;

    out.content_type = isEpisode ? "episode" : node["@type"] === "Movie" ? "movie" : "clip";
    out.title = series || node.name || null;
    if (isEpisode) {
      out.episode_title = node.name || null;
      if (node.episodeNumber != null) out.episode = Number(node.episodeNumber);
      const season = node.partOfSeason && node.partOfSeason.seasonNumber;
      if (season != null) out.season = Number(season);
    }
    out.description = node.description || null;
    out.duration_seconds = parseIsoDuration(node.duration) || null;
    out.release_year = yearOf(node.datePublished || node.dateCreated);
    out.genres = names(node.genre);
    out.content_rating = node.contentRating || null;
    out.cast = names(node.actor || node.actors);
    out.directors = names(node.director);
    out.studio = (names(node.productionCompany || node.creator || node.author) || [])[0] || null;
    out.thumbnail_url =
      (typeof node.thumbnailUrl === "string" ? node.thumbnailUrl : null) ||
      (node.image && (typeof node.image === "string" ? node.image : node.image.url)) ||
      null;
    out.content_id = node.url || node["@id"] || null;
    return out;
  }

  // -----------------------------------------------------------------------
  // Generic OpenGraph / <video> / document fallbacks
  // -----------------------------------------------------------------------

  function fromOpenGraph() {
    return {
      title: metaContent(['meta[property="og:title"]', 'meta[name="twitter:title"]']),
      description: metaContent([
        'meta[property="og:description"]',
        'meta[name="description"]',
      ]),
      thumbnail_url: metaContent([
        'meta[property="og:image"]',
        'meta[name="twitter:image"]',
      ]),
      content_type: metaContent(['meta[property="og:type"]']) === "video.movie" ? "movie" : null,
    };
  }

  function fromVideoElement() {
    const video = getLargestVideo();
    if (!video) return null;
    const out = {};
    if (Number.isFinite(video.duration) && video.duration > 0) {
      out.duration_seconds = Math.round(video.duration);
    }
    return out;
  }

  // -----------------------------------------------------------------------
  // Per-site de→para strategies (override the generic layer)
  // -----------------------------------------------------------------------

  const siteStrategies = {
    // Netflix renders the title as stacked spans: show title, "E<n>" badge and the
    // episode title. Structured data is absent, so read the player overlay.
    "netflix.com": function () {
      const out = {};
      const title = firstText([
        '[data-uia="video-title"] h4',
        '[data-uia="video-title"]',
        ".video-title h4",
      ]);
      if (title) out.title = title;
      const spans = document.querySelectorAll('[data-uia="video-title"] span');
      if (spans.length >= 2) {
        const epBadge = text(spans[0]); // e.g. "E1"
        const epNum = epBadge && epBadge.match(/E(\d+)/i);
        if (epNum) out.episode = Number(epNum[1]);
        const epTitle = text(spans[spans.length - 1]);
        if (epTitle && epTitle !== epBadge) out.episode_title = epTitle;
        out.content_type = "episode";
      }
      return out;
    },

    "primevideo.com": function () {
      return {
        title: firstText([
          ".atvwebplayersdk-title-text",
          '[data-automation-id="title"]',
          ".dv-node-dp-title",
        ]),
        episode_title: firstText([".atvwebplayersdk-subtitle-text"]),
      };
    },

    "crunchyroll.com": function () {
      // JSON-LD already carries the show/season/episode reliably; only force the
      // content type so unknown JSON-LD shapes still classify as an episode.
      return { content_type: "episode" };
    },

    "disneyplus.com": function () {
      return {
        title: firstText(['[data-testid="title-field"]', ".title-field"]),
      };
    },

    "max.com": function () {
      return {
        title: firstText(['[class*="StyledPlayerMetaTitle"]', '[data-testid="player-title"]']),
      };
    },
  };

  function matchSiteStrategy(hostname) {
    for (const domain of Object.keys(siteStrategies)) {
      if (hostname === domain || hostname.endsWith("." + domain)) {
        return siteStrategies[domain];
      }
    }
    return null;
  }

  const PLATFORM_NAMES = {
    "netflix.com": "netflix",
    "primevideo.com": "prime_video",
    "amazon.com": "prime_video",
    "crunchyroll.com": "crunchyroll",
    "disneyplus.com": "disney_plus",
    "max.com": "max",
    "hbomax.com": "max",
    "youtube.com": "youtube",
    "twitch.tv": "twitch",
    "vimeo.com": "vimeo",
    "dailymotion.com": "dailymotion",
    "globoplay.globo.com": "globoplay",
    "mercadolivre.com.br": "mercado_play",
    "pluto.tv": "pluto_tv",
  };

  function platformFor(hostname) {
    for (const domain of Object.keys(PLATFORM_NAMES)) {
      if (hostname === domain || hostname.endsWith("." + domain)) {
        return PLATFORM_NAMES[domain];
      }
    }
    // Fall back to the registrable-ish host (drop a leading "www.").
    return hostname.replace(/^www\./, "");
  }

  // -----------------------------------------------------------------------
  // Build + emit
  // -----------------------------------------------------------------------

  function buildEpisodeKey(meta) {
    if (meta.content_id) return `${meta.platform}:${meta.content_id}`;
    if (meta.title && meta.season != null && meta.episode != null) {
      return `${meta.platform}:${meta.title}:S${meta.season}E${meta.episode}`;
    }
    return null;
  }

  function buildMetadata() {
    const hostname = location.hostname;
    const meta = {};

    // Generic layers first (lowest precedence), then the site strategy on top.
    mergeDefined(meta, fromOpenGraph());
    mergeDefined(meta, fromVideoElement());
    mergeDefined(meta, fromJsonLd());

    const strategy = matchSiteStrategy(hostname);
    if (strategy) {
      try {
        mergeDefined(meta, strategy());
      } catch {
        // a broken strategy never blocks the generic metadata
      }
    }

    meta.platform = platformFor(hostname);
    meta.source_url = location.href;
    meta.captured_at = new Date().toISOString();
    if (!meta.content_type) meta.content_type = meta.episode != null ? "episode" : "clip";

    const key = buildEpisodeKey(meta);
    if (key) meta.episode_key = key;

    return meta;
  }

  function hasUsefulMetadata(meta) {
    return Boolean(meta.title || meta.episode_title || meta.description || meta.duration_seconds);
  }

  let lastSerialized = null;

  function emitMetadata() {
    const meta = buildMetadata();
    if (!hasUsefulMetadata(meta)) return;

    const serialized = JSON.stringify(meta);
    if (serialized === lastSerialized) return;
    lastSerialized = serialized;

    window.dispatchEvent(
      new CustomEvent("__stream_grabber_metadata__", { detail: { metadata: meta } })
    );
  }

  function scheduleEmit(delay) {
    setTimeout(emitMetadata, delay);
  }

  // Explicit re-detect request (e.g. just before recording starts).
  window.addEventListener("__stream_grabber_request_metadata__", () => {
    lastSerialized = null;
    emitMetadata();
  });

  // Initial detection plus a few delayed re-runs: sites such as Crunchyroll
  // inject the rich TVEpisode JSON-LD a beat AFTER the first paint, so an early
  // single pass only sees the generic OpenGraph snapshot.
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => scheduleEmit(600));
  } else {
    scheduleEmit(400);
  }
  scheduleEmit(1500);
  scheduleEmit(3000);

  // Re-detect on SPA navigation (URL change).
  let lastUrl = location.href;
  const urlObserver = new MutationObserver(() => {
    if (location.href !== lastUrl) {
      lastUrl = location.href;
      lastSerialized = null;
      scheduleEmit(900);
      scheduleEmit(2000);
    }
  });
  urlObserver.observe(document.documentElement, { childList: true, subtree: true });

  // Re-detect when structured data (JSON-LD) or title/meta tags are added or
  // changed on the SAME url. This is the fix for the first episode keeping only
  // the early generic snapshot: the episode's TVEpisode JSON-LD lands after load,
  // and dedup (lastSerialized) makes these extra passes cheap no-ops otherwise.
  const head = document.head || document.documentElement;
  const dataObserver = new MutationObserver((mutations) => {
    for (const m of mutations) {
      const tag = m.target && m.target.tagName;
      if (tag === "SCRIPT" || tag === "META" || tag === "TITLE") {
        scheduleEmit(200);
        return;
      }
      for (const node of m.addedNodes || []) {
        const ntag = node && node.tagName;
        if (ntag === "SCRIPT" || ntag === "META") {
          scheduleEmit(200);
          return;
        }
      }
    }
  });
  dataObserver.observe(head, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ["content"],
  });
})();

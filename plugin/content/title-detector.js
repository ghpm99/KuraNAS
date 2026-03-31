/* ========================================================================
 * KuraNAS Stream Grabber — Title Detector (MAIN world)
 *
 * Detects the current media title by inspecting domain-specific metadata,
 * structured data, and page elements. Dispatches results to the bridge
 * via CustomEvent.
 * ======================================================================== */

(function () {
  "use strict";

  // -----------------------------------------------------------------------
  // Domain registry — each extractor returns a title string or null
  // -----------------------------------------------------------------------

  const domainExtractors = {
    // ----- YouTube -----
    "youtube.com": function () {
      // ytInitialPlayerResponse is the richest source
      if (window.ytInitialPlayerResponse) {
        const vd = window.ytInitialPlayerResponse.videoDetails;
        if (vd && vd.title) return vd.title;
      }

      // ytInitialData for playlist / channel page titles
      if (window.ytInitialData) {
        try {
          const contents =
            window.ytInitialData.contents?.twoColumnWatchNextResults?.results
              ?.results?.contents;
          if (contents) {
            for (const c of contents) {
              const title =
                c.videoPrimaryInfoRenderer?.title?.runs?.[0]?.text;
              if (title) return title;
            }
          }
        } catch {
          // structure changed, fall through
        }
      }

      // DOM fallback — meta tag
      const metaTitle = document.querySelector(
        'meta[name="title"], meta[property="og:title"]'
      );
      if (metaTitle && metaTitle.content) return metaTitle.content;

      // DOM fallback — heading
      const h1 = document.querySelector(
        "h1.ytd-watch-metadata yt-formatted-string, #title h1 yt-formatted-string"
      );
      if (h1 && h1.textContent.trim()) return h1.textContent.trim();

      return null;
    },

    // ----- Netflix -----
    "netflix.com": function () {
      // Netflix player API exposed on the cadmium player
      const videoPlayer = window.netflix?.appContext
        ?.getPlayerApp?.()
        ?.getAPI?.();
      if (videoPlayer) {
        try {
          const sessions = videoPlayer.videoPlayer.getAllPlayerSessionIds();
          if (sessions && sessions.length > 0) {
            const player = videoPlayer.videoPlayer.getVideoPlayerBySessionId(
              sessions[0]
            );
            if (player) {
              const titleId = player.getMovieId?.();
              // Try to get from metadata cache
              const metadataCache =
                window.netflix?.reactContext?.models?.userInfo?.data
                  ?.memberContext;
              if (metadataCache) {
                // Sometimes available
              }
            }
          }
        } catch {
          // API changed
        }
      }

      // Netflix renders the title in various places
      const nfTitle = document.querySelector(
        '[data-uia="video-title"], .video-title h4, .video-title'
      );
      if (nfTitle && nfTitle.textContent.trim())
        return nfTitle.textContent.trim();

      // Preload link with title
      const preloadLink = document.querySelector('link[rel="preload"][as="fetch"][href*="/metadata"]');
      // Not always useful, but og:title usually works
      return extractOpenGraphTitle();
    },

    // ----- Disney+ -----
    "disneyplus.com": function () {
      const titleEl = document.querySelector(
        '[data-testid="title-field"], .title-field, h2[class*="title"]'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    // ----- Amazon Prime Video -----
    "primevideo.com": function () {
      // Prime exposes catalog data in a script tag
      const scripts = document.querySelectorAll(
        'script[type="text/template"]'
      );
      for (const script of scripts) {
        try {
          const data = JSON.parse(script.textContent);
          if (data.title) return data.title;
          if (data.catalogMetadata?.catalog?.title)
            return data.catalogMetadata.catalog.title;
        } catch {
          // not JSON or wrong structure
        }
      }

      const titleEl = document.querySelector(
        '[data-automation-id="title"], .atvwebplayersdk-title-text, .dv-node-dp-title'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    "amazon.com": function () {
      return domainExtractors["primevideo.com"]();
    },

    // ----- HBO Max / Max -----
    "max.com": function () {
      const titleEl = document.querySelector(
        '[class*="StyledPlayerMetaTitle"], [data-testid="player-title"]'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    "hbomax.com": function () {
      return domainExtractors["max.com"]();
    },

    // ----- Twitch -----
    "twitch.tv": function () {
      // Twitch has rich structured data
      const ldJson = extractLdJsonTitle();
      if (ldJson) return ldJson;

      const titleEl = document.querySelector(
        '[data-a-target="stream-title"], h2[data-a-target="stream-title"]'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      // Channel name as fallback
      const channel = document.querySelector(
        'h1[data-a-target="stream-channel-link"], a[data-a-target="user-channel-header-item"] h1'
      );
      if (channel && channel.textContent.trim())
        return channel.textContent.trim();

      return extractOpenGraphTitle();
    },

    // ----- Crunchyroll -----
    "crunchyroll.com": function () {
      const titleEl = document.querySelector(
        'h1.hero-heading-line, [class*="CurrentMediaInfo"] h4, .erc-current-media-info h4'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      const episodeTitle = document.querySelector(
        '[class*="CurrentMediaInfo"] h1'
      );
      if (episodeTitle && episodeTitle.textContent.trim()) {
        const show = titleEl ? titleEl.textContent.trim() + " - " : "";
        return show + episodeTitle.textContent.trim();
      }

      return extractOpenGraphTitle();
    },

    // ----- Vimeo -----
    "vimeo.com": function () {
      if (window.vimeo?.clip_page_config?.clip?.title) {
        return window.vimeo.clip_page_config.clip.title;
      }

      const ldJson = extractLdJsonTitle();
      if (ldJson) return ldJson;

      return extractOpenGraphTitle();
    },

    // ----- Dailymotion -----
    "dailymotion.com": function () {
      const titleEl = document.querySelector(
        '[class*="VideoInfoTitle"], .VideoInfoTitle'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    // ----- Globoplay -----
    "globoplay.globo.com": function () {
      const titleEl = document.querySelector(
        '[class*="playback-title"], .playback-title'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    // ----- Pluto TV -----
    "pluto.tv": function () {
      const titleEl = document.querySelector(
        '[class*="MetadataTitle"], .player-metadata-title'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },
  };

  // -----------------------------------------------------------------------
  // Generic extractors (used as fallback chain)
  // -----------------------------------------------------------------------

  function extractOpenGraphTitle() {
    const og = document.querySelector('meta[property="og:title"]');
    if (og && og.content) return og.content;

    const twitter = document.querySelector('meta[name="twitter:title"]');
    if (twitter && twitter.content) return twitter.content;

    return null;
  }

  function extractLdJsonTitle() {
    const scripts = document.querySelectorAll(
      'script[type="application/ld+json"]'
    );
    for (const script of scripts) {
      try {
        const data = JSON.parse(script.textContent);
        const items = Array.isArray(data) ? data : [data];
        for (const item of items) {
          if (
            item["@type"] === "VideoObject" ||
            item["@type"] === "Movie" ||
            item["@type"] === "TVEpisode" ||
            item["@type"] === "TVSeries"
          ) {
            if (item.name) return item.name;
          }
          // Check nested items (e.g., @graph)
          if (item["@graph"]) {
            for (const node of item["@graph"]) {
              if (
                (node["@type"] === "VideoObject" ||
                  node["@type"] === "Movie" ||
                  node["@type"] === "TVEpisode") &&
                node.name
              ) {
                return node.name;
              }
            }
          }
        }
      } catch {
        // invalid JSON
      }
    }
    return null;
  }

  function extractFromVideoElement() {
    const videos = document.querySelectorAll("video");
    for (const video of videos) {
      if (video.title) return video.title;
      const ariaLabel = video.getAttribute("aria-label");
      if (ariaLabel) return ariaLabel;
    }
    return null;
  }

  function extractFromDocumentTitle() {
    const title = document.title;
    if (!title) return null;

    // Clean common suffixes like " - YouTube", " | Netflix", " — Twitch"
    return title
      .replace(/\s*[-|—]\s*(YouTube|Netflix|Twitch|Disney\+|Prime Video|HBO Max|Max|Crunchyroll|Vimeo|Dailymotion|Globoplay).*$/i, "")
      .trim() || null;
  }

  function extractFromStreamUrl(url) {
    if (!url) return null;
    try {
      const pathname = new URL(url).pathname;
      const segments = pathname.split("/").filter(Boolean);
      const last = segments[segments.length - 1];
      if (!last) return null;
      // Remove extension and decode
      const name = decodeURIComponent(last.replace(/\.[^.]+$/, ""));
      // Only use if it looks like a meaningful name (not a hash/uuid)
      if (name.length > 3 && !/^[a-f0-9-]{20,}$/i.test(name)) {
        return name.replace(/[_-]+/g, " ");
      }
    } catch {
      // invalid URL
    }
    return null;
  }

  // -----------------------------------------------------------------------
  // Main detection logic
  // -----------------------------------------------------------------------

  function matchDomain(hostname) {
    for (const domain of Object.keys(domainExtractors)) {
      if (hostname === domain || hostname.endsWith("." + domain)) {
        return domain;
      }
    }
    return null;
  }

  function detectTitle() {
    const hostname = location.hostname;
    let title = null;

    // 1. Try domain-specific extractor
    const domain = matchDomain(hostname);
    if (domain) {
      try {
        title = domainExtractors[domain]();
      } catch {
        // extractor failed, continue to fallback
      }
    }

    // 2. JSON-LD structured data
    if (!title) title = extractLdJsonTitle();

    // 3. OpenGraph / Twitter meta
    if (!title) title = extractOpenGraphTitle();

    // 4. Video element attributes
    if (!title) title = extractFromVideoElement();

    // 5. Document title (cleaned)
    if (!title) title = extractFromDocumentTitle();

    return title;
  }

  // -----------------------------------------------------------------------
  // Dispatch & observe
  // -----------------------------------------------------------------------

  let lastTitle = null;

  function emitTitle(title, source) {
    if (!title || title === lastTitle) return;
    lastTitle = title;

    document.dispatchEvent(
      new CustomEvent("__stream_grabber_title__", {
        detail: {
          title: title.trim(),
          source,
          url: location.href,
          hostname: location.hostname,
        },
      })
    );
  }

  function runDetection(source) {
    const title = detectTitle();
    if (title) emitTitle(title, source || "auto");
  }

  // Listen for explicit requests from bridge (e.g., popup asking for title)
  window.addEventListener("__stream_grabber_request_title__", () => {
    lastTitle = null; // force re-emit
    runDetection("request");
  });

  // Listen for stream URL hint to try extracting name from URL
  window.addEventListener("__stream_grabber_title_hint__", (e) => {
    const url = e.detail && e.detail.url;
    const fromUrl = extractFromStreamUrl(url);
    if (fromUrl && !lastTitle) {
      emitTitle(fromUrl, "stream_url");
    }
  });

  // -----------------------------------------------------------------------
  // Observation strategy:
  // 1. Run on load
  // 2. Re-run on URL changes (SPA navigation)
  // 3. Re-run on DOM mutations to <head> or title-like elements
  // -----------------------------------------------------------------------

  // Initial detection — wait for DOM to settle
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
      setTimeout(() => runDetection("dom_ready"), 500);
    });
  } else {
    setTimeout(() => runDetection("immediate"), 300);
  }

  // Re-detect on SPA navigation
  let lastUrl = location.href;
  const urlObserver = new MutationObserver(() => {
    if (location.href !== lastUrl) {
      lastUrl = location.href;
      lastTitle = null;
      setTimeout(() => runDetection("navigation"), 800);
    }
  });
  urlObserver.observe(document.documentElement, {
    childList: true,
    subtree: true,
  });

  // Re-detect when <title> or meta tags change
  const headObserver = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (
        mutation.target.tagName === "TITLE" ||
        (mutation.target.tagName === "META" &&
          (mutation.target.getAttribute("property") === "og:title" ||
            mutation.target.getAttribute("name") === "title"))
      ) {
        setTimeout(() => runDetection("head_mutation"), 300);
        return;
      }
    }
  });

  const head = document.head || document.documentElement;
  headObserver.observe(head, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ["content"],
  });
})();

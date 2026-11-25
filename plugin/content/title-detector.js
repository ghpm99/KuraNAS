(function () {
  "use strict";

  const domainExtractors = {
    "youtube.com": function () {
      if (window.ytInitialPlayerResponse) {
        const vd = window.ytInitialPlayerResponse.videoDetails;
        if (vd && vd.title) return vd.title;
      }

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
        }
      }

      const metaTitle = document.querySelector(
        'meta[name="title"], meta[property="og:title"]'
      );
      if (metaTitle && metaTitle.content) return metaTitle.content;

      const h1 = document.querySelector(
        "h1.ytd-watch-metadata yt-formatted-string, #title h1 yt-formatted-string"
      );
      if (h1 && h1.textContent.trim()) return h1.textContent.trim();

      return null;
    },

    "netflix.com": function () {
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
              const metadataCache =
                window.netflix?.reactContext?.models?.userInfo?.data
                  ?.memberContext;
              if (metadataCache) {
              }
            }
          }
        } catch {
        }
      }

      const nfTitle = document.querySelector(
        '[data-uia="video-title"], .video-title h4, .video-title'
      );
      if (nfTitle && nfTitle.textContent.trim())
        return nfTitle.textContent.trim();

      const preloadLink = document.querySelector('link[rel="preload"][as="fetch"][href*="/metadata"]');
      return extractOpenGraphTitle();
    },

    "disneyplus.com": function () {
      const titleEl = document.querySelector(
        '[data-testid="title-field"], .title-field, h2[class*="title"]'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    "primevideo.com": function () {
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

    "twitch.tv": function () {
      const ldJson = extractLdJsonTitle();
      if (ldJson) return ldJson;

      const titleEl = document.querySelector(
        '[data-a-target="stream-title"], h2[data-a-target="stream-title"]'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      const channel = document.querySelector(
        'h1[data-a-target="stream-channel-link"], a[data-a-target="user-channel-header-item"] h1'
      );
      if (channel && channel.textContent.trim())
        return channel.textContent.trim();

      return extractOpenGraphTitle();
    },

    "crunchyroll.com": function () {
      const episodeName = extractEpisodeFromLdJson();
      if (episodeName) return episodeName;

      return extractOpenGraphTitle();
    },

    "vimeo.com": function () {
      if (window.vimeo?.clip_page_config?.clip?.title) {
        return window.vimeo.clip_page_config.clip.title;
      }

      const ldJson = extractLdJsonTitle();
      if (ldJson) return ldJson;

      return extractOpenGraphTitle();
    },

    "dailymotion.com": function () {
      const titleEl = document.querySelector(
        '[class*="VideoInfoTitle"], .VideoInfoTitle'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    "globoplay.globo.com": function () {
      const titleEl = document.querySelector(
        '[class*="playback-title"], .playback-title'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },

    "pluto.tv": function () {
      const titleEl = document.querySelector(
        '[class*="MetadataTitle"], .player-metadata-title'
      );
      if (titleEl && titleEl.textContent.trim())
        return titleEl.textContent.trim();

      return extractOpenGraphTitle();
    },
  };

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
      }
    }
    return null;
  }

  function pad2(n) {
    return String(n).padStart(2, "0");
  }

  function buildEpisodeName(node) {
    const series = node.partOfSeries && node.partOfSeries.name;
    const epNum = node.episodeNumber;
    if (series != null && epNum != null) {
      const season = node.partOfSeason && node.partOfSeason.seasonNumber;
      const se = (season != null ? `S${pad2(season)}` : "") + `E${pad2(epNum)}`;
      return `${series} - ${se}`;
    }
    if (series && node.name) return `${series} - ${node.name}`;
    return null;
  }

  function extractEpisodeFromLdJson() {
    const scripts = document.querySelectorAll(
      'script[type="application/ld+json"]'
    );
    for (const script of scripts) {
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
        if (item && item["@type"] === "TVEpisode") {
          const name = buildEpisodeName(item);
          if (name) return name;
        }
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
      const name = decodeURIComponent(last.replace(/\.[^.]+$/, ""));
      if (name.length > 3 && !/^[a-f0-9-]{20,}$/i.test(name)) {
        return name.replace(/[_-]+/g, " ");
      }
    } catch {
    }
    return null;
  }

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

    const domain = matchDomain(hostname);
    if (domain) {
      try {
        title = domainExtractors[domain]();
      } catch {
      }
    }

    if (!title) title = extractEpisodeFromLdJson();

    if (!title) title = extractLdJsonTitle();

    if (!title) title = extractOpenGraphTitle();

    if (!title) title = extractFromVideoElement();

    if (!title) title = extractFromDocumentTitle();

    return title;
  }

  let lastTitle = null;

  function emitTitle(title, source) {
    if (!title || title === lastTitle) return;
    lastTitle = title;

    window.dispatchEvent(
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

  window.addEventListener("__stream_grabber_request_title__", () => {
    lastTitle = null;
    runDetection("request");
  });

  window.addEventListener("__stream_grabber_title_hint__", (e) => {
    const url = e.detail && e.detail.url;
    const fromUrl = extractFromStreamUrl(url);
    if (fromUrl && !lastTitle) {
      emitTitle(fromUrl, "stream_url");
    }
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
      setTimeout(() => runDetection("dom_ready"), 500);
    });
  } else {
    setTimeout(() => runDetection("immediate"), 300);
  }

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

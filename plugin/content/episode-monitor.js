/* ========================================================================
 * KuraNAS Stream Grabber — Episode Monitor (ISOLATED world)
 *
 * Drives the smart per-episode capture. When an adapter matches the host, it
 * polls the main <video> + the adapter, builds a normalized snapshot, and ships
 * it to the service worker as `episode_state` (consumed by capture-session.js).
 * Hosts with no adapter arm nothing here — the manual hybrid mode still applies.
 * ======================================================================== */

(function () {
  "use strict";

  if (typeof window === "undefined" || typeof document === "undefined") return;

  const registry = globalThis.__kuraEpisodeAdapters;
  if (!registry) return;

  const adapter = registry.resolve(location.hostname);
  if (!adapter) return; // degrade: no smart capture on this host

  function send(message) {
    try {
      chrome.runtime.sendMessage(message).catch(() => {});
    } catch {
      // extension context invalidated
    }
  }

  function getMainVideo() {
    const videos = document.querySelectorAll("video");
    let largest = null;
    let largestArea = 0;
    for (const video of videos) {
      const rect = video.getBoundingClientRect();
      const area = rect.width * rect.height;
      if (area > largestArea) {
        largestArea = area;
        largest = video;
      }
    }
    return largest;
  }

  function isVideoLikelyFullscreen(video) {
    if (document.fullscreenElement) {
      return (
        document.fullscreenElement === video ||
        document.fullscreenElement.contains(video)
      );
    }
    const rect = video.getBoundingClientRect();
    const viewportArea = window.innerWidth * window.innerHeight;
    const videoArea = rect.width * rect.height;
    return viewportArea > 0 && videoArea / viewportArea >= 0.85;
  }

  function buildSnapshot() {
    const episode = adapter.getEpisode({ location, document });
    const video = getMainVideo();
    if (!video) {
      return {
        service: adapter.service,
        episodeId: episode.episodeId,
        title: episode.title,
        hasVideo: false,
        isPlaying: false,
        isEnded: false,
        isFullscreen: false,
        currentTime: 0,
        duration: 0,
      };
    }
    return {
      service: adapter.service,
      episodeId: episode.episodeId,
      title: episode.title,
      hasVideo: true,
      isPlaying: !video.paused && !video.ended,
      isEnded: video.ended,
      isFullscreen: isVideoLikelyFullscreen(video),
      currentTime: video.currentTime || 0,
      duration: Number.isFinite(video.duration) ? video.duration : 0,
    };
  }

  function tick() {
    send({ action: "episode_state", snapshot: buildSnapshot() });
  }

  const videoEventNames = ["play", "pause", "ended", "seeked"];
  for (const name of videoEventNames) {
    document.addEventListener(name, tick, true);
  }
  document.addEventListener("fullscreenchange", tick);
  window.addEventListener("resize", tick);

  // A 1s heartbeat catches the end-of-episode crossing even while the player
  // sits idle (owner asleep) and no DOM event fires.
  setInterval(tick, 1000);
  tick();
})();

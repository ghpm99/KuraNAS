/* ========================================================================
 * KuraNAS Stream Grabber — Bridge (ISOLATED world)
 * Relays between MAIN world events and background service worker.
 * ======================================================================== */

(function () {
  "use strict";

  let monitorActive = false;
  let lastSnapshot = null;
  let rafId = null;

  // -----------------------------------------------------------------------
  // Safe runtime messaging
  // -----------------------------------------------------------------------

  function safeRuntimeSendMessage(msg) {
    try {
      chrome.runtime.sendMessage(msg).catch(() => {});
    } catch {
      // Extension context invalidated
    }
  }

  // -----------------------------------------------------------------------
  // 1. Blob Relay — MAIN → Background
  // -----------------------------------------------------------------------

  window.addEventListener("__stream_grabber_blob__", (e) => {
    const detail = e.detail || {};
    safeRuntimeSendMessage({
      action: "blob_detected",
      blobUrl: detail.blobUrl,
      mimeType: detail.mimeType,
      size: detail.size,
      isMediaSource: detail.isMediaSource,
    });
  });

  window.addEventListener("__stream_grabber_chunk__", (e) => {
    const detail = e.detail || {};
    safeRuntimeSendMessage({
      action: "blob_detected",
      mimeType: detail.mimeType,
      chunkCount: detail.chunkCount,
      totalSize: detail.totalSize,
    });
  });

  // -----------------------------------------------------------------------
  // 1b. Title Relay — MAIN → Background
  // -----------------------------------------------------------------------

  window.addEventListener("__stream_grabber_title__", (e) => {
    const detail = e.detail || {};
    safeRuntimeSendMessage({
      action: "title_detected",
      title: detail.title,
      source: detail.source,
      url: detail.url,
      hostname: detail.hostname,
    });
  });

  // -----------------------------------------------------------------------
  // 2. Capture MediaSource Relay
  // -----------------------------------------------------------------------

  chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
    if (msg.action === "capture_mediasource") {
      window.dispatchEvent(new Event("__stream_grabber_download_request__"));

      const handler = (e) => {
        window.removeEventListener(
          "__stream_grabber_download_response__",
          handler
        );
        sendResponse(e.detail);
      };
      window.addEventListener("__stream_grabber_download_response__", handler);
      return true;
    }

    if (msg.action === "request_title") {
      // Ask MAIN world title-detector to re-run and emit
      window.dispatchEvent(new Event("__stream_grabber_request_title__"));
    }

    if (msg.action === "hybrid_monitor_start") {
      startMonitor();
    }

    if (msg.action === "hybrid_monitor_stop") {
      stopMonitor();
    }
  });

  // -----------------------------------------------------------------------
  // 3. Hybrid Video Monitor
  // -----------------------------------------------------------------------

  function getMainVideo() {
    const videos = document.querySelectorAll("video");
    let largest = null;
    let largestArea = 0;

    for (const v of videos) {
      const rect = v.getBoundingClientRect();
      const area = rect.width * rect.height;
      if (area > largestArea) {
        largestArea = area;
        largest = v;
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

  function buildHybridSnapshot() {
    const video = getMainVideo();
    if (!video) {
      return {
        hasVideo: false,
        isPlaying: false,
        isEnded: false,
        isFullscreen: false,
        url: location.href,
      };
    }

    return {
      hasVideo: true,
      isPlaying: !video.paused && !video.ended,
      isEnded: video.ended,
      isFullscreen: isVideoLikelyFullscreen(video),
      url: location.href,
    };
  }

  function snapshotChanged(a, b) {
    if (!a || !b) return true;
    return (
      a.hasVideo !== b.hasVideo ||
      a.isPlaying !== b.isPlaying ||
      a.isEnded !== b.isEnded ||
      a.isFullscreen !== b.isFullscreen ||
      a.url !== b.url
    );
  }

  function sendSnapshot() {
    const snapshot = buildHybridSnapshot();
    if (snapshotChanged(snapshot, lastSnapshot)) {
      lastSnapshot = snapshot;
      safeRuntimeSendMessage({
        action: "hybrid_video_state",
        snapshot,
      });
    }
  }

  function monitorLoop() {
    if (!monitorActive) return;
    sendSnapshot();
    rafId = requestAnimationFrame(monitorLoop);
  }

  function startMonitor() {
    if (monitorActive) return;
    monitorActive = true;
    lastSnapshot = null;
    bindVideoEvents();
    monitorLoop();
  }

  function stopMonitor() {
    monitorActive = false;
    if (rafId) {
      cancelAnimationFrame(rafId);
      rafId = null;
    }
    unbindVideoEvents();
  }

  // -----------------------------------------------------------------------
  // 4. Event Listeners for Immediate Snapshots
  // -----------------------------------------------------------------------

  const videoEventNames = ["play", "pause", "ended", "seeked"];
  const docEventNames = ["fullscreenchange"];
  const winEventNames = ["resize"];

  function onImmediateEvent() {
    if (monitorActive) sendSnapshot();
  }

  function bindVideoEvents() {
    for (const name of videoEventNames) {
      document.addEventListener(name, onImmediateEvent, true);
    }
    for (const name of docEventNames) {
      document.addEventListener(name, onImmediateEvent);
    }
    for (const name of winEventNames) {
      window.addEventListener(name, onImmediateEvent);
    }
  }

  function unbindVideoEvents() {
    for (const name of videoEventNames) {
      document.removeEventListener(name, onImmediateEvent, true);
    }
    for (const name of docEventNames) {
      document.removeEventListener(name, onImmediateEvent);
    }
    for (const name of winEventNames) {
      window.removeEventListener(name, onImmediateEvent);
    }
  }

  // -----------------------------------------------------------------------
  // 5. URL Change Hooks (SPA navigation)
  // -----------------------------------------------------------------------

  const originalPushState = history.pushState;
  history.pushState = function (...args) {
    originalPushState.apply(this, args);
    if (monitorActive) sendSnapshot();
  };

  const originalReplaceState = history.replaceState;
  history.replaceState = function (...args) {
    originalReplaceState.apply(this, args);
    if (monitorActive) sendSnapshot();
  };

  window.addEventListener("popstate", () => {
    if (monitorActive) sendSnapshot();
  });

  window.addEventListener("hashchange", () => {
    if (monitorActive) sendSnapshot();
  });
})();

/* ========================================================================
 * KuraNAS Stream Grabber — Control Hider strategies (ISOLATED world)
 *
 * Used by "Armar v2": instead of waiting for the player controls to auto-hide,
 * it edits the DOM so the controls are not visible in the recording. The generic
 * strategy lifts the <video> above the page (max z-index) so the control <div>s
 * layered in front of it stop showing. This is fragile (it depends on the
 * player's stacking layout, and it also covers subtitle overlays), so it is a
 * separate mode — the normal Armar (settle-wait) stays available as fallback.
 *
 * Per-site strategies can be registered to do something more precise (e.g. hide
 * only the controls container, keeping subtitles). Resolution is by hostname,
 * falling back to the generic z-index strategy.
 * ======================================================================== */

(function () {
  "use strict";

  if (globalThis.__kuraControlHider) return;

  const strategies = [];

  // Generic: give the <video> the maximum z-index so it paints above the
  // control overlays. z-index only applies to positioned elements, so ensure a
  // non-static position first. Original inline styles are saved for restore.
  const genericStrategy = {
    name: "generic-zindex",
    matches: () => false,
    apply(video) {
      if (!video) return;
      const saved = {
        zIndex: video.style.zIndex,
        position: video.style.position,
      };
      video.dataset.kuraSavedStyle = JSON.stringify(saved);
      if (getComputedStyle(video).position === "static") {
        video.style.position = "relative";
      }
      video.style.zIndex = "2147483647";
    },
    restore(video) {
      if (!video) return;
      let saved = {};
      try {
        saved = JSON.parse(video.dataset.kuraSavedStyle || "{}");
      } catch {
        saved = {};
      }
      video.style.zIndex = saved.zIndex || "";
      video.style.position = saved.position || "";
      delete video.dataset.kuraSavedStyle;
    },
  };

  function register(strategy) {
    strategies.push(strategy);
  }

  function resolve(hostname) {
    for (const strategy of strategies) {
      try {
        if (strategy.matches && strategy.matches(hostname)) return strategy;
      } catch {
        // a broken strategy never blocks resolution
      }
    }
    return genericStrategy;
  }

  globalThis.__kuraControlHider = { register, resolve, generic: genericStrategy };
})();

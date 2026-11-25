(function () {
  "use strict";

  if (globalThis.__kuraControlHider) return;

  const strategies = [];

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
      }
    }
    return genericStrategy;
  }

  globalThis.__kuraControlHider = { register, resolve, generic: genericStrategy };
})();

/* ========================================================================
 * KuraNAS Stream Grabber — Popup UI
 * ======================================================================== */

(function () {
  "use strict";

  const $ = (sel) => document.querySelector(sel);
  const mediaListEl = $("#mediaList");
  const emptyStateEl = $("#emptyState");
  const btnArm = $("#btnArm");
  const btnDisarm = $("#btnDisarm");
  const btnStopNow = $("#btnStopNow");
  const hybridStatusEl = $("#hybridStatus");
  const qualityModal = $("#qualityModal");
  const qualityListEl = $("#qualityList");
  const btnCloseModal = $("#btnCloseModal");
  const apiUrlInput = $("#apiUrl");
  const btnSaveSettings = $("#btnSaveSettings");
  const detectedTitleEl = $("#detectedTitle");

  let currentTabId = null;

  // -----------------------------------------------------------------------
  // Init
  // -----------------------------------------------------------------------

  async function init() {
    const [tab] = await chrome.tabs.query({
      active: true,
      currentWindow: true,
    });
    if (!tab) return;
    currentTabId = tab.id;

    loadSettings();
    loadMedia();
    loadHybridStatus();
    loadDetectedTitle();

    // Ask content script to re-detect title (in case it wasn't detected yet)
    chrome.tabs.sendMessage(currentTabId, { action: "request_title" }).catch(() => {});
  }

  function loadDetectedTitle() {
    chrome.runtime.sendMessage(
      { action: "get_title", tabId: currentTabId },
      (response) => {
        if (response && response.title && detectedTitleEl) {
          detectedTitleEl.textContent = response.title;
          detectedTitleEl.title = `Fonte: ${response.source}`;
          detectedTitleEl.classList.remove("hidden");
        }
      }
    );
  }

  function loadSettings() {
    chrome.storage.sync.get("apiBaseUrl", (result) => {
      if (result.apiBaseUrl) {
        apiUrlInput.value = result.apiBaseUrl;
      }
    });
  }

  function loadMedia() {
    chrome.runtime.sendMessage(
      { action: "get_media", tabId: currentTabId },
      (response) => {
        if (response && response.media) {
          renderMediaList(response.media);
        }
      }
    );
  }

  function loadHybridStatus() {
    chrome.runtime.sendMessage(
      { action: "get_hybrid_status", tabId: currentTabId },
      (response) => {
        if (response) {
          updateHybridUI(response);
        }
      }
    );
  }

  // -----------------------------------------------------------------------
  // Render Media List
  // -----------------------------------------------------------------------

  function renderMediaList(mediaItems) {
    if (!mediaItems || mediaItems.length === 0) {
      emptyStateEl.classList.remove("hidden");
      return;
    }

    emptyStateEl.classList.add("hidden");

    const sorted = [...mediaItems].sort((a, b) => {
      const order = { hls: 0, dash: 1, blob: 2 };
      const oa = order[a.type] ?? 3;
      const ob = order[b.type] ?? 3;
      return oa - ob;
    });

    const existingItems = mediaListEl.querySelectorAll(".media-item");
    existingItems.forEach((el) => el.remove());

    for (const item of sorted) {
      mediaListEl.appendChild(createMediaItem(item));
    }
  }

  function createMediaItem(item) {
    const el = document.createElement("div");
    el.className = "media-item";

    const badge = document.createElement("span");
    badge.className = `media-type-badge ${item.type}`;
    badge.textContent = item.type.toUpperCase();

    const info = document.createElement("div");
    info.className = "media-info";

    const urlText = document.createElement("div");
    urlText.className = "media-url";
    urlText.textContent = truncateUrl(item.url);
    urlText.title = item.url;
    info.appendChild(urlText);

    if (item.size) {
      const sizeText = document.createElement("div");
      sizeText.className = "media-url";
      sizeText.textContent = formatSize(item.size);
      info.appendChild(sizeText);
    }

    const actions = document.createElement("div");
    actions.className = "media-actions";

    if (item.type === "hls") {
      actions.appendChild(
        createActionButton("Baixar", "btn btn-primary btn-small", () =>
          handleHLSDownload(item)
        )
      );
    } else if (item.type === "dash") {
      actions.appendChild(
        createActionButton("Baixar", "btn btn-primary btn-small", () =>
          handleDASHDownload(item)
        )
      );
    } else if (item.type === "blob") {
      actions.appendChild(
        createActionButton("Capturar", "btn btn-primary btn-small", () =>
          handleBlobCapture(item)
        )
      );
    } else {
      actions.appendChild(
        createActionButton("Baixar", "btn btn-primary btn-small", () =>
          handleDirectDownload(item)
        )
      );
    }

    actions.appendChild(
      createActionButton("Copiar", "btn btn-secondary btn-small", () =>
        copyToClipboard(item.url)
      )
    );

    el.appendChild(badge);
    el.appendChild(info);
    el.appendChild(actions);

    return el;
  }

  function createActionButton(label, className, onClick) {
    const btn = document.createElement("button");
    btn.className = className;
    btn.textContent = label;
    btn.addEventListener("click", onClick);
    return btn;
  }

  // -----------------------------------------------------------------------
  // Download Handlers
  // -----------------------------------------------------------------------

  async function getAutoTitle() {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(
        { action: "get_title", tabId: currentTabId },
        (response) => {
          resolve(response && response.title ? response.title : null);
        }
      );
    });
  }

  async function promptName(fallbackName) {
    const autoTitle = await getAutoTitle();
    const suggestion = autoTitle || fallbackName || "";
    const name = prompt("Nome da captura:", suggestion);
    return name ? name.trim() : null;
  }

  async function handleHLSDownload(item) {
    const name = await promptName("video_hls");
    if (!name) return;

    const response = await chrome.runtime.sendMessage({
      action: "download_hls",
      url: item.url,
      name,
    });

    if (response && response.type === "master") {
      showQualityModal("HLS", response.variants, (variant) => {
        chrome.runtime.sendMessage({
          action: "download_hls",
          url: variant.url,
          name,
        });
      });
    }
  }

  async function handleDASHDownload(item) {
    const name = await promptName("video_dash");
    if (!name) return;

    const response = await chrome.runtime.sendMessage({
      action: "download_dash",
      url: item.url,
      name,
    });

    if (response && response.type === "dash_manifest") {
      showQualityModal("DASH", response.representations, (rep) => {
        chrome.runtime.sendMessage({
          action: "download_dash",
          url: item.url,
          name,
          representationId: rep.id,
        });
      });
    }
  }

  async function handleBlobCapture(item) {
    const name = await promptName("captura_blob");
    if (!name) return;

    chrome.runtime.sendMessage({
      action: "upload_blob_capture",
      tabId: currentTabId,
      blobUrl: item.url,
      name,
    });
  }

  async function handleDirectDownload(item) {
    const name = await promptName("video_direto");
    if (!name) return;

    chrome.runtime.sendMessage({
      action: "download_direct",
      url: item.url,
      name,
    });
  }

  // -----------------------------------------------------------------------
  // Quality Modal
  // -----------------------------------------------------------------------

  function showQualityModal(type, items, onSelect) {
    qualityListEl.innerHTML = "";

    for (const item of items) {
      const li = document.createElement("li");
      const label =
        type === "HLS"
          ? `${item.resolution || "?"} - ${formatBitrate(item.bandwidth)}`
          : `${item.width}x${item.height} - ${formatBitrate(item.bandwidth)} (${item.codecs})`;

      const labelSpan = document.createElement("span");
      labelSpan.textContent = label;
      li.appendChild(labelSpan);

      li.addEventListener("click", () => {
        onSelect(item);
        closeQualityModal();
      });

      qualityListEl.appendChild(li);
    }

    qualityModal.classList.remove("hidden");
  }

  function closeQualityModal() {
    qualityModal.classList.add("hidden");
  }

  // -----------------------------------------------------------------------
  // Hybrid Controls
  // -----------------------------------------------------------------------

  function updateHybridUI(status) {
    hybridStatusEl.className = "hybrid-badge";

    if (!status.armed) {
      hybridStatusEl.textContent = "";
      btnArm.classList.remove("hidden");
      btnDisarm.classList.add("hidden");
      btnStopNow.classList.add("hidden");
      return;
    }

    btnArm.classList.add("hidden");
    btnDisarm.classList.remove("hidden");

    if (status.state === "RECORDING") {
      hybridStatusEl.textContent = "REC";
      hybridStatusEl.classList.add("recording");
      btnStopNow.classList.remove("hidden");
    } else if (status.state === "ARMED") {
      hybridStatusEl.textContent = "ARMADO";
      hybridStatusEl.classList.add("armed");
      btnStopNow.classList.add("hidden");
    } else {
      hybridStatusEl.textContent = status.state;
      btnStopNow.classList.add("hidden");
    }
  }

  // -----------------------------------------------------------------------
  // Event Listeners
  // -----------------------------------------------------------------------

  btnArm.addEventListener("click", () => {
    chrome.runtime.sendMessage({ action: "hybrid_arm", tabId: currentTabId });
  });

  btnDisarm.addEventListener("click", () => {
    chrome.runtime.sendMessage({
      action: "hybrid_disarm",
      tabId: currentTabId,
    });
  });

  btnStopNow.addEventListener("click", () => {
    chrome.runtime.sendMessage({
      action: "hybrid_stop_now",
      tabId: currentTabId,
    });
  });

  btnCloseModal.addEventListener("click", closeQualityModal);

  btnSaveSettings.addEventListener("click", () => {
    const url = apiUrlInput.value.trim();
    chrome.storage.sync.set({ apiBaseUrl: url });
  });

  // Real-time updates
  chrome.runtime.onMessage.addListener((msg) => {
    if (msg.action === "media_detected" && msg.tabId === currentTabId) {
      loadMedia();
    }
    if (msg.action === "hybrid_status" && msg.tabId === currentTabId) {
      updateHybridUI(msg.status);
    }
    if (msg.action === "title_detected" && detectedTitleEl) {
      detectedTitleEl.textContent = msg.title;
      detectedTitleEl.title = `Fonte: ${msg.source}`;
      detectedTitleEl.classList.remove("hidden");
    }
  });

  // -----------------------------------------------------------------------
  // Utilities
  // -----------------------------------------------------------------------

  function truncateUrl(url) {
    if (!url) return "";
    if (url.length <= 50) return url;
    return url.substring(0, 25) + "..." + url.substring(url.length - 22);
  }

  function formatSize(bytes) {
    if (!bytes) return "";
    const units = ["B", "KB", "MB", "GB"];
    let idx = 0;
    let size = bytes;
    while (size >= 1024 && idx < units.length - 1) {
      size /= 1024;
      idx++;
    }
    return `${size.toFixed(1)} ${units[idx]}`;
  }

  function formatBitrate(bps) {
    if (!bps) return "?";
    if (bps >= 1000000) return `${(bps / 1000000).toFixed(1)} Mbps`;
    return `${(bps / 1000).toFixed(0)} kbps`;
  }

  function copyToClipboard(text) {
    navigator.clipboard.writeText(text).catch(() => {});
  }

  // -----------------------------------------------------------------------
  init();
})();

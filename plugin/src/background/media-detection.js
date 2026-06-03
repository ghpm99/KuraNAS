export function classifyByUrl(url, mediaPatterns) {
  for (const { regex, type } of mediaPatterns) {
    if (regex.test(url)) return type;
  }
  return null;
}

export function classifyByContentType(contentType, mediaContentTypes) {
  for (const { pattern, type } of mediaContentTypes) {
    if (pattern.test(contentType)) return type;
  }
  return null;
}

export function createMediaDetectionManager({
  chromeApi,
  detectedMedia,
  mediaPatterns,
  mediaContentTypes,
  now = () => Date.now(),
}) {
  function updateBadge(tabId) {
    const list = detectedMedia.get(tabId) || [];
    const text = list.length > 0 ? String(list.length) : "";
    chromeApi.action.setBadgeText({ text, tabId }).catch(() => {});
    chromeApi.action
      .setBadgeBackgroundColor({ color: "#4CAF50", tabId })
      .catch(() => {});
  }

  function addMedia(tabId, item) {
    if (!detectedMedia.has(tabId)) {
      detectedMedia.set(tabId, []);
    }

    const list = detectedMedia.get(tabId);
    const isDuplicate = list.some(
      (mediaItem) => mediaItem.url === item.url && mediaItem.type === item.type
    );
    if (isDuplicate) return;

    list.push(item);
    updateBadge(tabId);

    chromeApi.runtime
      .sendMessage({ action: "media_detected", tabId, item })
      .catch(() => {});
  }

  function registerNetworkListeners() {
    chromeApi.webRequest.onBeforeRequest.addListener(
      (details) => {
        if (details.tabId < 0) return;
        const type = classifyByUrl(details.url, mediaPatterns);
        if (type) {
          addMedia(details.tabId, {
            url: details.url,
            type,
            source: "network",
            timestamp: now(),
          });
        }
      },
      { urls: ["<all_urls>"] }
    );

    chromeApi.webRequest.onHeadersReceived.addListener(
      (details) => {
        if (details.tabId < 0) return;
        const ctHeader = (details.responseHeaders || []).find(
          (header) => header.name.toLowerCase() === "content-type"
        );
        if (!ctHeader) return;

        const type = classifyByContentType(ctHeader.value, mediaContentTypes);
        if (type) {
          addMedia(details.tabId, {
            url: details.url,
            type,
            source: "network",
            timestamp: now(),
          });
        }
      },
      { urls: ["<all_urls>"] },
      ["responseHeaders"]
    );
  }

  return {
    addMedia,
    registerNetworkListeners,
    updateBadge,
  };
}

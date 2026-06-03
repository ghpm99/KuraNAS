export function routeRuntimeMessage(msg, sender, sendResponse, handlers) {
  const tabId = sender && sender.tab ? sender.tab.id : msg.tabId;

  switch (msg.action) {
    case "blob_detected":
      handlers.handleBlobDetected(tabId, msg);
      return false;

    case "title_detected":
      handlers.handleTitleDetected(tabId, msg);
      return false;

    case "get_title":
      sendResponse(handlers.getTitleForTab(msg.tabId));
      return false;

    case "hybrid_video_state":
      handlers.handleHybridVideoState(tabId, msg.snapshot);
      return false;

    case "get_media":
      sendResponse({ media: handlers.getDetectedMedia(msg.tabId) });
      return false;

    case "hybrid_arm":
      handlers.armHybrid(msg.tabId);
      sendResponse({ ok: true });
      return false;

    case "hybrid_disarm":
      handlers.disarmHybrid(msg.tabId);
      sendResponse({ ok: true });
      return false;

    case "hybrid_stop_now":
      handlers.stopHybridRecording(msg.tabId);
      sendResponse({ ok: true });
      return false;

    case "hybrid_offscreen_started":
      handlers.handleOffscreenStarted(tabId, msg);
      return false;

    case "hybrid_offscreen_stopped":
      handlers.handleOffscreenStopped(tabId, msg);
      return false;

    case "hybrid_offscreen_error":
      handlers.handleOffscreenError(tabId, msg);
      return false;

    case "hybrid_recording_chunk":
      handlers.handleHybridRecordingChunk(tabId, msg);
      return false;

    case "hybrid_recording_complete":
      handlers.handleHybridRecordingComplete(tabId);
      return false;

    case "hybrid_save_recording_blob":
      handlers.handleSaveRecordingBlob(msg);
      return false;

    case "download_hls":
      handlers.downloadHLS(msg.url, msg.name).then(sendResponse);
      return true;

    case "download_dash":
      handlers.downloadDASH(msg.url, msg.name).then(sendResponse);
      return true;

    case "download_direct":
      handlers.downloadDirect(msg.url, msg.name).then(sendResponse);
      return true;

    case "upload_blob_capture":
      handlers.uploadBlobCapture(msg.tabId, msg.blobUrl, msg.name).then(sendResponse);
      return true;

    case "get_hybrid_status":
      sendResponse(handlers.getHybridStatus(msg.tabId));
      return false;

    default:
      return false;
  }
}

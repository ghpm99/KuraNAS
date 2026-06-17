/* ========================================================================
 * KuraNAS Stream Grabber — Offscreen Recorder
 * Invisible document that performs actual tab capture recording.
 * ======================================================================== */

const recordings = new Map();

const LOG_PREFIX = "[KuraNAS rec]";

function logRec(tabId, ...args) {
  console.log(LOG_PREFIX, `tab=${tabId}`, ...args);
  // Mirror into the service worker console (the offscreen doc console is hidden).
  try {
    chrome.runtime
      .sendMessage({ action: "kuranas_log", source: LOG_PREFIX, args: [`tab=${tabId}`, ...args] })
      .catch(() => {});
  } catch (_e) {
    // sendMessage can throw if the SW is momentarily unavailable; the local
    // console.log above still happened.
  }
}

// describeTrack dumps everything that tells empty-frame-by-DRM apart from other
// causes: a Widevine-protected tab capture typically hands back a video track
// that is `muted: true` and/or reports 0x0 with no real frames, while a normal
// capture reports the page's dimensions and `muted: false`.
function describeTrack(track) {
  const settings = typeof track.getSettings === "function" ? track.getSettings() : {};
  return {
    kind: track.kind,
    label: track.label,
    enabled: track.enabled,
    muted: track.muted,
    readyState: track.readyState,
    width: settings.width,
    height: settings.height,
    frameRate: settings.frameRate,
  };
}

function logStreamDiagnostics(tabId, stream) {
  const videoTracks = stream.getVideoTracks();
  const audioTracks = stream.getAudioTracks();

  logRec(tabId, `stream obtido: ${videoTracks.length} vídeo, ${audioTracks.length} áudio`);

  videoTracks.forEach((track, i) => {
    const info = describeTrack(track);
    logRec(tabId, `vídeo[${i}]`, info);
    if (info.muted) {
      console.warn(
        LOG_PREFIX,
        `tab=${tabId}`,
        "track de vídeo está MUTED — forte indício de DRM/Widevine (frames não chegam)."
      );
    }
    // DRM/policy changes can mute the track AFTER capture starts; watch for it.
    track.onmute = () =>
      console.warn(LOG_PREFIX, `tab=${tabId}`, "vídeo MUTED durante a gravação (DRM?).");
    track.onunmute = () => logRec(tabId, "vídeo UNMUTED (frames voltaram).");
    track.onended = () => logRec(tabId, "track de vídeo ENDED.");
  });

  audioTracks.forEach((track, i) => logRec(tabId, `áudio[${i}]`, describeTrack(track)));
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.action === "offscreen_start_recording") {
    startRecording(msg.tabId, msg.streamId, Boolean(msg.streamUpload));
  }
  if (msg.action === "offscreen_stop_recording") {
    stopRecording(msg.tabId);
  }
});

async function startRecording(tabId, streamId, streamUpload) {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: {
        mandatory: {
          chromeMediaSource: "tab",
          chromeMediaSourceId: streamId,
        },
      },
      video: {
        mandatory: {
          chromeMediaSource: "tab",
          chromeMediaSourceId: streamId,
          maxWidth: 3840,
          maxHeight: 2160,
          maxFrameRate: 60,
        },
      },
    });

    logStreamDiagnostics(tabId, stream);

    const mimeType = selectMimeType();
    logRec(tabId, `MediaRecorder mimeType=${mimeType}, streamUpload=${streamUpload}`);
    const recorder = new MediaRecorder(stream, { mimeType });
    const chunks = [];

    let dataEvents = 0;
    let emptyDataEvents = 0;
    let streamedBytes = 0;
    const startedAt = Date.now();

    recorder.onstart = () => logRec(tabId, "recorder START");

    recorder.ondataavailable = (e) => {
      dataEvents += 1;
      if (e.data.size === 0) {
        emptyDataEvents += 1;
        logRec(
          tabId,
          `ondataavailable #${dataEvents}: 0 byte (vazio ${emptyDataEvents}x) — sem frames (DRM?)`
        );
        return;
      }

      logRec(tabId, `ondataavailable #${dataEvents}: ${e.data.size} bytes`);

      if (streamUpload) {
        // A Blob does NOT survive chrome.runtime.sendMessage (JSON serialization
        // turns it into {}), which silently dropped every streamed chunk. Pass a
        // blob: URL string instead — the service worker fetches it to get the
        // bytes. Same proven trick the non-stream path uses for the final blob.
        streamedBytes += e.data.size;
        const chunkUrl = URL.createObjectURL(e.data);
        chrome.runtime
          .sendMessage({
            action: "hybrid_recording_chunk",
            tabId,
            chunkUrl,
            size: e.data.size,
          })
          .catch(() => {});
        return;
      }

      chunks.push(e.data);
    };

    recorder.onstop = () => {
      const elapsedMs = Date.now() - startedAt;
      const totalBytes = chunks.reduce((sum, chunk) => sum + chunk.size, 0);
      logRec(
        tabId,
        `recorder STOP: ${elapsedMs}ms, ${dataEvents} eventos (${emptyDataEvents} vazios), ` +
          (streamUpload
            ? `${streamedBytes} bytes transmitidos (stream)`
            : `${totalBytes} bytes acumulados`)
      );

      if (streamUpload) {
        chrome.runtime
          .sendMessage({
            action: "hybrid_recording_complete",
            tabId,
          })
          .catch(() => {});
      } else {
        if (totalBytes === 0) {
          // The stream produced no media data, so the recording is empty. This is
          // typical of DRM-protected playback (e.g. Widevine on Crunchyroll) or a
          // recording stopped before any frame was captured. Fail loudly here
          // instead of creating an empty blob that the backend would reject.
          // The diagnostics above (track muted?, dataEvents, elapsedMs) tell which.
          const likelyDrm = dataEvents === 0 || emptyDataEvents === dataEvents;
          console.warn(
            LOG_PREFIX,
            `tab=${tabId}`,
            `Gravação vazia: ${dataEvents} eventos, ${emptyDataEvents} vazios, ${elapsedMs}ms. ` +
              (likelyDrm
                ? "Nenhum frame chegou — provável DRM/Widevine."
                : "Houve dados mas o total é 0 — verifique acima.")
          );
          chrome.runtime.sendMessage({
            action: "hybrid_offscreen_error",
            tabId,
            error:
              "Gravação vazia: o stream não produziu dados (conteúdo protegido por DRM ou gravação muito curta).",
          });
        } else {
          const blob = new Blob(chunks, { type: mimeType });
          const url = URL.createObjectURL(blob);

          chrome.runtime.sendMessage({
            action: "hybrid_save_recording_blob",
            tabId,
            blobUrl: url,
            name: `recording_${tabId}_${Date.now()}`,
          });
        }
      }

      chrome.runtime.sendMessage({
        action: "hybrid_offscreen_stopped",
        tabId,
      });

      cleanupRecording(tabId);
    };

    recorder.onerror = (e) => {
      const message = e.error ? e.error.message : "Unknown recorder error";
      console.error(LOG_PREFIX, `tab=${tabId}`, "recorder ERROR:", message);
      chrome.runtime.sendMessage({
        action: "hybrid_offscreen_error",
        tabId,
        error: message,
      });
      cleanupRecording(tabId);
    };

    recordings.set(tabId, { recorder, stream, url: null });
    recorder.start(1000);
    logRec(tabId, "recorder.start(1000) chamado");

    chrome.runtime.sendMessage({
      action: "hybrid_offscreen_started",
      tabId,
    });
  } catch (err) {
    // getUserMedia/tabCapture failed outright (permission, invalid streamId, or
    // the capture was denied for the protected content).
    console.error(LOG_PREFIX, `tab=${tabId}`, "falha ao iniciar captura:", err.message);
    chrome.runtime.sendMessage({
      action: "hybrid_offscreen_error",
      tabId,
      error: err.message,
    });
  }
}

function stopRecording(tabId) {
  const rec = recordings.get(tabId);
  if (!rec) return;

  if (rec.recorder.state !== "inactive") {
    rec.recorder.stop();
  }
}

function cleanupRecording(tabId) {
  const rec = recordings.get(tabId);
  if (!rec) return;
  rec.stream.getTracks().forEach((track) => track.stop());
  if (rec.url) {
    URL.revokeObjectURL(rec.url);
  }
  recordings.delete(tabId);
}

function selectMimeType() {
  const preferred = [
    "video/webm;codecs=vp9,opus",
    "video/webm;codecs=vp8,opus",
    "video/webm",
  ];

  for (const type of preferred) {
    if (MediaRecorder.isTypeSupported(type)) return type;
  }

  return "video/webm";
}

/* ========================================================================
 * KuraNAS Stream Grabber — Offscreen Recorder
 * Invisible document that performs actual tab capture recording.
 * ======================================================================== */

const recordings = new Map();

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

    const mimeType = selectMimeType();
    const recorder = new MediaRecorder(stream, { mimeType });
    const chunks = [];

    recorder.ondataavailable = (e) => {
      if (e.data.size > 0) {
        if (streamUpload) {
          chrome.runtime
            .sendMessage({
              action: "hybrid_recording_chunk",
              tabId,
              chunk: e.data,
            })
            .catch(() => {});
          return;
        }

        chunks.push(e.data);
      }
    };

    recorder.onstop = () => {
      if (streamUpload) {
        chrome.runtime
          .sendMessage({
            action: "hybrid_recording_complete",
            tabId,
          })
          .catch(() => {});
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

      chrome.runtime.sendMessage({
        action: "hybrid_offscreen_stopped",
        tabId,
      });

      cleanupRecording(tabId);
    };

    recorder.onerror = (e) => {
      chrome.runtime.sendMessage({
        action: "hybrid_offscreen_error",
        tabId,
        error: e.error ? e.error.message : "Unknown recorder error",
      });
      cleanupRecording(tabId);
    };

    recordings.set(tabId, { recorder, stream, url: null });
    recorder.start(1000);

    chrome.runtime.sendMessage({
      action: "hybrid_offscreen_started",
      tabId,
    });
  } catch (err) {
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

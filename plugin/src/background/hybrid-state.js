export function createHybridStateMachine({
  hybridStates,
  broadcastHybridStatus,
  initHybridUploadSession,
  ensureOffscreen,
  getMediaStreamId,
  sendRuntimeMessage,
  sendTabMessage,
  stopOffscreenRecording,
  hybridStabilityMs,
  hybridStopGraceMs,
  setTimeoutFn = setTimeout,
  clearTimeoutFn = clearTimeout,
}) {
  function getHybridStatus(tabId) {
    const state = hybridStates.get(tabId);
    if (!state) return { armed: false, state: "IDLE" };
    return {
      armed: state.armed,
      state: state.recordingState,
      monitorEnabled: state.monitorEnabled,
    };
  }

  function armHybrid(tabId) {
    let state = hybridStates.get(tabId);
    if (!state) {
      state = {
        armed: true,
        monitorEnabled: true,
        recordingState: "ARMED",
        stabilityTimer: null,
        graceTimer: null,
        lastSnapshot: null,
        recording: false,
        uploadSession: null,
      };
      hybridStates.set(tabId, state);
    } else {
      state.armed = true;
      state.monitorEnabled = true;
      state.recordingState = "ARMED";
    }

    sendTabMessage(tabId, { action: "hybrid_monitor_start" });
    broadcastHybridStatus(tabId);
  }

  function disarmHybrid(tabId) {
    const state = hybridStates.get(tabId);
    if (!state) return;

    clearTimeoutFn(state.stabilityTimer);
    clearTimeoutFn(state.graceTimer);
    state.armed = false;
    state.monitorEnabled = false;
    state.recordingState = "IDLE";

    if (state.recording) {
      stopOffscreenRecording(tabId);
    }
    state.uploadSession = null;

    sendTabMessage(tabId, { action: "hybrid_monitor_stop" });
    broadcastHybridStatus(tabId);
  }

  function handleHybridVideoState(tabId, snapshot) {
    const state = hybridStates.get(tabId);
    if (!state || !state.armed) return;

    state.lastSnapshot = snapshot;

    const shouldRecord =
      snapshot.hasVideo &&
      snapshot.isPlaying &&
      snapshot.isFullscreen &&
      !snapshot.isEnded;

    if (state.recordingState === "ARMED") {
      if (shouldRecord) {
        clearTimeoutFn(state.stabilityTimer);
        state.stabilityTimer = setTimeoutFn(() => {
          if (state.armed && state.recordingState === "ARMED") {
            startHybridRecording(tabId);
          }
        }, hybridStabilityMs);
      } else {
        clearTimeoutFn(state.stabilityTimer);
      }
    } else if (state.recordingState === "RECORDING") {
      if (snapshot.isEnded) {
        clearTimeoutFn(state.graceTimer);
        stopHybridRecording(tabId);
      } else if (!shouldRecord) {
        if (!state.graceTimer) {
          state.graceTimer = setTimeoutFn(() => {
            if (state.recordingState === "RECORDING") {
              stopHybridRecording(tabId);
            }
          }, hybridStopGraceMs);
        }
      } else {
        clearTimeoutFn(state.graceTimer);
        state.graceTimer = null;
      }
    }

    if (state.lastUrl && snapshot.url !== state.lastUrl) {
      if (state.recordingState === "RECORDING") {
        stopHybridRecording(tabId);
      }
    }
    state.lastUrl = snapshot.url;
  }

  async function startHybridRecording(tabId) {
    const state = hybridStates.get(tabId);
    if (!state) return;
    state.recordingState = "RECORDING";
    broadcastHybridStatus(tabId);

    try {
      await initHybridUploadSession(tabId);
      const streamId = await getMediaStreamId(tabId);
      await ensureOffscreen();
      sendRuntimeMessage({
        action: "offscreen_start_recording",
        tabId,
        streamId,
        streamUpload: true,
      });
    } catch {
      state.uploadSession = null;
      state.recordingState = "ARMED";
      broadcastHybridStatus(tabId);
    }
  }

  function stopHybridRecording(tabId) {
    const state = hybridStates.get(tabId);
    if (!state) return;

    clearTimeoutFn(state.stabilityTimer);
    clearTimeoutFn(state.graceTimer);
    state.graceTimer = null;
    state.recordingState = "STOPPED";
    broadcastHybridStatus(tabId);

    stopOffscreenRecording(tabId);

    setTimeoutFn(() => {
      if (state.armed) {
        state.recordingState = "ARMED";
        broadcastHybridStatus(tabId);
      }
    }, 1000);
  }

  function handleOffscreenStarted(tabId) {
    const state = hybridStates.get(tabId);
    if (state) {
      state.recording = true;
    }
  }

  function handleOffscreenStopped(tabId) {
    const state = hybridStates.get(tabId);
    if (state) {
      state.recording = false;
    }
  }

  function handleOffscreenError(tabId) {
    const state = hybridStates.get(tabId);
    if (state) {
      state.recording = false;
      state.uploadSession = null;
      state.recordingState = "ARMED";
      broadcastHybridStatus(tabId);
    }
  }

  function cleanupTab(tabId) {
    const state = hybridStates.get(tabId);
    if (!state) return;
    clearTimeoutFn(state.stabilityTimer);
    clearTimeoutFn(state.graceTimer);
    if (state.recording) {
      stopOffscreenRecording(tabId);
    }
    hybridStates.delete(tabId);
  }

  return {
    armHybrid,
    cleanupTab,
    disarmHybrid,
    getHybridStatus,
    handleHybridVideoState,
    handleOffscreenError,
    handleOffscreenStarted,
    handleOffscreenStopped,
    startHybridRecording,
    stopHybridRecording,
  };
}

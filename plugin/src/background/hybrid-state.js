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
  hybridPrepareSettleMs = 3000,
  setTimeoutFn = setTimeout,
  clearTimeoutFn = clearTimeout,
}) {
  const LOG = "[KuraNAS][hybrid]";
  function log(...args) {
    console.log(LOG, ...args);
  }

  function getHybridStatus(tabId) {
    const state = hybridStates.get(tabId);
    if (!state) return { armed: false, state: "IDLE" };
    return {
      armed: state.armed,
      state: state.recordingState,
      monitorEnabled: state.monitorEnabled,
    };
  }

  // mode: "auto" (Armar — wait for controls to auto-hide) or "dom" (Armar v2 —
  // edit the DOM to hide controls). Defaults to "auto".
  function armHybrid(tabId, mode = "auto") {
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
        preparing: false,
        mode,
        uploadSession: null,
      };
      hybridStates.set(tabId, state);
    } else {
      state.armed = true;
      state.monitorEnabled = true;
      state.recordingState = "ARMED";
      state.preparing = false;
      state.mode = mode;
    }

    log("armado", { tabId, mode });
    sendTabMessage(tabId, { action: "hybrid_monitor_start" });
    broadcastHybridStatus(tabId);
  }

  function disarmHybrid(tabId) {
    const state = hybridStates.get(tabId);
    if (!state) return;
    log("desarmado", { tabId, recordingState: state.recordingState });

    clearTimeoutFn(state.stabilityTimer);
    clearTimeoutFn(state.graceTimer);
    state.armed = false;
    state.monitorEnabled = false;
    state.recordingState = "IDLE";
    state.preparing = false;

    if (state.recording) {
      stopOffscreenRecording(tabId);
    }
    if (state.mode === "dom") {
      sendTabMessage(tabId, { action: "hybrid_restore_dom" });
    }
    state.uploadSession = null;

    sendTabMessage(tabId, { action: "hybrid_monitor_stop" });
    broadcastHybridStatus(tabId);
  }

  function handleHybridVideoState(tabId, snapshot) {
    const state = hybridStates.get(tabId);
    if (!state || !state.armed) {
      log("snapshot ignorado (não armado)", { tabId, armed: state?.armed, snapshot });
      return;
    }

    state.lastSnapshot = snapshot;

    const shouldRecord =
      snapshot.hasVideo &&
      snapshot.isPlaying &&
      snapshot.isFullscreen &&
      !snapshot.isEnded;

    log("avaliando gatilho", {
      tabId,
      recordingState: state.recordingState,
      preparing: state.preparing,
      shouldRecord,
      bloqueadoPor: shouldRecord
        ? null
        : {
            hasVideo: snapshot.hasVideo,
            isPlaying: snapshot.isPlaying,
            isFullscreen: snapshot.isFullscreen,
            isEnded: snapshot.isEnded,
          },
      snapshot,
      stabilityMs: hybridStabilityMs,
    });

    if (state.recordingState === "ARMED") {
      // Precondition met (playing + fullscreen). Ask the page to rewind to 0,
      // let the controls overlay fade, then resume playback — only then record,
      // so the capture starts clean from the beginning. Fire once per arm.
      if (shouldRecord) {
        if (!state.preparing) {
          clearTimeoutFn(state.stabilityTimer);
          log("condições OK — aguardando estabilidade antes de preparar", { tabId, stabilityMs: hybridStabilityMs });
          state.stabilityTimer = setTimeoutFn(() => {
            if (state.armed && state.recordingState === "ARMED" && !state.preparing) {
              state.preparing = true;
              log("estabilidade confirmada — disparando prepare_capture", { tabId, mode: state.mode || "auto" });
              sendTabMessage(tabId, {
                action: "hybrid_prepare_capture",
                mode: state.mode || "auto",
                settleMs: hybridPrepareSettleMs,
              });
            } else {
              log("estabilidade expirou mas estado mudou — prepare cancelado", {
                tabId,
                armed: state.armed,
                recordingState: state.recordingState,
                preparing: state.preparing,
              });
            }
          }, hybridStabilityMs);
        } else {
          log("condições OK mas já está preparando — ignorado", { tabId });
        }
      } else {
        log("condições não atendidas — timer de estabilidade cancelado", { tabId });
        clearTimeoutFn(state.stabilityTimer);
      }
    } else if (state.recordingState === "RECORDING") {
      if (snapshot.isEnded) {
        log("vídeo terminou — parando gravação", { tabId });
        clearTimeoutFn(state.graceTimer);
        stopHybridRecording(tabId);
      } else if (!shouldRecord) {
        if (!state.graceTimer) {
          log("condições perdidas durante gravação — iniciando grace timer", { tabId, graceMs: hybridStopGraceMs });
          state.graceTimer = setTimeoutFn(() => {
            if (state.recordingState === "RECORDING") {
              log("grace expirou — parando gravação", { tabId });
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
        log("URL mudou durante gravação — parando", { tabId, de: state.lastUrl, para: snapshot.url });
        stopHybridRecording(tabId);
      }
    }
    state.lastUrl = snapshot.url;
  }

  // The page finished the prepare sequence (rewound + settled + playing). Now
  // that no controls overlay is showing, begin the actual recording.
  function handleHybridPrepared(tabId) {
    const state = hybridStates.get(tabId);
    if (!state || !state.armed || state.recordingState !== "ARMED") {
      log("prepared recebido mas estado inválido — ignorado", {
        tabId,
        armed: state?.armed,
        recordingState: state?.recordingState,
      });
      return;
    }
    log("página preparada — iniciando gravação", { tabId });
    startHybridRecording(tabId);
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
      log("gravação iniciada", { tabId, streamId });
      sendRuntimeMessage({
        action: "offscreen_start_recording",
        tabId,
        streamId,
        streamUpload: true,
      });
    } catch (err) {
      log("falha ao iniciar gravação — voltando para ARMED", { tabId, erro: String(err) });
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
    // Reset the prepare guard. It was set true when THIS recording was prepared
    // and is never cleared elsewhere — without zeroing it here, re-arming for the
    // next episode (continuous capture) would see preparing=true and never start
    // the prepare→record sequence again, so only the first episode ever recorded.
    state.preparing = false;
    state.recordingState = "STOPPED";
    broadcastHybridStatus(tabId);

    if (state.mode === "dom") {
      sendTabMessage(tabId, { action: "hybrid_restore_dom" });
    }
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
    handleHybridPrepared,
    handleHybridVideoState,
    handleOffscreenError,
    handleOffscreenStarted,
    handleOffscreenStopped,
    startHybridRecording,
    stopHybridRecording,
  };
}

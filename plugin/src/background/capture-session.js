/* ========================================================================
 * KuraNAS Stream Grabber — Capture Session (episode-keyed state machine)
 *
 * Fase 2 da ingestão de mídia. Diferente do modo híbrido manual
 * (hybrid-state.js), esta máquina segue o player sozinho a partir do estado
 * normalizado que um adapter de página emite, e é chaveada por *episódio*:
 *
 *   - IDLE -> RECORDING(id) quando o vídeo toca em tela cheia e há episodeId;
 *   - fim do episódio (currentTime ~ duration) ou ended -> finaliza e volta a
 *     IDLE (sobrevive ao dono dormir: o player para no fim e a gravação para);
 *   - autoplay do próximo (episodeId novo) -> finaliza o atual e começa o novo
 *     (arquivo separado, sem ação manual);
 *   - mesmo episodeId reaparecendo (retomada após interrupção curta) -> continua
 *     a MESMA captura lógica via janela de graça, sem abrir arquivo novo. Para
 *     retomada após um stop completo, a idempotência por episode_key no backend
 *     (`/captures/upload/init`) garante um único arquivo.
 *
 * A máquina é pura/injetável para ser testável com `node --test`: toda a borda
 * (offscreen, upload, mensagens) chega por callbacks.
 * ======================================================================== */

const DEFAULT_END_EPSILON_SECONDS = 2;

export function createCaptureSessionMachine({
  captureSessions,
  startCapture,
  stopCapture,
  broadcastStatus = () => {},
  graceMs,
  endEpsilonSeconds = DEFAULT_END_EPSILON_SECONDS,
  setTimeoutFn = setTimeout,
  clearTimeoutFn = clearTimeout,
}) {
  function getState(tabId) {
    let state = captureSessions.get(tabId);
    if (!state) {
      state = {
        status: "IDLE",
        episodeId: null,
        episodeKey: null,
        title: null,
        graceTimer: null,
        archivedKeys: new Set(),
        lastSnapshot: null,
      };
      captureSessions.set(tabId, state);
    }
    return state;
  }

  function getSessionStatus(tabId) {
    const state = captureSessions.get(tabId);
    if (!state) return { state: "IDLE", episodeKey: null };
    return {
      state: state.status,
      episodeKey: state.episodeKey,
      title: state.title,
    };
  }

  function buildEpisodeKey(snapshot) {
    if (!snapshot || !snapshot.service || !snapshot.episodeId) return null;
    return `${snapshot.service}:${snapshot.episodeId}`;
  }

  function isPlayableSnapshot(snapshot) {
    return Boolean(
      snapshot &&
        snapshot.episodeId &&
        snapshot.isPlaying &&
        snapshot.isFullscreen &&
        snapshot.duration > 0
    );
  }

  function isAtEnd(snapshot) {
    if (!snapshot) return false;
    if (snapshot.isEnded) return true;
    if (!(snapshot.duration > 0)) return false;
    return snapshot.currentTime >= snapshot.duration - endEpsilonSeconds;
  }

  function clearGrace(state) {
    if (state.graceTimer) {
      clearTimeoutFn(state.graceTimer);
      state.graceTimer = null;
    }
  }

  function begin(tabId, state, snapshot) {
    const episodeKey = buildEpisodeKey(snapshot);
    if (!episodeKey || state.archivedKeys.has(episodeKey)) return;

    state.status = "RECORDING";
    state.episodeId = snapshot.episodeId;
    state.episodeKey = episodeKey;
    state.title = snapshot.title || snapshot.episodeId;
    clearGrace(state);
    broadcastStatus(tabId);

    Promise.resolve(
      startCapture(tabId, { episodeKey, title: state.title, snapshot })
    )
      .then((result) => {
        // The backend may report this episode is already fully archived; in that
        // case the capture never armed — drop back to IDLE and remember the key
        // so we don't hammer init on every frame at the end screen.
        if (result && result.recording === false) {
          state.archivedKeys.add(episodeKey);
          if (state.episodeKey === episodeKey && state.status === "RECORDING") {
            state.status = "IDLE";
            state.episodeId = null;
            state.episodeKey = null;
            broadcastStatus(tabId);
          }
        }
      })
      .catch(() => {
        if (state.episodeKey === episodeKey && state.status === "RECORDING") {
          state.status = "IDLE";
          state.episodeId = null;
          state.episodeKey = null;
          broadcastStatus(tabId);
        }
      });
  }

  function finalize(tabId, state, { archive }) {
    if (state.status !== "RECORDING") return;
    const episodeKey = state.episodeKey;
    clearGrace(state);
    state.status = "IDLE";
    state.episodeId = null;
    state.episodeKey = null;
    if (archive && episodeKey) state.archivedKeys.add(episodeKey);
    stopCapture(tabId, { episodeKey });
    broadcastStatus(tabId);
  }

  function handleEpisodeState(tabId, snapshot) {
    const state = getState(tabId);
    state.lastSnapshot = snapshot;

    const episodeKey = buildEpisodeKey(snapshot);

    if (state.status === "IDLE") {
      if (isPlayableSnapshot(snapshot) && !isAtEnd(snapshot)) {
        begin(tabId, state, snapshot);
      }
      return;
    }

    // RECORDING
    if (episodeKey && episodeKey !== state.episodeKey) {
      // Autoplay engaged the next episode -> close the current file and start a
      // fresh one for the new id.
      finalize(tabId, state, { archive: true });
      if (isPlayableSnapshot(snapshot) && !isAtEnd(snapshot)) {
        begin(tabId, state, snapshot);
      }
      return;
    }

    if (isAtEnd(snapshot)) {
      // Episode reached its end (or the owner fell asleep and the player ran
      // out) -> finalize without cutting earlier.
      finalize(tabId, state, { archive: true });
      return;
    }

    if (!isPlayableSnapshot(snapshot)) {
      // Short pause / left fullscreen: don't cut immediately. Open a grace
      // window; only finalize if it stays not-playable. Resuming inside the
      // window keeps the same logical capture (no new file).
      if (!state.graceTimer) {
        state.graceTimer = setTimeoutFn(() => {
          state.graceTimer = null;
          const latest = state.lastSnapshot;
          if (state.status === "RECORDING" && !isPlayableSnapshot(latest)) {
            finalize(tabId, state, { archive: false });
          }
        }, graceMs);
      }
      return;
    }

    // Still playing the same episode in fullscreen -> keep recording.
    clearGrace(state);
  }

  function cleanupTab(tabId) {
    const state = captureSessions.get(tabId);
    if (!state) return;
    clearGrace(state);
    if (state.status === "RECORDING") {
      stopCapture(tabId, { episodeKey: state.episodeKey });
    }
    captureSessions.delete(tabId);
  }

  return {
    buildEpisodeKey,
    cleanupTab,
    getSessionStatus,
    handleEpisodeState,
  };
}

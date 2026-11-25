import {
  HYBRID_UPLOAD_RECONNECT_GIVE_UP_MS,
  HYBRID_UPLOAD_RECONNECT_INITIAL_DELAY_MS,
  HYBRID_UPLOAD_RECONNECT_MAX_DELAY_MS,
} from "../shared/constants.js";
import { ChunkOffsetMismatchError } from "./uploader.js";

export const HYBRID_UPLOAD_STATUS = {
  UPLOADING: "uploading",
  RECONNECTING: "reconnecting",
  FAILED: "failed",
};

class ServerSessionLostError extends Error {
  constructor() {
    super("Server no longer has the open upload session");
  }
}

export function createHybridUploadSession({
  apiUrl,
  uploadID,
  startOffset = 0,
  buildInitBody,
  uploadChunkWithRetry,
  fetchImpl = fetch,
  wait,
  now = () => Date.now(),
  onStatusChange = () => {},
  reconnectInitialDelayMs = HYBRID_UPLOAD_RECONNECT_INITIAL_DELAY_MS,
  reconnectMaxDelayMs = HYBRID_UPLOAD_RECONNECT_MAX_DELAY_MS,
  reconnectGiveUpMs = HYBRID_UPLOAD_RECONNECT_GIVE_UP_MS,
}) {
  const bufferedChunks = [];
  let uploadedOffset = startOffset;
  let chunkIndex = 0;
  let status = HYBRID_UPLOAD_STATUS.UPLOADING;
  let failureMessage = "";
  let isDraining = false;
  let drainPromise = Promise.resolve();

  function setStatus(nextStatus, message = "") {
    status = nextStatus;
    failureMessage = message;
    onStatusChange({ status, error: failureMessage });
  }

  function fail(message) {
    bufferedChunks.length = 0;
    setStatus(HYBRID_UPLOAD_STATUS.FAILED, message);
  }

  function bufferedByteCount() {
    return bufferedChunks.reduce((total, chunk) => total + chunk.size, 0);
  }

  function alignWithServer(serverReceivedSize) {
    const bytesToDiscard = serverReceivedSize - uploadedOffset;
    if (bytesToDiscard < 0) {
      throw new Error(
        `Server holds ${serverReceivedSize} bytes but ${uploadedOffset} were already released from memory`
      );
    }
    if (bytesToDiscard > bufferedByteCount()) {
      throw new Error(
        `Server reports ${serverReceivedSize} bytes, more than this recording produced`
      );
    }

    let remainingToDiscard = bytesToDiscard;
    while (remainingToDiscard > 0) {
      const headChunk = bufferedChunks[0];
      if (headChunk.size <= remainingToDiscard) {
        bufferedChunks.shift();
        remainingToDiscard -= headChunk.size;
        chunkIndex += 1;
        continue;
      }
      bufferedChunks[0] = headChunk.slice(remainingToDiscard);
      remainingToDiscard = 0;
    }
    uploadedOffset = serverReceivedSize;
  }

  async function reopenServerSession() {
    const response = await fetchImpl(`${apiUrl}/captures/upload/init`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(buildInitBody()),
    });
    if (!response.ok) {
      throw new Error(`Init upload failed (${response.status})`);
    }
    const payload = await response.json();
    if (!payload.resumed || payload.upload_id !== uploadID) {
      throw new ServerSessionLostError();
    }
    return Number(payload.received_size || 0);
  }

  async function reconnect() {
    setStatus(HYBRID_UPLOAD_STATUS.RECONNECTING);
    const reconnectStartedAt = now();
    let delayMs = reconnectInitialDelayMs;

    while (now() - reconnectStartedAt < reconnectGiveUpMs) {
      await wait(delayMs);
      delayMs = Math.min(delayMs * 2, reconnectMaxDelayMs);
      let serverReceivedSize;
      try {
        serverReceivedSize = await reopenServerSession();
      } catch (error) {
        if (error instanceof ServerSessionLostError) {
          throw error;
        }
        continue;
      }
      alignWithServer(serverReceivedSize);
      setStatus(HYBRID_UPLOAD_STATUS.UPLOADING);
      return;
    }

    throw new Error("Connection to the server was not restored in time");
  }

  async function uploadHeadChunk() {
    const headChunk = bufferedChunks[0];
    await uploadChunkWithRetry(apiUrl, uploadID, headChunk, uploadedOffset, chunkIndex);
    bufferedChunks.shift();
    uploadedOffset += headChunk.size;
    chunkIndex += 1;
  }

  async function recoverFromChunkError(error) {
    const isRealignableMismatch =
      error instanceof ChunkOffsetMismatchError && error.receivedSize !== uploadedOffset;
    if (isRealignableMismatch) {
      alignWithServer(error.receivedSize);
      return;
    }
    await reconnect();
  }

  async function drainBufferedChunks() {
    try {
      while (bufferedChunks.length > 0 && status !== HYBRID_UPLOAD_STATUS.FAILED) {
        try {
          await uploadHeadChunk();
        } catch (error) {
          await recoverFromChunkError(error);
        }
      }
    } catch (error) {
      fail(error && error.message ? error.message : "Chunk upload failed");
    }
  }

  function startDraining() {
    if (isDraining) return;
    isDraining = true;
    drainPromise = drainBufferedChunks().finally(() => {
      isDraining = false;
      if (bufferedChunks.length > 0 && status !== HYBRID_UPLOAD_STATUS.FAILED) {
        startDraining();
      }
    });
  }

  function enqueueChunk(chunkBlob) {
    if (status === HYBRID_UPLOAD_STATUS.FAILED) return;
    bufferedChunks.push(chunkBlob);
    startDraining();
  }

  async function waitUntilDrained() {
    let observedPromise;
    do {
      observedPromise = drainPromise;
      await observedPromise;
    } while (observedPromise !== drainPromise);
  }

  return {
    apiUrl,
    uploadID,
    enqueueChunk,
    abort: fail,
    waitUntilDrained,
    getStatus: () => ({ status, error: failureMessage }),
    hasFailed: () => status === HYBRID_UPLOAD_STATUS.FAILED,
    getUploadedOffset: () => uploadedOffset,
    getChunkIndex: () => chunkIndex,
  };
}

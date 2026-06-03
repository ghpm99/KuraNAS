import {
  CAPTURE_UPLOAD_CHUNK_SIZE,
  CAPTURE_UPLOAD_MAX_RETRIES,
} from "../shared/constants.js";

export function createUploader({
  getApiBaseUrl,
  guessExtension,
  sanitizeFileName,
  waitFn,
  fetchImpl = fetch,
}) {
  const wait = waitFn || ((ms) => new Promise((resolve) => setTimeout(resolve, ms)));

  async function uploadToKuraNAS(blob, name, mediaType) {
    const ext = guessExtension(blob.type, mediaType);
    const fileName = `${sanitizeFileName(name)}.${ext}`;
    const apiUrl = await getApiBaseUrl();
    return uploadCaptureChunked(apiUrl, blob, {
      name,
      mediaType,
      mimeType: blob.type || "application/octet-stream",
      fileName,
    });
  }

  async function uploadCaptureChunked(apiUrl, blob, metadata) {
    const initResp = await fetchImpl(`${apiUrl}/captures/upload/init`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: metadata.name,
        media_type: metadata.mediaType,
        mime_type: metadata.mimeType,
        size: blob.size,
        file_name: metadata.fileName,
      }),
    });

    if (!initResp.ok) {
      const body = await initResp.text();
      throw new Error(`Init upload failed (${initResp.status}): ${body}`);
    }

    const initPayload = await initResp.json();
    const uploadID = initPayload.upload_id;
    const serverChunkSize = Number(initPayload.chunk_size || 0);
    const chunkSize = serverChunkSize > 0 ? serverChunkSize : CAPTURE_UPLOAD_CHUNK_SIZE;
    if (!uploadID) {
      throw new Error("Invalid chunked upload init response: upload_id is required");
    }

    let offset = 0;
    let chunkIndex = 0;

    while (offset < blob.size) {
      const chunkBlob = blob.slice(offset, offset + chunkSize);
      await uploadChunkWithRetry(apiUrl, uploadID, chunkBlob, offset, chunkIndex);
      offset += chunkBlob.size;
      chunkIndex += 1;
    }

    const completeResp = await fetchImpl(`${apiUrl}/captures/upload/complete`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ upload_id: uploadID }),
    });

    if (!completeResp.ok) {
      const body = await completeResp.text();
      throw new Error(`Complete upload failed (${completeResp.status}): ${body}`);
    }

    return completeResp.json();
  }

  async function uploadChunkWithRetry(apiUrl, uploadID, chunkBlob, offset, chunkIndex) {
    let lastError = null;

    for (let attempt = 1; attempt <= CAPTURE_UPLOAD_MAX_RETRIES; attempt++) {
      try {
        await uploadChunk(apiUrl, uploadID, chunkBlob, offset, chunkIndex);
        return;
      } catch (error) {
        lastError = error;
        if (attempt < CAPTURE_UPLOAD_MAX_RETRIES) {
          await wait(200 * attempt);
        }
      }
    }

    throw lastError || new Error("chunk upload failed");
  }

  async function uploadChunk(apiUrl, uploadID, chunkBlob, offset, chunkIndex) {
    const formData = new FormData();
    formData.append("upload_id", uploadID);
    formData.append("offset", String(offset));
    formData.append("chunk_index", String(chunkIndex));
    formData.append("chunk", chunkBlob, `chunk_${chunkIndex}`);

    const response = await fetchImpl(`${apiUrl}/captures/upload/chunk`, {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      const body = await response.text();
      throw new Error(`Chunk upload failed (${response.status}): ${body}`);
    }
  }

  async function uploadBlobCapture(tabId, blobUrl, name) {
    try {
      const resp = await fetchImpl(blobUrl);
      const blob = await resp.blob();
      const captureName = name || `blob_${Date.now()}`;
      await uploadToKuraNAS(blob, captureName, "blob");
      return { ok: true, name: captureName };
    } catch (error) {
      return { error: error.message };
    }
  }

  async function handleSaveRecordingBlob(msg) {
    try {
      const response = await fetchImpl(msg.blobUrl);
      const blob = await response.blob();

      const name = msg.name || `recording_${Date.now()}`;
      await uploadToKuraNAS(blob, name, "recording");
    } catch {
      // Blob URL may have been revoked
    }
  }

  return {
    handleSaveRecordingBlob,
    uploadBlobCapture,
    uploadChunkWithRetry,
    uploadToKuraNAS,
  };
}

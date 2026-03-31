/* ========================================================================
 * KuraNAS Stream Grabber — Blob Interceptor (MAIN world)
 * Monkey-patches browser APIs to capture blob/MediaSource data.
 * ======================================================================== */

(function () {
  "use strict";

  const capturedBlobs = new Map();

  // -----------------------------------------------------------------------
  // 1. URL.createObjectURL
  // -----------------------------------------------------------------------

  const originalCreateObjectURL = URL.createObjectURL;
  URL.createObjectURL = function (obj) {
    const url = originalCreateObjectURL.call(this, obj);

    if (obj instanceof Blob || obj instanceof MediaSource) {
      const info = {
        url,
        mimeType: obj instanceof Blob ? obj.type : "mediasource",
        size: obj instanceof Blob ? obj.size : 0,
        isMediaSource: obj instanceof MediaSource,
        timestamp: Date.now(),
      };

      capturedBlobs.set(url, info);

      document.dispatchEvent(
        new CustomEvent("__stream_grabber_blob__", {
          detail: {
            blobUrl: url,
            mimeType: info.mimeType,
            size: info.size,
            isMediaSource: info.isMediaSource,
          },
        })
      );
    }

    return url;
  };

  // -----------------------------------------------------------------------
  // 2. URL.revokeObjectURL — keep reference but allow browser revocation
  // -----------------------------------------------------------------------

  const originalRevokeObjectURL = URL.revokeObjectURL;
  URL.revokeObjectURL = function (url) {
    originalRevokeObjectURL.call(this, url);
  };

  // -----------------------------------------------------------------------
  // 3. MediaSource.prototype.addSourceBuffer — capture chunks
  // -----------------------------------------------------------------------

  if (typeof MediaSource !== "undefined") {
    const originalAddSourceBuffer = MediaSource.prototype.addSourceBuffer;
    MediaSource.prototype.addSourceBuffer = function (mimeType) {
      const sourceBuffer = originalAddSourceBuffer.call(this, mimeType);

      sourceBuffer.__sg_chunks = [];
      sourceBuffer.__sg_mimeType = mimeType;

      const originalAppendBuffer = sourceBuffer.appendBuffer;
      sourceBuffer.appendBuffer = function (data) {
        if (data && data.byteLength > 0) {
          const copy = data instanceof ArrayBuffer
            ? new Uint8Array(data).slice(0)
            : new Uint8Array(data.buffer, data.byteOffset, data.byteLength).slice(0);

          sourceBuffer.__sg_chunks.push(copy);

          let totalSize = 0;
          for (const chunk of sourceBuffer.__sg_chunks) {
            totalSize += chunk.byteLength;
          }

          document.dispatchEvent(
            new CustomEvent("__stream_grabber_chunk__", {
              detail: {
                mimeType: sourceBuffer.__sg_mimeType,
                chunkCount: sourceBuffer.__sg_chunks.length,
                totalSize,
              },
            })
          );
        }

        return originalAppendBuffer.call(this, data);
      };

      return sourceBuffer;
    };

    // -----------------------------------------------------------------------
    // 4. MediaSource Proxy — track instances
    // -----------------------------------------------------------------------

    window.__sg_mediaSources = [];

    const OriginalMediaSource = MediaSource;
    window.MediaSource = new Proxy(OriginalMediaSource, {
      construct(target, args, newTarget) {
        const instance = Reflect.construct(target, args, newTarget);
        window.__sg_mediaSources.push(instance);
        return instance;
      },
    });

    window.MediaSource.isTypeSupported =
      OriginalMediaSource.isTypeSupported.bind(OriginalMediaSource);
  }

  // -----------------------------------------------------------------------
  // 5. Download Captured Buffers
  // -----------------------------------------------------------------------

  window.addEventListener("__stream_grabber_download_request__", () => {
    try {
      const allChunks = [];
      let totalSize = 0;

      if (typeof MediaSource !== "undefined" && window.__sg_mediaSources) {
        for (const ms of window.__sg_mediaSources) {
          if (!ms.sourceBuffers) continue;
          for (let i = 0; i < ms.sourceBuffers.length; i++) {
            const sb = ms.sourceBuffers[i];
            if (sb.__sg_chunks) {
              for (const chunk of sb.__sg_chunks) {
                allChunks.push(chunk);
                totalSize += chunk.byteLength;
              }
            }
          }
        }
      }

      if (allChunks.length === 0) {
        document.dispatchEvent(
          new CustomEvent("__stream_grabber_download_response__", {
            detail: { error: "No captured data" },
          })
        );
        return;
      }

      const merged = new Uint8Array(totalSize);
      let offset = 0;
      for (const chunk of allChunks) {
        merged.set(chunk, offset);
        offset += chunk.byteLength;
      }

      const blob = new Blob([merged], { type: "video/webm" });
      const url = URL.createObjectURL.call(URL, blob);

      document.dispatchEvent(
        new CustomEvent("__stream_grabber_download_response__", {
          detail: { blobUrl: url, size: totalSize },
        })
      );
    } catch (err) {
      document.dispatchEvent(
        new CustomEvent("__stream_grabber_download_response__", {
          detail: { error: err.message },
        })
      );
    }
  });
})();

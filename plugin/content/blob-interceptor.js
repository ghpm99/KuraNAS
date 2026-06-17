/* ========================================================================
 * KuraNAS Stream Grabber — Blob Interceptor (MAIN world)
 *
 * Detects blob/MediaSource usage on the page so the popup can list it. It is
 * DETECTION-ONLY: it never retains the media bytes.
 *
 * A previous version deep-copied every segment appended to a MediaSource into an
 * unbounded array (sourceBuffer.__sg_chunks) to allow "downloading" the MSE
 * stream. That feeder was dead code (no message ever triggered the download), yet
 * it duplicated the entire audio+video stream in the page heap for the whole
 * playback — a steadily growing memory leak on every MSE site (Netflix,
 * Crunchyroll, YouTube, Prime…), visible during long recordings. The hybrid
 * tab-capture recorder owns the real bytes, so here we only emit lightweight
 * detection events (running count/size) and keep nothing.
 * ======================================================================== */

(function () {
  "use strict";

  // -----------------------------------------------------------------------
  // 1. URL.createObjectURL — announce blob/MediaSource creation (no retention)
  // -----------------------------------------------------------------------

  const originalCreateObjectURL = URL.createObjectURL;
  URL.createObjectURL = function (obj) {
    const url = originalCreateObjectURL.call(this, obj);

    if (obj instanceof Blob || obj instanceof MediaSource) {
      document.dispatchEvent(
        new CustomEvent("__stream_grabber_blob__", {
          detail: {
            blobUrl: url,
            mimeType: obj instanceof Blob ? obj.type : "mediasource",
            size: obj instanceof Blob ? obj.size : 0,
            isMediaSource: obj instanceof MediaSource,
          },
        })
      );
    }

    return url;
  };

  // -----------------------------------------------------------------------
  // 2. MediaSource.prototype.addSourceBuffer — COUNT appended media, keep nothing
  // -----------------------------------------------------------------------

  if (typeof MediaSource !== "undefined") {
    const originalAddSourceBuffer = MediaSource.prototype.addSourceBuffer;
    MediaSource.prototype.addSourceBuffer = function (mimeType) {
      const sourceBuffer = originalAddSourceBuffer.call(this, mimeType);

      let chunkCount = 0;
      let totalSize = 0;

      const originalAppendBuffer = sourceBuffer.appendBuffer;
      sourceBuffer.appendBuffer = function (data) {
        const byteLength = data ? data.byteLength || 0 : 0;
        if (byteLength > 0) {
          chunkCount += 1;
          totalSize += byteLength;
          document.dispatchEvent(
            new CustomEvent("__stream_grabber_chunk__", {
              detail: { mimeType, chunkCount, totalSize },
            })
          );
        }
        return originalAppendBuffer.call(this, data);
      };

      return sourceBuffer;
    };
  }
})();

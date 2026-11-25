(function () {
  "use strict";

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

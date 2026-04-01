export function resolveUrl(base, relative) {
  if (!relative) return base;
  try {
    return new URL(relative, base).href;
  } catch {
    const basePath = base.substring(0, base.lastIndexOf("/") + 1);
    return basePath + relative;
  }
}

export function sanitizeFileName(name) {
  return name
    .replace(/[<>:"/\\|?*\x00-\x1f]/g, "_")
    .replace(/\s+/g, "_")
    .substring(0, 200);
}

export function guessExtension(mimeType, mediaType) {
  if (mimeType) {
    if (mimeType.includes("mp2t")) return "ts";
    if (mimeType.includes("mp4")) return "mp4";
    if (mimeType.includes("webm")) return "webm";
    if (mimeType.includes("m4a") || mimeType.includes("x-m4a")) return "m4a";
    if (mimeType.includes("aac")) return "aac";
    if (mimeType.includes("mpeg") && mimeType.includes("audio")) return "mp3";
  }
  if (mediaType === "hls") return "ts";
  if (mediaType === "dash") return "mp4";
  return "bin";
}

export function wait(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

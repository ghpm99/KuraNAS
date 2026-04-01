export const DEFAULT_KURANAS_API_BASE = "http://localhost:8000/api/v1";
export const DISCOVERY_HEALTH_SUFFIX = "/health";
export const DISCOVERY_REQUEST_TIMEOUT_MS = 1500;
export const DISCOVERY_TAB_HOST_LIMIT = 10;

export const MEDIA_PATTERNS = [
  { regex: /\.m3u8(\?|$)/i, type: "hls" },
  { regex: /\.mpd(\?|$)/i, type: "dash" },
  { regex: /\.ts(\?|$)/i, type: "ts" },
  { regex: /\.mp4(\?|$)/i, type: "mp4" },
  { regex: /\.m4s(\?|$)/i, type: "m4s" },
  { regex: /\.aac(\?|$)/i, type: "aac" },
  { regex: /\.webm(\?|$)/i, type: "webm" },
];

export const MEDIA_CONTENT_TYPES = [
  { pattern: /mpegurl/i, type: "hls" },
  { pattern: /dash\+xml/i, type: "dash" },
  { pattern: /^video\//i, type: "video" },
  { pattern: /^audio\//i, type: "audio" },
];

export const HYBRID_STABILITY_MS = 200;
export const HYBRID_STOP_GRACE_MS = 5000;
export const CAPTURE_UPLOAD_CHUNK_SIZE = 2 * 1024 * 1024;
export const CAPTURE_UPLOAD_MAX_RETRIES = 3;

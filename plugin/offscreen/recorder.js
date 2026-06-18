/* ========================================================================
 * KuraNAS Stream Grabber — Offscreen Recorder
 * Invisible document that performs actual tab capture recording.
 *
 * Two encode paths, picked by capability:
 *   - WebCodecs (preferred): H.264 video + AAC/Opus audio muxed into a
 *     fragmented MP4. VideoEncoder runs with `hardwareAcceleration:
 *     "prefer-hardware"`, so the encode runs on the GPU when one is available
 *     — far cheaper on CPU than the libvpx VP9 software encoder, and it lands
 *     a universal MP4 straight in the library (no server-side transcode).
 *   - MediaRecorder/WebM (fallback): the original VP9 software path, used when
 *     WebCodecs/H.264 is unavailable.
 * The chosen container is reported to the service worker via
 * `offscreen_probe_format` BEFORE the upload session is created, so the server
 * names the file and stores the mime type correctly.
 * ======================================================================== */

import { Muxer, ArrayBufferTarget, StreamTarget } from "../vendor/mp4-muxer.mjs";

const recordings = new Map();

const LOG_PREFIX = "[KuraNAS rec]";

function logRec(tabId, ...args) {
  console.log(LOG_PREFIX, `tab=${tabId}`, ...args);
  // Mirror into the service worker console (the offscreen doc console is hidden).
  try {
    chrome.runtime
      .sendMessage({ action: "kuranas_log", source: LOG_PREFIX, args: [`tab=${tabId}`, ...args] })
      .catch(() => {});
  } catch (_e) {
    // sendMessage can throw if the SW is momentarily unavailable; the local
    // console.log above still happened.
  }
}

// describeTrack dumps everything that tells empty-frame-by-DRM apart from other
// causes: a Widevine-protected tab capture typically hands back a video track
// that is `muted: true` and/or reports 0x0 with no real frames, while a normal
// capture reports the page's dimensions and `muted: false`.
function describeTrack(track) {
  const settings = typeof track.getSettings === "function" ? track.getSettings() : {};
  return {
    kind: track.kind,
    label: track.label,
    enabled: track.enabled,
    muted: track.muted,
    readyState: track.readyState,
    width: settings.width,
    height: settings.height,
    frameRate: settings.frameRate,
  };
}

function logStreamDiagnostics(tabId, stream) {
  const videoTracks = stream.getVideoTracks();
  const audioTracks = stream.getAudioTracks();

  logRec(tabId, `stream obtido: ${videoTracks.length} vídeo, ${audioTracks.length} áudio`);

  videoTracks.forEach((track, i) => {
    const info = describeTrack(track);
    logRec(tabId, `vídeo[${i}]`, info);
    if (info.muted) {
      console.warn(
        LOG_PREFIX,
        `tab=${tabId}`,
        "track de vídeo está MUTED — forte indício de DRM/Widevine (frames não chegam)."
      );
    }
    // DRM/policy changes can mute the track AFTER capture starts; watch for it.
    track.onmute = () =>
      console.warn(LOG_PREFIX, `tab=${tabId}`, "vídeo MUTED durante a gravação (DRM?).");
    track.onunmute = () => logRec(tabId, "vídeo UNMUTED (frames voltaram).");
    track.onended = () => logRec(tabId, "track de vídeo ENDED.");
  });

  audioTracks.forEach((track, i) => logRec(tabId, `áudio[${i}]`, describeTrack(track)));
}

// setupAudioPassthrough re-plays the captured tab audio through the speakers.
// tabCapture redirects the tab's audio into the stream, so the owner stops
// hearing it; piping the stream's audio to the AudioContext destination gives
// the sound back while the recording still captures it. Returns the context so
// cleanup can close it (returns null when there is no audio track / on failure).
function setupAudioPassthrough(tabId, stream) {
  try {
    if (stream.getAudioTracks().length === 0) return null;
    const audioContext = new AudioContext();
    const source = audioContext.createMediaStreamSource(stream);
    source.connect(audioContext.destination);
    if (audioContext.state === "suspended") {
      audioContext.resume().catch(() => {});
    }
    logRec(tabId, "áudio reencaminhado aos alto-falantes (você ouve enquanto grava)");
    return audioContext;
  } catch (e) {
    logRec(tabId, "falha ao reencaminhar áudio:", e && e.message);
    return null;
  }
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.action === "offscreen_probe_format") {
    // Report the container we will record in so the service worker can name the
    // upload file and set its mime type BEFORE the recording starts. Deterministic
    // and cached, so it matches what startRecording later produces.
    chooseRecordingFormat()
      .then((fmt) => sendResponse({ mimeType: fmt.mimeType, fileExt: fmt.fileExt, container: fmt.container }))
      .catch(() => sendResponse({ mimeType: "video/webm", fileExt: "webm", container: "webm" }));
    return true; // keep the message channel open for the async sendResponse
  }
  if (msg.action === "offscreen_start_recording") {
    startRecording(msg.tabId, msg.streamId, Boolean(msg.streamUpload));
  }
  if (msg.action === "offscreen_stop_recording") {
    stopRecording(msg.tabId);
  }
  if (msg.action === "offscreen_revoke_url" && msg.url) {
    // The service worker finished reading a streamed chunk's blob URL. Revoke it
    // here (the doc that created it) since URL.revokeObjectURL is unavailable in
    // the service worker.
    try {
      URL.revokeObjectURL(msg.url);
    } catch (_e) {
      // best effort
    }
  }
  return undefined;
});

async function startRecording(tabId, streamId, streamUpload) {
  let stream;
  try {
    stream = await navigator.mediaDevices.getUserMedia({
      audio: {
        mandatory: {
          chromeMediaSource: "tab",
          chromeMediaSourceId: streamId,
        },
      },
      video: {
        mandatory: {
          chromeMediaSource: "tab",
          chromeMediaSourceId: streamId,
          maxWidth: 3840,
          maxHeight: 2160,
          maxFrameRate: 60,
        },
      },
    });
  } catch (err) {
    // getUserMedia/tabCapture failed outright (permission, invalid streamId, or
    // the capture was denied for the protected content).
    console.error(LOG_PREFIX, `tab=${tabId}`, "falha ao iniciar captura:", err.message);
    chrome.runtime.sendMessage({
      action: "hybrid_offscreen_error",
      tabId,
      error: err.message,
    });
    return;
  }

  logStreamDiagnostics(tabId, stream);

  const fmt = await chooseRecordingFormat();
  if (fmt.container === "mp4") {
    await startWebCodecsRecording(tabId, stream, streamUpload, fmt);
  } else {
    startMediaRecorderRecording(tabId, stream, streamUpload, fmt);
  }
}

// ---------------------------------------------------------------------------
// WebCodecs (GPU) path — H.264 + AAC/Opus → fragmented MP4
// ---------------------------------------------------------------------------

async function startWebCodecsRecording(tabId, stream, streamUpload, fmt) {
  const startedAt = Date.now();
  const videoTrack = stream.getVideoTracks()[0];
  const audioTrack = stream.getAudioTracks()[0] || null;

  const vs = videoTrack && typeof videoTrack.getSettings === "function" ? videoTrack.getSettings() : {};
  const width = vs.width || 1920;
  const height = vs.height || 1080;
  const frameRate = Math.min(vs.frameRate || 60, 60);
  const { videoBitsPerSecond, audioBitsPerSecond } = selectBitrates(stream);

  // Audio passthrough drives the speakers from the captured stream; the encoder
  // reads a CLONE of the audio track so cancelling the encoder's reader on stop
  // does not cut the passthrough (both clones pull from the same source).
  const audioContext = setupAudioPassthrough(tabId, stream);
  const audioForEncode =
    audioTrack && fmt.audio && typeof audioTrack.clone === "function" ? audioTrack.clone() : null;
  const audioChannels = vs.channelCount || (audioTrack && audioTrack.getSettings && audioTrack.getSettings().channelCount) || 2;
  const audioSampleRate = (audioTrack && audioTrack.getSettings && audioTrack.getSettings().sampleRate) || 48000;

  logRec(
    tabId,
    `WebCodecs(GPU) codec=${fmt.videoCodec}, áudio=${fmt.audio ? fmt.audio.codec : "nenhum"}, ` +
      `${width}x${height}@${frameRate}, vídeo=${Math.round(videoBitsPerSecond / 1e6)}Mbps, streamUpload=${streamUpload}`
  );

  let streamedBytes = 0;
  let target;
  if (streamUpload) {
    // Fragmented MP4 written sequentially: every onData buffer, concatenated in
    // arrival order, is a valid growing MP4. Forward each as a streamed chunk via
    // a blob: URL (a Blob/Uint8Array cannot cross chrome.runtime.sendMessage; the
    // SW fetches the URL to recover the bytes — same trick the WebM path uses).
    target = new StreamTarget({
      onData: (data, _position) => {
        const copy = new Uint8Array(data); // copy: the muxer may reuse the buffer
        streamedBytes += copy.byteLength;
        const chunkUrl = URL.createObjectURL(new Blob([copy], { type: fmt.mimeType }));
        chrome.runtime
          .sendMessage({ action: "hybrid_recording_chunk", tabId, chunkUrl, size: copy.byteLength })
          .catch(() => {});
      },
      chunked: true,
    });
  } else {
    target = new ArrayBufferTarget();
  }

  const muxer = new Muxer({
    target,
    video: { codec: fmt.muxerVideoCodec, width, height, frameRate },
    audio: audioForEncode ? { codec: fmt.audio.muxerCodec, numberOfChannels: audioChannels, sampleRate: audioSampleRate } : undefined,
    fastStart: streamUpload ? "fragmented" : "in-memory",
    firstTimestampBehavior: "offset",
  });

  const rec = {
    type: "webcodecs",
    stream,
    audioContext,
    audioForEncode,
    streamUpload,
    fmt,
    startedAt,
    muxer,
    target,
    videoChunks: 0,
    stopped: false,
    finished: false,
  };
  recordings.set(tabId, rec);

  const onEncoderError = (where) => (e) => {
    const message = e && e.message ? e.message : `${where} error`;
    console.error(LOG_PREFIX, `tab=${tabId}`, `${where} ERROR:`, message);
    if (rec.finished) return;
    rec.finished = true;
    chrome.runtime.sendMessage({ action: "hybrid_offscreen_error", tabId, error: message });
    cleanupRecording(tabId);
  };

  // Video encoder (GPU-preferred).
  const videoEncoder = new VideoEncoder({
    output: (chunk, meta) => {
      rec.videoChunks += 1;
      try {
        muxer.addVideoChunk(chunk, meta);
      } catch (e) {
        onEncoderError("muxer.addVideoChunk")(e);
      }
    },
    error: onEncoderError("VideoEncoder"),
  });
  videoEncoder.configure({
    codec: fmt.videoCodec,
    width,
    height,
    bitrate: videoBitsPerSecond,
    framerate: frameRate,
    hardwareAcceleration: "prefer-hardware",
    latencyMode: "quality",
  });
  rec.videoEncoder = videoEncoder;

  // Audio encoder (optional).
  if (audioForEncode) {
    const audioEncoder = new AudioEncoder({
      output: (chunk, meta) => {
        try {
          muxer.addAudioChunk(chunk, meta);
        } catch (e) {
          onEncoderError("muxer.addAudioChunk")(e);
        }
      },
      error: onEncoderError("AudioEncoder"),
    });
    audioEncoder.configure({
      codec: fmt.audio.codec,
      sampleRate: audioSampleRate,
      numberOfChannels: audioChannels,
      bitrate: audioBitsPerSecond,
    });
    rec.audioEncoder = audioEncoder;
  }

  // Pump frames from the tracks into the encoders. MediaStreamTrackProcessor
  // exposes each track as a stream of VideoFrame/AudioData we feed and close.
  const keyFrameEvery = Math.max(1, Math.round(frameRate * 2)); // keyframe ~every 2s
  const videoReader = new MediaStreamTrackProcessor({ track: videoTrack }).readable.getReader();
  rec.videoReader = videoReader;
  let frameIndex = 0;
  rec.videoPump = (async () => {
    while (!rec.stopped) {
      const { value: frame, done } = await videoReader.read();
      if (done) break;
      if (!frame) continue;
      try {
        if (videoEncoder.state === "configured" && videoEncoder.encodeQueueSize <= 30) {
          videoEncoder.encode(frame, { keyFrame: frameIndex % keyFrameEvery === 0 });
          frameIndex += 1;
        }
      } catch (_e) {
        // encoder may have closed mid-stop; the frame is closed below regardless
      } finally {
        frame.close();
      }
    }
  })().catch(() => {});

  if (audioForEncode && rec.audioEncoder) {
    const audioReader = new MediaStreamTrackProcessor({ track: audioForEncode }).readable.getReader();
    rec.audioReader = audioReader;
    rec.audioPump = (async () => {
      while (!rec.stopped) {
        const { value: data, done } = await audioReader.read();
        if (done) break;
        if (!data) continue;
        try {
          if (rec.audioEncoder.state === "configured" && rec.audioEncoder.encodeQueueSize <= 30) {
            rec.audioEncoder.encode(data);
          }
        } catch (_e) {
          // ditto
        } finally {
          data.close();
        }
      }
    })().catch(() => {});
  }

  logRec(tabId, "WebCodecs recorder iniciado");
  chrome.runtime.sendMessage({ action: "hybrid_offscreen_started", tabId });
}

async function finishWebCodecsRecording(tabId, rec) {
  if (rec.finished) return;
  rec.finished = true;

  try {
    await rec.videoPump;
  } catch (_e) {
    // pump already settled
  }
  if (rec.audioPump) {
    try {
      await rec.audioPump;
    } catch (_e) {
      // ditto
    }
  }

  try {
    if (rec.videoEncoder && rec.videoEncoder.state === "configured") await rec.videoEncoder.flush();
  } catch (_e) {
    // flush can reject if the encoder errored; the empty-frame check below still runs
  }
  try {
    if (rec.audioEncoder && rec.audioEncoder.state === "configured") await rec.audioEncoder.flush();
  } catch (_e) {
    // ditto
  }

  let finalizeOk = true;
  try {
    rec.muxer.finalize();
  } catch (e) {
    finalizeOk = false;
    console.error(LOG_PREFIX, `tab=${tabId}`, "muxer.finalize ERROR:", e && e.message);
  }

  const elapsedMs = Date.now() - rec.startedAt;
  logRec(tabId, `WebCodecs STOP: ${elapsedMs}ms, ${rec.videoChunks} chunks de vídeo encodados`);

  if (rec.videoChunks === 0 || !finalizeOk) {
    // No frames reached the encoder — typical of DRM-protected playback (Widevine)
    // or a recording stopped before the first frame. Fail loudly instead of
    // shipping an unplayable file.
    console.warn(
      LOG_PREFIX,
      `tab=${tabId}`,
      `Gravação vazia/inválida: ${rec.videoChunks} chunks, ${elapsedMs}ms.`
    );
    chrome.runtime.sendMessage({
      action: "hybrid_offscreen_error",
      tabId,
      error:
        "Gravação vazia: o stream não produziu dados (conteúdo protegido por DRM ou gravação muito curta).",
    });
  } else if (rec.streamUpload) {
    chrome.runtime.sendMessage({ action: "hybrid_recording_complete", tabId });
  } else {
    const blob = new Blob([rec.target.buffer], { type: rec.fmt.mimeType });
    const url = URL.createObjectURL(blob);
    chrome.runtime.sendMessage({
      action: "hybrid_save_recording_blob",
      tabId,
      blobUrl: url,
      name: `recording_${tabId}_${Date.now()}`,
    });
  }

  chrome.runtime.sendMessage({ action: "hybrid_offscreen_stopped", tabId });
  cleanupRecording(tabId);
}

// ---------------------------------------------------------------------------
// MediaRecorder (fallback) path — VP9 → WebM, original software encode
// ---------------------------------------------------------------------------

function startMediaRecorderRecording(tabId, stream, streamUpload, fmt) {
  const audioContext = setupAudioPassthrough(tabId, stream);

  const mimeType = fmt.mimeType;
  const { videoBitsPerSecond, audioBitsPerSecond } = selectBitrates(stream);
  logRec(
    tabId,
    `MediaRecorder mimeType=${mimeType}, vídeo=${Math.round(videoBitsPerSecond / 1e6)}Mbps, ` +
      `áudio=${Math.round(audioBitsPerSecond / 1e3)}kbps, streamUpload=${streamUpload}`
  );
  const recorder = new MediaRecorder(stream, {
    mimeType,
    videoBitsPerSecond,
    audioBitsPerSecond,
  });
  const chunks = [];

  let dataEvents = 0;
  let emptyDataEvents = 0;
  let streamedBytes = 0;
  const startedAt = Date.now();

  recorder.onstart = () => logRec(tabId, "recorder START");

  recorder.ondataavailable = (e) => {
    dataEvents += 1;
    if (e.data.size === 0) {
      emptyDataEvents += 1;
      logRec(
        tabId,
        `ondataavailable #${dataEvents}: 0 byte (vazio ${emptyDataEvents}x) — sem frames (DRM?)`
      );
      return;
    }

    logRec(tabId, `ondataavailable #${dataEvents}: ${e.data.size} bytes`);

    if (streamUpload) {
      // A Blob does NOT survive chrome.runtime.sendMessage (JSON serialization
      // turns it into {}), which silently dropped every streamed chunk. Pass a
      // blob: URL string instead — the service worker fetches it to get the
      // bytes. Same proven trick the non-stream path uses for the final blob.
      streamedBytes += e.data.size;
      const chunkUrl = URL.createObjectURL(e.data);
      chrome.runtime
        .sendMessage({
          action: "hybrid_recording_chunk",
          tabId,
          chunkUrl,
          size: e.data.size,
        })
        .catch(() => {});
      return;
    }

    chunks.push(e.data);
  };

  recorder.onstop = () => {
    const elapsedMs = Date.now() - startedAt;
    const totalBytes = chunks.reduce((sum, chunk) => sum + chunk.size, 0);
    logRec(
      tabId,
      `recorder STOP: ${elapsedMs}ms, ${dataEvents} eventos (${emptyDataEvents} vazios), ` +
        (streamUpload
          ? `${streamedBytes} bytes transmitidos (stream)`
          : `${totalBytes} bytes acumulados`)
    );

    if (streamUpload) {
      chrome.runtime
        .sendMessage({
          action: "hybrid_recording_complete",
          tabId,
        })
        .catch(() => {});
    } else {
      if (totalBytes === 0) {
        // The stream produced no media data, so the recording is empty. This is
        // typical of DRM-protected playback (e.g. Widevine on Crunchyroll) or a
        // recording stopped before any frame was captured. Fail loudly here
        // instead of creating an empty blob that the backend would reject.
        // The diagnostics above (track muted?, dataEvents, elapsedMs) tell which.
        const likelyDrm = dataEvents === 0 || emptyDataEvents === dataEvents;
        console.warn(
          LOG_PREFIX,
          `tab=${tabId}`,
          `Gravação vazia: ${dataEvents} eventos, ${emptyDataEvents} vazios, ${elapsedMs}ms. ` +
            (likelyDrm
              ? "Nenhum frame chegou — provável DRM/Widevine."
              : "Houve dados mas o total é 0 — verifique acima.")
        );
        chrome.runtime.sendMessage({
          action: "hybrid_offscreen_error",
          tabId,
          error:
            "Gravação vazia: o stream não produziu dados (conteúdo protegido por DRM ou gravação muito curta).",
        });
      } else {
        const blob = new Blob(chunks, { type: mimeType });
        const url = URL.createObjectURL(blob);

        chrome.runtime.sendMessage({
          action: "hybrid_save_recording_blob",
          tabId,
          blobUrl: url,
          name: `recording_${tabId}_${Date.now()}`,
        });
      }
    }

    chrome.runtime.sendMessage({
      action: "hybrid_offscreen_stopped",
      tabId,
    });

    cleanupRecording(tabId);
  };

  recorder.onerror = (e) => {
    const message = e.error ? e.error.message : "Unknown recorder error";
    console.error(LOG_PREFIX, `tab=${tabId}`, "recorder ERROR:", message);
    chrome.runtime.sendMessage({
      action: "hybrid_offscreen_error",
      tabId,
      error: message,
    });
    cleanupRecording(tabId);
  };

  recordings.set(tabId, { type: "mediarecorder", recorder, stream, url: null, audioContext });
  recorder.start(1000);
  logRec(tabId, "recorder.start(1000) chamado");

  chrome.runtime.sendMessage({
    action: "hybrid_offscreen_started",
    tabId,
  });
}

function stopRecording(tabId) {
  const rec = recordings.get(tabId);
  if (!rec) return;

  if (rec.type === "webcodecs") {
    if (rec.stopped) return;
    rec.stopped = true;
    if (rec.videoReader) rec.videoReader.cancel().catch(() => {});
    if (rec.audioReader) rec.audioReader.cancel().catch(() => {});
    finishWebCodecsRecording(tabId, rec);
    return;
  }

  if (rec.recorder && rec.recorder.state !== "inactive") {
    rec.recorder.stop();
  }
}

function cleanupRecording(tabId) {
  const rec = recordings.get(tabId);
  if (!rec) return;
  if (rec.stream) {
    rec.stream.getTracks().forEach((track) => track.stop());
  }
  if (rec.audioForEncode) {
    try {
      rec.audioForEncode.stop();
    } catch (_e) {
      // best effort
    }
  }
  if (rec.videoEncoder && rec.videoEncoder.state !== "closed") {
    try {
      rec.videoEncoder.close();
    } catch (_e) {
      // best effort
    }
  }
  if (rec.audioEncoder && rec.audioEncoder.state !== "closed") {
    try {
      rec.audioEncoder.close();
    } catch (_e) {
      // best effort
    }
  }
  if (rec.audioContext) {
    rec.audioContext.close().catch(() => {});
  }
  if (rec.url) {
    URL.revokeObjectURL(rec.url);
  }
  recordings.delete(tabId);
}

// ---------------------------------------------------------------------------
// Quality target. The design goal is to save the capture as close to the frames
// the browser actually rendered as practical — the LAN (2.5 Gbps) and the server
// favor quality over size, and the recording machine only records (all heavy
// processing is async on the server). So we provision a HIGH bitrate sized to the
// real captured resolution/framerate via a bits-per-pixel-per-frame budget, with
// a ceiling. The codecs (VP9 / H.264) are VBR: complex scenes spend up to this
// ceiling while calm scenes stay small, so over-provisioning costs little.
//
// Tuned for 1080p (the primary capture resolution): 0.20 bpp puts 1080p60 at
// ~25 Mbps (transparent), with the budget scaling up for the 3440x1440
// ultrawide and down for the 1366x768 panel. The ceiling drops the unused 4K
// headroom while still covering the ultrawide at 60 fps.
const QUALITY_BITS_PER_PIXEL = 0.2; // ~visually lossless 1080p for screen capture
const MIN_VIDEO_BITS_PER_SECOND = 12000000; // 12 Mbps floor (low-res panels stay crisp)
const MAX_VIDEO_BITS_PER_SECOND = 60000000; // 60 Mbps ceiling (covers 3440x1440@60)
const AUDIO_BITS_PER_SECOND = 320000; // 320 kbps (transparent)

// selectBitrates derives the encoder bitrate from the ACTUAL captured track
// (its rendered width/height/frameRate), so 4K60 gets far more than 1080p30.
function selectBitrates(stream) {
  const track = stream.getVideoTracks()[0];
  const settings =
    track && typeof track.getSettings === "function" ? track.getSettings() : {};
  const width = settings.width || 1920;
  const height = settings.height || 1080;
  const frameRate = Math.min(settings.frameRate || 60, 60);

  const target = Math.round(width * height * frameRate * QUALITY_BITS_PER_PIXEL);
  const videoBitsPerSecond = Math.max(
    MIN_VIDEO_BITS_PER_SECOND,
    Math.min(target, MAX_VIDEO_BITS_PER_SECOND)
  );
  return { videoBitsPerSecond, audioBitsPerSecond: AUDIO_BITS_PER_SECOND };
}

// ---------------------------------------------------------------------------
// Format selection — prefer a GPU-encoded MP4, fall back to MediaRecorder WebM.
// The decision is deterministic and cached so the probe (which sizes the upload
// session) matches what startRecording later produces.
// ---------------------------------------------------------------------------

let cachedFormat = null;

// H.264 codec strings ordered widest-coverage first. The exact level only needs
// to pass isConfigSupported; the real avcC (and its level) comes from the
// encoder's decoderConfig and is written by the muxer, so the precise string
// here is a capability probe, not the final stored profile.
//   avc1.640033 = High @ L5.1 (covers 4K)
//   avc1.64002A = High @ L4.2 (1080p60)
//   avc1.640028 = High @ L4.0 (1080p30)
//   avc1.42E01F = Baseline @ L3.1 (broadest)
const H264_CODECS = ["avc1.640033", "avc1.64002A", "avc1.640028", "avc1.42E01F"];

async function chooseRecordingFormat() {
  if (cachedFormat) return cachedFormat;
  cachedFormat = await detectRecordingFormat();
  return cachedFormat;
}

async function detectRecordingFormat() {
  const webCodecsReady =
    typeof VideoEncoder !== "undefined" &&
    typeof AudioEncoder !== "undefined" &&
    typeof MediaStreamTrackProcessor !== "undefined";

  if (webCodecsReady) {
    const videoCodec = await firstSupportedH264();
    if (videoCodec) {
      const audio = await firstSupportedAudio();
      return {
        container: "mp4",
        mimeType: "video/mp4",
        fileExt: "mp4",
        videoCodec, // WebCodecs codec string (e.g. avc1.640028)
        muxerVideoCodec: "avc",
        audio, // { codec, muxerCodec } | null
      };
    }
  }

  return {
    container: "webm",
    mimeType: selectWebmMimeType(),
    fileExt: "webm",
  };
}

async function firstSupportedH264() {
  for (const codec of H264_CODECS) {
    try {
      const res = await VideoEncoder.isConfigSupported({
        codec,
        width: 1920,
        height: 1080,
        bitrate: MIN_VIDEO_BITS_PER_SECOND,
        framerate: 60,
        hardwareAcceleration: "prefer-hardware",
      });
      if (res && res.supported) return codec;
    } catch (_e) {
      // unsupported codec string throws on some builds; try the next
    }
  }
  return null;
}

async function firstSupportedAudio() {
  const candidates = [
    { codec: "mp4a.40.2", muxerCodec: "aac" }, // AAC-LC — most universal in MP4
    { codec: "opus", muxerCodec: "opus" },
  ];
  for (const candidate of candidates) {
    try {
      const res = await AudioEncoder.isConfigSupported({
        codec: candidate.codec,
        sampleRate: 48000,
        numberOfChannels: 2,
        bitrate: AUDIO_BITS_PER_SECOND,
      });
      if (res && res.supported) return candidate;
    } catch (_e) {
      // try the next
    }
  }
  return null; // video-only MP4 if no audio codec is available
}

function selectWebmMimeType() {
  const preferred = [
    "video/webm;codecs=vp9,opus",
    "video/webm;codecs=vp8,opus",
    "video/webm",
  ];

  for (const type of preferred) {
    if (MediaRecorder.isTypeSupported(type)) return type;
  }

  return "video/webm";
}

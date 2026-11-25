import { Muxer, ArrayBufferTarget, StreamTarget } from "../vendor/mp4-muxer.mjs";

const recordings = new Map();

const LOG_PREFIX = "[KuraNAS rec]";

function logRec(tabId, ...args) {
  console.log(LOG_PREFIX, `tab=${tabId}`, ...args);
  try {
    chrome.runtime
      .sendMessage({ action: "kuranas_log", source: LOG_PREFIX, args: [`tab=${tabId}`, ...args] })
      .catch(() => {});
  } catch (_e) {
  }
}

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
    track.onmute = () =>
      console.warn(LOG_PREFIX, `tab=${tabId}`, "vídeo MUTED durante a gravação (DRM?).");
    track.onunmute = () => logRec(tabId, "vídeo UNMUTED (frames voltaram).");
    track.onended = () => logRec(tabId, "track de vídeo ENDED.");
  });

  audioTracks.forEach((track, i) => logRec(tabId, `áudio[${i}]`, describeTrack(track)));
}

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
    chooseRecordingFormat()
      .then((fmt) => sendResponse({ mimeType: fmt.mimeType, fileExt: fmt.fileExt, container: fmt.container }))
      .catch(() => sendResponse({ mimeType: "video/webm", fileExt: "webm", container: "webm" }));
    return true;
  }
  if (msg.action === "offscreen_start_recording") {
    startRecording(msg.tabId, msg.streamId, Boolean(msg.streamUpload));
  }
  if (msg.action === "offscreen_stop_recording") {
    stopRecording(msg.tabId);
  }
  if (msg.action === "offscreen_revoke_url" && msg.url) {
    try {
      URL.revokeObjectURL(msg.url);
    } catch (_e) {
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

async function startWebCodecsRecording(tabId, stream, streamUpload, fmt) {
  const startedAt = Date.now();
  const videoTrack = stream.getVideoTracks()[0];
  const audioTrack = stream.getAudioTracks()[0] || null;

  const vs = videoTrack && typeof videoTrack.getSettings === "function" ? videoTrack.getSettings() : {};
  const width = vs.width || 1920;
  const height = vs.height || 1080;
  const frameRate = Math.min(vs.frameRate || 60, 60);
  const { videoBitsPerSecond, audioBitsPerSecond } = selectBitrates(stream);

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
    target = new StreamTarget({
      onData: (data, _position) => {
        const copy = new Uint8Array(data);
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

  const keyFrameEvery = Math.max(1, Math.round(frameRate * 2));
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
  }
  if (rec.audioPump) {
    try {
      await rec.audioPump;
    } catch (_e) {
    }
  }

  try {
    if (rec.videoEncoder && rec.videoEncoder.state === "configured") await rec.videoEncoder.flush();
  } catch (_e) {
  }
  try {
    if (rec.audioEncoder && rec.audioEncoder.state === "configured") await rec.audioEncoder.flush();
  } catch (_e) {
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
    }
  }
  if (rec.videoEncoder && rec.videoEncoder.state !== "closed") {
    try {
      rec.videoEncoder.close();
    } catch (_e) {
    }
  }
  if (rec.audioEncoder && rec.audioEncoder.state !== "closed") {
    try {
      rec.audioEncoder.close();
    } catch (_e) {
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

const QUALITY_BITS_PER_PIXEL = 0.2;
const MIN_VIDEO_BITS_PER_SECOND = 12000000;
const MAX_VIDEO_BITS_PER_SECOND = 60000000;
const AUDIO_BITS_PER_SECOND = 320000;

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

let cachedFormat = null;

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
        videoCodec,
        muxerVideoCodec: "avc",
        audio,
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
    }
  }
  return null;
}

async function firstSupportedAudio() {
  const candidates = [
    { codec: "mp4a.40.2", muxerCodec: "aac" },
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
    }
  }
  return null;
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

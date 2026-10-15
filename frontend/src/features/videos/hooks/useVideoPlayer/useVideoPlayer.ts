import {
    nextVideoPlayback,
    previousVideoPlayback,
    startVideoPlayback,
    updateVideoPlaybackState,
    type UpdateVideoPlaybackStateRequest,
    VideoPlaybackSessionDto,
} from '@/service/videoPlayback';
import {
    buildDirectVideoStreamUrl,
    buildRemuxVideoStreamUrl,
    buildTranscodeVideoStreamUrl,
    isContainerUnplayableByBrowser,
} from '@/features/videos/videoPlayer/videoStreamSource';
import {
    getPlaybackErrorKindFromPlayRejection,
    type PlaybackErrorKind,
} from '@/features/videos/videoPlayer/playbackError';
import { flushVideoPlaybackState } from '@/service/playerStateFlush';
import { hasReachedCompletionThreshold } from './playbackCompletion';
import {
    DEFAULT_UNMUTE_VOLUME,
    loadVideoPlaybackPreferences,
    saveVideoPlaybackPreferences,
    type VideoPlaybackPreferences,
} from './videoPlaybackPreferences';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

type Status = 'waiting' | 'playing' | 'paused' | 'stopped';

type StreamMode = 'direct' | 'remux' | 'transcode';

const VOLUME_PERSIST_DEBOUNCE_MS = 300;

const TRANSCODE_FALLBACK_ERROR_KINDS: PlaybackErrorKind[] = ['unsupported', 'decode'];

const findFallbackStreamMode = (
    currentStreamMode: StreamMode,
    errorKind: PlaybackErrorKind
): StreamMode | null => {
    if (currentStreamMode === 'direct' && errorKind === 'unsupported') return 'remux';
    if (currentStreamMode === 'remux' && TRANSCODE_FALLBACK_ERROR_KINDS.includes(errorKind)) {
        return 'transcode';
    }
    return null;
};

const findVideoFormat = (session: VideoPlaybackSessionDto | null, videoId: number) =>
    session?.playlist.items.find((item) => item.video.id === videoId)?.video.format;

const useVideoPlayer = ({
    videoId,
    playlistId,
    persistProgress = true,
}: {
    videoId: string;
    playlistId?: number | null;
    persistProgress?: boolean;
}) => {
    const [status, setStatus] = useState<Status>('waiting');
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);
    const [initialPreferences] = useState(loadVideoPlaybackPreferences);
    const [volume, setVolume] = useState(initialPreferences.volume);
    const [isMuted, setIsMuted] = useState(initialPreferences.isMuted);
    const [playbackRate, setPlaybackRate] = useState(initialPreferences.playbackRate);
    const [playbackError, setPlaybackError] = useState<PlaybackErrorKind | null>(null);
    const [session, setSession] = useState<VideoPlaybackSessionDto | null>(null);
    const videoRef = useRef<HTMLVideoElement>(null);
    const remuxOffsetSecondsRef = useRef<number | null>(null);
    const streamModeRef = useRef<StreamMode>('direct');
    const reloadWithStreamModeRef = useRef<
        (videoToPlayId: number, seekSeconds: number, streamMode: StreamMode) => void
    >(() => undefined);
    const syncTimerRef = useRef<number | null>(null);
    const preferencesRef = useRef<VideoPlaybackPreferences>(initialPreferences);
    const volumePersistTimerRef = useRef<number | null>(null);
    const latestPlaybackRef = useRef({
        currentTime: 0,
        duration: 0,
        status: 'waiting' as Status,
        session: null as VideoPlaybackSessionDto | null,
    });

    useEffect(() => {
        latestPlaybackRef.current = {
            currentTime,
            duration,
            status,
            session,
        };
    }, [currentTime, duration, status, session]);

    const setVideoStarred = useCallback((videoId: number, isStarred: boolean) => {
        setSession((currentSession) =>
            currentSession
                ? {
                      ...currentSession,
                      playlist: {
                          ...currentSession.playlist,
                          items: currentSession.playlist.items.map((item) =>
                              item.video.id === videoId
                                  ? { ...item, video: { ...item.video, starred: isStarred } }
                                  : item
                          ),
                      },
                  }
                : currentSession
        );
    }, []);

    const currentVideo = useMemo(() => {
        if (!session?.playback_state.video_id) return null;
        return (
            session.playlist.items.find((item) => item.video.id === session.playback_state.video_id)
                ?.video ?? null
        );
    }, [session]);

    const playlist = session?.playlist ?? null;
    const playbackState = session?.playback_state ?? null;

    const applyPreferencesToVideo = useCallback((videoElement: HTMLVideoElement) => {
        const { volume: savedVolume, isMuted, playbackRate: savedRate } = preferencesRef.current;
        videoElement.volume = savedVolume;
        videoElement.muted = isMuted;
        videoElement.playbackRate = savedRate;
    }, []);

    const loadStreamSource = useCallback(
        (videoToPlayId: number, seekSeconds: number, streamMode: StreamMode) => {
            const videoElement = videoRef.current;
            if (!videoElement) return;
            setPlaybackError(null);
            applyPreferencesToVideo(videoElement);
            streamModeRef.current = streamMode;
            if (streamMode === 'remux' || streamMode === 'transcode') {
                remuxOffsetSecondsRef.current = seekSeconds;
                videoElement.src =
                    streamMode === 'remux'
                        ? buildRemuxVideoStreamUrl(videoToPlayId, seekSeconds)
                        : buildTranscodeVideoStreamUrl(videoToPlayId, seekSeconds);
                setCurrentTime(seekSeconds);
            } else {
                remuxOffsetSecondsRef.current = null;
                videoElement.src = buildDirectVideoStreamUrl(videoToPlayId);
                if (seekSeconds > 0) {
                    videoElement.currentTime = seekSeconds;
                }
            }
            videoElement
                .play()
                .then(() => setStatus('playing'))
                .catch((rejection: unknown) => {
                    const errorKind = getPlaybackErrorKindFromPlayRejection(rejection);
                    const fallbackStreamMode = errorKind
                        ? findFallbackStreamMode(streamMode, errorKind)
                        : null;
                    if (fallbackStreamMode) {
                        reloadWithStreamModeRef.current(
                            videoToPlayId,
                            seekSeconds,
                            fallbackStreamMode
                        );
                        return;
                    }
                    setStatus('paused');
                    setPlaybackError(errorKind);
                });
        },
        [applyPreferencesToVideo]
    );

    useEffect(() => {
        reloadWithStreamModeRef.current = loadStreamSource;
    }, [loadStreamSource]);

    const attachVideoSource = useCallback(
        (videoToPlayId: number, seekSeconds: number, format?: string) => {
            const isContainerUnplayable = isContainerUnplayableByBrowser(
                format,
                videoRef.current?.canPlayType?.bind(videoRef.current)
            );
            loadStreamSource(
                videoToPlayId,
                seekSeconds,
                isContainerUnplayable ? 'remux' : 'direct'
            );
        },
        [loadStreamSource]
    );

    const reportPlaybackError = useCallback(
        (errorKind: PlaybackErrorKind) => {
            const latest = latestPlaybackRef.current;
            const currentVideoId = latest.session?.playback_state.video_id;
            if (!currentVideoId) {
                setStatus('paused');
                setPlaybackError(errorKind);
                return;
            }
            const nextStreamMode = findFallbackStreamMode(streamModeRef.current, errorKind);
            if (nextStreamMode) {
                loadStreamSource(currentVideoId, latest.currentTime, nextStreamMode);
                return;
            }
            setStatus('paused');
            setPlaybackError(errorKind);
        },
        [loadStreamSource]
    );

    const retryPlayback = useCallback(() => {
        const latest = latestPlaybackRef.current;
        const currentVideoId = latest.session?.playback_state.video_id;
        if (!currentVideoId) return;
        attachVideoSource(
            currentVideoId,
            latest.currentTime,
            findVideoFormat(latest.session, currentVideoId)
        );
    }, [attachVideoSource]);

    const buildPlaybackStateRequest = useCallback(
        (
            payload?: Partial<{
                currentTime: number;
                duration: number;
                isPaused: boolean;
                completed: boolean;
            }>
        ): UpdateVideoPlaybackStateRequest | null => {
            const latest = latestPlaybackRef.current;
            const playlistIdInSession = latest.session?.playback_state.playlist_id;
            const videoIdInSession = latest.session?.playback_state.video_id;
            if (!playlistIdInSession || !videoIdInSession) return null;
            const positionSeconds = payload?.currentTime ?? latest.currentTime;
            const durationSeconds = payload?.duration ?? latest.duration;
            const isCompleted =
                payload?.completed ??
                hasReachedCompletionThreshold(positionSeconds, durationSeconds);
            return {
                playlist_id: playlistIdInSession,
                video_id: videoIdInSession,
                current_time: persistProgress ? positionSeconds : 0,
                duration: durationSeconds,
                is_paused: payload?.isPaused ?? latest.status !== 'playing',
                completed: persistProgress ? isCompleted : false,
            };
        },
        [persistProgress]
    );

    const syncState = useCallback(
        async (
            payload?: Partial<{
                currentTime: number;
                duration: number;
                isPaused: boolean;
                completed: boolean;
            }>
        ) => {
            const request = buildPlaybackStateRequest(payload);
            if (!request) return;
            try {
                await updateVideoPlaybackState(request);
            } catch {
                // best effort sync
            }
        },
        [buildPlaybackStateRequest]
    );

    useEffect(() => {
        const flushCurrentState = () => {
            const request = buildPlaybackStateRequest();
            if (request) flushVideoPlaybackState(request);
        };
        const flushWhenPageBecomesHidden = () => {
            if (document.visibilityState === 'hidden') flushCurrentState();
        };

        window.addEventListener('pagehide', flushCurrentState);
        document.addEventListener('visibilitychange', flushWhenPageBecomesHidden);
        return () => {
            window.removeEventListener('pagehide', flushCurrentState);
            document.removeEventListener('visibilitychange', flushWhenPageBecomesHidden);
            flushCurrentState();
        };
    }, [buildPlaybackStateRequest]);

    const playVideo = useCallback(async () => {
        const response = await startVideoPlayback(Number(videoId), playlistId ?? null);
        setSession(response);
        const resumeTime = persistProgress ? response.playback_state.current_time || 0 : 0;
        setCurrentTime(resumeTime);
        setDuration(response.playback_state.duration || 0);
        if (response.playback_state.video_id) {
            attachVideoSource(
                response.playback_state.video_id,
                resumeTime,
                findVideoFormat(response, response.playback_state.video_id)
            );
            if (!persistProgress) {
                void syncState({ currentTime: 0, completed: false, isPaused: false });
            }
        }
    }, [attachVideoSource, persistProgress, playlistId, syncState, videoId]);

    const pause = useCallback(() => {
        if (videoRef.current) {
            videoRef.current.pause();
            setStatus('paused');
            syncState({ isPaused: true });
        }
    }, [syncState]);

    const resume = useCallback(() => {
        if (videoRef.current) {
            videoRef.current.play();
            setStatus('playing');
            syncState({ isPaused: false });
        }
    }, [syncState]);

    const seekTo = useCallback(
        (time: number) => {
            if (!videoRef.current) return;
            const currentVideoId = latestPlaybackRef.current.session?.playback_state.video_id;
            if (streamModeRef.current !== 'direct' && currentVideoId) {
                loadStreamSource(currentVideoId, time, streamModeRef.current);
                return;
            }
            videoRef.current.currentTime = time;
            setCurrentTime(time);
        },
        [loadStreamSource]
    );

    const persistPreferencesNow = useCallback(() => {
        if (volumePersistTimerRef.current != null) {
            window.clearTimeout(volumePersistTimerRef.current);
            volumePersistTimerRef.current = null;
        }
        saveVideoPlaybackPreferences(preferencesRef.current);
    }, []);

    const persistPreferencesDebounced = useCallback(() => {
        if (volumePersistTimerRef.current != null) {
            window.clearTimeout(volumePersistTimerRef.current);
        }
        volumePersistTimerRef.current = window.setTimeout(
            persistPreferencesNow,
            VOLUME_PERSIST_DEBOUNCE_MS
        );
    }, [persistPreferencesNow]);

    useEffect(
        () => () => {
            if (volumePersistTimerRef.current != null) persistPreferencesNow();
        },
        [persistPreferencesNow]
    );

    const setVolumeHandler = useCallback(
        (newVolume: number) => {
            if (!videoRef.current) return;
            const clampedVolume = Math.max(0, Math.min(1, newVolume));
            videoRef.current.volume = clampedVolume;
            videoRef.current.muted = false;
            setVolume(clampedVolume);
            setIsMuted(false);
            preferencesRef.current = {
                ...preferencesRef.current,
                volume: clampedVolume,
                isMuted: false,
            };
            persistPreferencesDebounced();
        },
        [persistPreferencesDebounced]
    );

    const toggleMute = useCallback(() => {
        if (!videoRef.current) return;
        const { volume: currentVolume, isMuted: wasMuted } = preferencesRef.current;
        const wasSilent = wasMuted || currentVolume === 0;
        const nextIsMuted = !wasSilent;
        const nextVolume = wasSilent && currentVolume === 0 ? DEFAULT_UNMUTE_VOLUME : currentVolume;
        videoRef.current.volume = nextVolume;
        videoRef.current.muted = nextIsMuted;
        setVolume(nextVolume);
        setIsMuted(nextIsMuted);
        preferencesRef.current = {
            ...preferencesRef.current,
            volume: nextVolume,
            isMuted: nextIsMuted,
        };
        persistPreferencesNow();
    }, [persistPreferencesNow]);

    const setPlaybackRateHandler = useCallback(
        (rate: number) => {
            if (!videoRef.current) return;
            videoRef.current.playbackRate = rate;
            setPlaybackRate(rate);
            preferencesRef.current = { ...preferencesRef.current, playbackRate: rate };
            persistPreferencesNow();
        },
        [persistPreferencesNow]
    );

    const nextVideo = useCallback(async () => {
        const response = await nextVideoPlayback();
        setSession(response);
        setCurrentTime(0);
        setDuration(0);
        if (response.playback_state.video_id) {
            attachVideoSource(
                response.playback_state.video_id,
                0,
                findVideoFormat(response, response.playback_state.video_id)
            );
        }
    }, [attachVideoSource]);

    const previousVideo = useCallback(async () => {
        const response = await previousVideoPlayback();
        setSession(response);
        setCurrentTime(0);
        setDuration(0);
        if (response.playback_state.video_id) {
            attachVideoSource(
                response.playback_state.video_id,
                0,
                findVideoFormat(response, response.playback_state.video_id)
            );
        }
    }, [attachVideoSource]);

    const togglePlayPause = useCallback(() => {
        if (status === 'playing') {
            pause();
        } else {
            resume();
        }
    }, [pause, resume, status]);

    const setCurrentTimeHandler = useCallback((time: number) => {
        setCurrentTime(time + (remuxOffsetSecondsRef.current ?? 0));
    }, []);

    const setDurationHandler = useCallback((loadedDuration: number) => {
        const remuxOffsetSeconds = remuxOffsetSecondsRef.current;
        if (remuxOffsetSeconds === null) {
            setDuration(loadedDuration);
            return;
        }
        const hasMetadataDuration = latestPlaybackRef.current.duration > 0;
        const hasLoadedDuration = Number.isFinite(loadedDuration) && loadedDuration > 0;
        if (!hasMetadataDuration && hasLoadedDuration) {
            setDuration(loadedDuration + remuxOffsetSeconds);
        }
    }, []);

    const onVideoEnded = useCallback(async () => {
        setStatus('paused');
        await syncState({ completed: true, isPaused: true, currentTime: duration });
    }, [duration, syncState]);

    useEffect(() => {
        syncTimerRef.current = window.setInterval(() => {
            syncState();
        }, 5000);

        return () => {
            if (syncTimerRef.current != null) {
                window.clearInterval(syncTimerRef.current);
                syncTimerRef.current = null;
            }
        };
    }, [syncState]);

    return {
        videoRef,
        playVideo,
        pause,
        resume,
        seekTo,
        setVolume: setVolumeHandler,
        toggleMute,
        setPlaybackRate: setPlaybackRateHandler,
        togglePlayPause,
        nextVideo,
        previousVideo,
        status,
        currentTime,
        duration,
        volume,
        isMuted,
        playbackRate,
        setCurrentTime: setCurrentTimeHandler,
        setDuration: setDurationHandler,
        playlistItems: session?.playlist.items ?? [],
        playlist,
        playbackState,
        currentVideo,
        onVideoEnded,
        playbackError,
        reportPlaybackError,
        retryPlayback,
        setVideoStarred,
    };
};

export default useVideoPlayer;

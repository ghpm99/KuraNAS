import { useCallback, useEffect, useRef, useState } from 'react';
import type { TranscodedStream } from './trackStreamSource';

export interface AudioEngineState {
    isPlaying: boolean;
    currentTime: number;
    duration: number;
}

export interface AudioEngine extends AudioEngineState {
    audioRef: React.RefObject<HTMLAudioElement | null>;
    loadAndPlayUrl: (url: string, transcodedStream?: TranscodedStream) => void;
    loadUrlPaused: (
        url: string,
        startPositionSeconds: number,
        transcodedStream?: TranscodedStream
    ) => void;
    canPlayType: (mimeType: string) => string;
    getPositionSeconds: () => number;
    preloadUrl: (url: string) => void;
    togglePlayPause: () => void;
    seek: (time: number) => void;
    setVolume: (volume: number) => void;
    stop: () => void;
    volume: number;
}

export default function useAudioEngine(onTrackEnded: () => void): AudioEngine {
    const [isPlaying, setIsPlaying] = useState(false);
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);
    const [volume, setVolumeState] = useState(1);

    const audioRef = useRef<HTMLAudioElement | null>(null);
    const preloadAudioRef = useRef<HTMLAudioElement | null>(null);
    const onTrackEndedRef = useRef(onTrackEnded);
    const endedHandledRef = useRef(false);
    const transcodedStreamRef = useRef<TranscodedStream | null>(null);
    const positionOffsetRef = useRef(0);

    useEffect(() => {
        onTrackEndedRef.current = onTrackEnded;
    }, [onTrackEnded]);

    useEffect(() => {
        const audio = new Audio();
        audio.volume = volume;
        audioRef.current = audio;

        const preload = new Audio();
        preload.preload = 'auto';
        preloadAudioRef.current = preload;

        const onTimeUpdate = () => setCurrentTime(audio.currentTime + positionOffsetRef.current);
        const onLoadedMetadata = () => {
            const transcodedStream = transcodedStreamRef.current;
            if (transcodedStream) {
                setDuration(transcodedStream.durationSeconds);
                return;
            }
            setDuration(audio.duration);
        };
        const onEnded = () => {
            if (endedHandledRef.current) return;
            endedHandledRef.current = true;
            onTrackEndedRef.current();
        };
        const onPause = () => setIsPlaying(false);
        const onPlay = () => setIsPlaying(true);
        const onError = () => {
            const failedTranscode = transcodedStreamRef.current;
            if (failedTranscode) {
                transcodedStreamRef.current = null;
                positionOffsetRef.current = 0;
                audio.src = failedTranscode.fallbackUrl;
                audio.play().catch(() => {});
                return;
            }
            if (audio.src && !endedHandledRef.current) {
                endedHandledRef.current = true;
                onTrackEndedRef.current();
            }
        };

        audio.addEventListener('timeupdate', onTimeUpdate);
        audio.addEventListener('loadedmetadata', onLoadedMetadata);
        audio.addEventListener('ended', onEnded);
        audio.addEventListener('pause', onPause);
        audio.addEventListener('play', onPlay);
        audio.addEventListener('error', onError);

        const fallbackInterval = setInterval(() => {
            if (endedHandledRef.current) return;
            if (!audio.src || audio.duration <= 0) return;
            if (audio.ended) {
                endedHandledRef.current = true;
                onTrackEndedRef.current();
            }
        }, 500);

        return () => {
            clearInterval(fallbackInterval);
            audio.removeEventListener('timeupdate', onTimeUpdate);
            audio.removeEventListener('loadedmetadata', onLoadedMetadata);
            audio.removeEventListener('ended', onEnded);
            audio.removeEventListener('pause', onPause);
            audio.removeEventListener('play', onPlay);
            audio.removeEventListener('error', onError);
            audio.pause();
            audio.src = '';
            preload.src = '';
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const preloadUrl = useCallback((url: string) => {
        const preload = preloadAudioRef.current;
        if (!preload || preload.src === url) return;
        preload.src = url;
        preload.load();
    }, []);

    const loadAndPlayUrl = useCallback((url: string, transcodedStream?: TranscodedStream) => {
        const audio = audioRef.current;
        if (!audio) return;
        endedHandledRef.current = false;
        transcodedStreamRef.current = transcodedStream ?? null;
        positionOffsetRef.current = 0;
        if (transcodedStream) setDuration(transcodedStream.durationSeconds);
        audio.src = url;
        audio.play().catch(() => {});
    }, []);

    const loadUrlPaused = useCallback(
        (url: string, startPositionSeconds: number, transcodedStream?: TranscodedStream) => {
            const audio = audioRef.current;
            if (!audio) return;
            endedHandledRef.current = false;
            transcodedStreamRef.current = transcodedStream ?? null;
            setCurrentTime(startPositionSeconds);
            if (transcodedStream) {
                positionOffsetRef.current = Math.max(startPositionSeconds, 0);
                setDuration(transcodedStream.durationSeconds);
                audio.src = transcodedStream.buildUrl(positionOffsetRef.current);
                return;
            }
            positionOffsetRef.current = 0;
            audio.src = url;
            if (startPositionSeconds <= 0) return;
            const seekToStartPosition = () => {
                audio.currentTime = startPositionSeconds;
            };
            audio.addEventListener('loadedmetadata', seekToStartPosition, { once: true });
        },
        []
    );

    const togglePlayPause = useCallback(() => {
        const audio = audioRef.current;
        if (!audio) return;
        if (audio.paused) {
            audio.play().catch(() => {});
        } else {
            audio.pause();
        }
    }, []);

    const seek = useCallback((time: number) => {
        const audio = audioRef.current;
        if (!audio) return;
        const transcodedStream = transcodedStreamRef.current;
        if (!transcodedStream) {
            audio.currentTime = time;
            return;
        }
        const wasPlaying = !audio.paused;
        endedHandledRef.current = false;
        positionOffsetRef.current = time;
        setCurrentTime(time);
        audio.src = transcodedStream.buildUrl(time);
        if (wasPlaying) audio.play().catch(() => {});
    }, []);

    const canPlayType = useCallback(
        (mimeType: string) => audioRef.current?.canPlayType?.(mimeType) ?? 'maybe',
        []
    );

    const getPositionSeconds = useCallback(
        () => (audioRef.current?.currentTime ?? 0) + positionOffsetRef.current,
        []
    );

    const setVolume = useCallback((newVolume: number) => {
        const clamped = Math.max(0, Math.min(1, newVolume));
        setVolumeState(clamped);
        if (audioRef.current) {
            audioRef.current.volume = clamped;
        }
    }, []);

    const stop = useCallback(() => {
        const audio = audioRef.current;
        if (audio) {
            audio.pause();
            audio.src = '';
        }
        endedHandledRef.current = false;
        transcodedStreamRef.current = null;
        positionOffsetRef.current = 0;
        setIsPlaying(false);
        setCurrentTime(0);
        setDuration(0);
    }, []);

    return {
        audioRef,
        isPlaying,
        currentTime,
        duration,
        volume,
        loadAndPlayUrl,
        loadUrlPaused,
        canPlayType,
        getPositionSeconds,
        preloadUrl,
        togglePlayPause,
        seek,
        setVolume,
        stop,
    };
}

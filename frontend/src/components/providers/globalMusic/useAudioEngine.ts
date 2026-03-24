import { useCallback, useEffect, useRef, useState } from 'react';

export interface AudioEngineState {
    isPlaying: boolean;
    currentTime: number;
    duration: number;
}

export interface AudioEngine extends AudioEngineState {
    audioRef: React.RefObject<HTMLAudioElement | null>;
    loadAndPlayUrl: (url: string) => void;
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
    // Guards against double-firing (onended + fallback interval)
    const endedHandledRef = useRef(false);

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

        const onTimeUpdate = () => setCurrentTime(audio.currentTime);
        const onLoadedMetadata = () => setDuration(audio.duration);
        const onEnded = () => {
            if (endedHandledRef.current) return;
            endedHandledRef.current = true;
            onTrackEndedRef.current();
        };
        const onPause = () => setIsPlaying(false);
        const onPlay = () => setIsPlaying(true);
        const onError = () => {
            // If an error occurs mid-playback, attempt to advance to next track
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

        // Fallback: poll every 500ms for cases where `ended` doesn't fire in background
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

    const loadAndPlayUrl = useCallback((url: string) => {
        const audio = audioRef.current;
        if (!audio) return;
        endedHandledRef.current = false;
        audio.src = url;
        audio.play().catch(() => {});
    }, []);

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
        if (audioRef.current) {
            audioRef.current.currentTime = time;
        }
    }, []);

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
        preloadUrl,
        togglePlayPause,
        seek,
        setVolume,
        stop,
    };
}

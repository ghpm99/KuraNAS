import { useEffect, useRef } from 'react';
import type { IMusicData } from '../musicProvider/musicProvider';
import { getMusicTitle, getMusicArtist } from '@/utils/music';

interface MediaSessionOptions {
    currentTrack: IMusicData | undefined;
    isPlaying: boolean;
    onPlay: () => void;
    onPause: () => void;
    onNext: () => void;
    onPrevious: () => void;
    onSeekTo: (time: number) => void;
    currentTime: number;
    duration: number;
}

export default function useMediaSession({
    currentTrack,
    isPlaying,
    onPlay,
    onPause,
    onNext,
    onPrevious,
    onSeekTo,
    currentTime,
    duration,
}: MediaSessionOptions) {
    // Use refs so action handlers are registered once and always call the latest callbacks
    const onPlayRef = useRef(onPlay);
    const onPauseRef = useRef(onPause);
    const onNextRef = useRef(onNext);
    const onPreviousRef = useRef(onPrevious);
    const onSeekToRef = useRef(onSeekTo);

    useEffect(() => { onPlayRef.current = onPlay; }, [onPlay]);
    useEffect(() => { onPauseRef.current = onPause; }, [onPause]);
    useEffect(() => { onNextRef.current = onNext; }, [onNext]);
    useEffect(() => { onPreviousRef.current = onPrevious; }, [onPrevious]);
    useEffect(() => { onSeekToRef.current = onSeekTo; }, [onSeekTo]);

    // Register action handlers once — stable refs prevent re-registration gaps
    useEffect(() => {
        if (!('mediaSession' in navigator)) return;

        navigator.mediaSession.setActionHandler('play', () => onPlayRef.current());
        navigator.mediaSession.setActionHandler('pause', () => onPauseRef.current());
        navigator.mediaSession.setActionHandler('nexttrack', () => onNextRef.current());
        navigator.mediaSession.setActionHandler('previoustrack', () => onPreviousRef.current());
        navigator.mediaSession.setActionHandler('seekto', (details) => {
            if (details.seekTime !== undefined) {
                onSeekToRef.current(details.seekTime);
            }
        });

        return () => {
            navigator.mediaSession.setActionHandler('play', null);
            navigator.mediaSession.setActionHandler('pause', null);
            navigator.mediaSession.setActionHandler('nexttrack', null);
            navigator.mediaSession.setActionHandler('previoustrack', null);
            navigator.mediaSession.setActionHandler('seekto', null);
        };
    }, []);

    // Update metadata when track changes
    useEffect(() => {
        if (!('mediaSession' in navigator)) return;

        if (!currentTrack) {
            navigator.mediaSession.metadata = null;
            return;
        }

        navigator.mediaSession.metadata = new MediaMetadata({
            title: getMusicTitle(currentTrack),
            artist: getMusicArtist(currentTrack),
            album: currentTrack.metadata?.album || '',
        });
    }, [currentTrack]);

    // Update playback state
    useEffect(() => {
        if (!('mediaSession' in navigator)) return;
        navigator.mediaSession.playbackState = isPlaying ? 'playing' : 'paused';
    }, [isPlaying]);

    // Update position state for seek bar on lock screen
    useEffect(() => {
        if (!('mediaSession' in navigator) || !('setPositionState' in navigator.mediaSession))
            return;
        if (!currentTrack || duration <= 0) return;

        try {
            navigator.mediaSession.setPositionState({
                duration,
                playbackRate: 1,
                position: Math.min(currentTime, duration),
            });
        } catch {
            // Ignore errors from invalid position state
        }
    }, [currentTrack, currentTime, duration]);
}

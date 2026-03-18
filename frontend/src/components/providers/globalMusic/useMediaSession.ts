import { useEffect } from 'react';
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

    // Register action handlers
    useEffect(() => {
        if (!('mediaSession' in navigator)) return;

        navigator.mediaSession.setActionHandler('play', onPlay);
        navigator.mediaSession.setActionHandler('pause', onPause);
        navigator.mediaSession.setActionHandler('nexttrack', onNext);
        navigator.mediaSession.setActionHandler('previoustrack', onPrevious);
        navigator.mediaSession.setActionHandler('seekto', (details) => {
            if (details.seekTime !== undefined) {
                onSeekTo(details.seekTime);
            }
        });

        return () => {
            navigator.mediaSession.setActionHandler('play', null);
            navigator.mediaSession.setActionHandler('pause', null);
            navigator.mediaSession.setActionHandler('nexttrack', null);
            navigator.mediaSession.setActionHandler('previoustrack', null);
            navigator.mediaSession.setActionHandler('seekto', null);
        };
    }, [onPlay, onPause, onNext, onPrevious, onSeekTo]);

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

import { useEffect, useRef } from 'react';
import type { IMusicData } from '../musicProvider/musicProvider';
import { getTrackCoverArtwork } from '@/service/musicCover';
import { getMusicTitle, getMusicArtist } from '@/utils/music';

const DEFAULT_SEEK_OFFSET_SECONDS = 10;

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
    const onPlayRef = useRef(onPlay);
    const onPauseRef = useRef(onPause);
    const onNextRef = useRef(onNext);
    const onPreviousRef = useRef(onPrevious);
    const onSeekToRef = useRef(onSeekTo);
    const isPlayingRef = useRef(isPlaying);
    const currentTimeRef = useRef(currentTime);
    const durationRef = useRef(duration);

    useEffect(() => { onPlayRef.current = onPlay; }, [onPlay]);
    useEffect(() => { onPauseRef.current = onPause; }, [onPause]);
    useEffect(() => { onNextRef.current = onNext; }, [onNext]);
    useEffect(() => { onPreviousRef.current = onPrevious; }, [onPrevious]);
    useEffect(() => { onSeekToRef.current = onSeekTo; }, [onSeekTo]);
    useEffect(() => { isPlayingRef.current = isPlaying; }, [isPlaying]);
    useEffect(() => { currentTimeRef.current = currentTime; }, [currentTime]);
    useEffect(() => { durationRef.current = duration; }, [duration]);

    useEffect(() => {
        if (!('mediaSession' in navigator)) return;

        const seekRelative = (offsetSeconds: number) => {
            const upperBound = durationRef.current > 0 ? durationRef.current : Infinity;
            const targetTime = Math.min(
                Math.max(currentTimeRef.current + offsetSeconds, 0),
                upperBound
            );
            onSeekToRef.current(targetTime);
        };

        const actionHandlers: [MediaSessionAction, MediaSessionActionHandler][] = [
            ['play', () => onPlayRef.current()],
            ['pause', () => onPauseRef.current()],
            ['nexttrack', () => onNextRef.current()],
            ['previoustrack', () => onPreviousRef.current()],
            [
                'seekto',
                (details) => {
                    if (details.seekTime !== undefined && details.seekTime !== null) {
                        onSeekToRef.current(details.seekTime);
                    }
                },
            ],
            [
                'seekbackward',
                (details) => seekRelative(-(details.seekOffset ?? DEFAULT_SEEK_OFFSET_SECONDS)),
            ],
            [
                'seekforward',
                (details) => seekRelative(details.seekOffset ?? DEFAULT_SEEK_OFFSET_SECONDS),
            ],
            [
                'stop',
                () => {
                    if (isPlayingRef.current) onPauseRef.current();
                    onSeekToRef.current(0);
                },
            ],
        ];

        const trySetActionHandler = (
            action: MediaSessionAction,
            handler: MediaSessionActionHandler | null
        ) => {
            try {
                navigator.mediaSession.setActionHandler(action, handler);
            } catch {
                return;
            }
        };

        actionHandlers.forEach(([action, handler]) => trySetActionHandler(action, handler));

        return () => {
            actionHandlers.forEach(([action]) => trySetActionHandler(action, null));
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
            artwork: getTrackCoverArtwork(currentTrack.id),
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

import { useEffect, useRef } from 'react';
import { getVideoThumbnailUrl, type VideoFileDto } from '@/service/videoPlayback';

const DEFAULT_SEEK_OFFSET_SECONDS = 10;
const ARTWORK_SIZE = 640;

interface VideoMediaSessionOptions {
    currentVideo: VideoFileDto | null;
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    hasPreviousVideo: boolean;
    hasNextVideo: boolean;
    togglePlayPause: () => void;
    seekTo: (time: number) => void;
    nextVideo: () => void;
    previousVideo: () => void;
}

const isMediaSessionAvailable = () =>
    typeof navigator !== 'undefined' && 'mediaSession' in navigator;

const setActionHandlerSafely = (
    action: MediaSessionAction,
    handler: MediaSessionActionHandler | null
) => {
    try {
        navigator.mediaSession.setActionHandler(action, handler);
    } catch {
        return;
    }
};

const getFolderName = (parentPath: string) => {
    const segments = parentPath.split(/[\\/]/).filter(Boolean);
    return segments[segments.length - 1] ?? '';
};

const buildMediaMetadata = (video: Pick<VideoFileDto, 'id' | 'name' | 'parent_path'>) =>
    new MediaMetadata({
        title: video.name,
        artist: getFolderName(video.parent_path),
        artwork: [
            {
                src: getVideoThumbnailUrl(video.id, ARTWORK_SIZE, ARTWORK_SIZE),
                sizes: `${ARTWORK_SIZE}x${ARTWORK_SIZE}`,
            },
        ],
    });

export default function useVideoMediaSession({
    currentVideo,
    isPlaying,
    currentTime,
    duration,
    hasPreviousVideo,
    hasNextVideo,
    togglePlayPause,
    seekTo,
    nextVideo,
    previousVideo,
}: VideoMediaSessionOptions) {
    const latestRef = useRef({
        isPlaying,
        currentTime,
        duration,
        togglePlayPause,
        seekTo,
        nextVideo,
        previousVideo,
    });

    useEffect(() => {
        latestRef.current = {
            isPlaying,
            currentTime,
            duration,
            togglePlayPause,
            seekTo,
            nextVideo,
            previousVideo,
        };
    });

    const currentVideoId = currentVideo?.id;
    const currentVideoName = currentVideo?.name;
    const currentVideoParentPath = currentVideo?.parent_path;

    useEffect(() => {
        if (!isMediaSessionAvailable() || typeof MediaMetadata === 'undefined') return;
        if (currentVideoId === undefined) return;

        navigator.mediaSession.metadata = buildMediaMetadata({
            id: currentVideoId,
            name: currentVideoName ?? '',
            parent_path: currentVideoParentPath ?? '',
        });

        return () => {
            if (!isMediaSessionAvailable()) return;
            navigator.mediaSession.metadata = null;
        };
    }, [currentVideoId, currentVideoName, currentVideoParentPath]);

    useEffect(() => {
        if (!isMediaSessionAvailable()) return;

        const seekRelative = (offsetSeconds: number) => {
            const { currentTime: positionSeconds, duration: totalSeconds } = latestRef.current;
            const upperBound = totalSeconds > 0 ? totalSeconds : Infinity;
            latestRef.current.seekTo(
                Math.min(Math.max(positionSeconds + offsetSeconds, 0), upperBound)
            );
        };

        const actionHandlers: [MediaSessionAction, MediaSessionActionHandler][] = [
            [
                'play',
                () => {
                    if (!latestRef.current.isPlaying) latestRef.current.togglePlayPause();
                },
            ],
            [
                'pause',
                () => {
                    if (latestRef.current.isPlaying) latestRef.current.togglePlayPause();
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
                'seekto',
                (details) => {
                    if (details.seekTime === undefined || details.seekTime === null) return;
                    latestRef.current.seekTo(details.seekTime);
                },
            ],
        ];
        if (hasPreviousVideo) {
            actionHandlers.push(['previoustrack', () => latestRef.current.previousVideo()]);
        }
        if (hasNextVideo) {
            actionHandlers.push(['nexttrack', () => latestRef.current.nextVideo()]);
        }

        actionHandlers.forEach(([action, handler]) => setActionHandlerSafely(action, handler));
        if (!hasPreviousVideo) setActionHandlerSafely('previoustrack', null);
        if (!hasNextVideo) setActionHandlerSafely('nexttrack', null);

        return () => {
            if (!isMediaSessionAvailable()) return;
            actionHandlers.forEach(([action]) => setActionHandlerSafely(action, null));
        };
    }, [hasPreviousVideo, hasNextVideo]);

    useEffect(() => {
        if (!isMediaSessionAvailable()) return;
        navigator.mediaSession.playbackState = isPlaying ? 'playing' : 'paused';
    }, [isPlaying]);

    useEffect(() => {
        if (!isMediaSessionAvailable()) return;
        return () => {
            if (!isMediaSessionAvailable()) return;
            navigator.mediaSession.playbackState = 'none';
        };
    }, []);

    useEffect(() => {
        if (!isMediaSessionAvailable() || !('setPositionState' in navigator.mediaSession)) return;
        if (currentVideoId === undefined || !Number.isFinite(duration) || duration <= 0) return;

        try {
            navigator.mediaSession.setPositionState({
                duration,
                playbackRate: 1,
                position: Math.min(Math.max(currentTime, 0), duration),
            });
        } catch {
            return;
        }
    }, [currentVideoId, currentTime, duration]);
}

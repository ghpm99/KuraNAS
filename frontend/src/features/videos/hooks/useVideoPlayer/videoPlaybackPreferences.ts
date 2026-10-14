import {
    MAX_PLAYBACK_RATE,
    MIN_PLAYBACK_RATE,
} from '@/features/videos/playerShortcuts/playerShortcutMapping';

export type VideoPlaybackPreferences = {
    volume: number;
    isMuted: boolean;
    playbackRate: number;
};

export const defaultVideoPlaybackPreferences: VideoPlaybackPreferences = {
    volume: 1,
    isMuted: false,
    playbackRate: 1,
};

export const DEFAULT_UNMUTE_VOLUME = 0.5;

export const videoPlaybackPreferencesStorageKey = 'kuranas.videos.playbackPreferences.v1';

const isNumberWithin = (candidate: unknown, minimum: number, maximum: number): boolean =>
    typeof candidate === 'number' &&
    Number.isFinite(candidate) &&
    candidate >= minimum &&
    candidate <= maximum;

const isVideoPlaybackPreferences = (candidate: unknown): candidate is VideoPlaybackPreferences => {
    if (typeof candidate !== 'object' || candidate === null) {
        return false;
    }
    const { volume, isMuted, playbackRate } = candidate as Partial<VideoPlaybackPreferences>;
    return (
        isNumberWithin(volume, 0, 1) &&
        typeof isMuted === 'boolean' &&
        isNumberWithin(playbackRate, MIN_PLAYBACK_RATE, MAX_PLAYBACK_RATE)
    );
};

export const loadVideoPlaybackPreferences = (): VideoPlaybackPreferences => {
    try {
        const storedPreferences = window.localStorage.getItem(videoPlaybackPreferencesStorageKey);
        if (!storedPreferences) {
            return defaultVideoPlaybackPreferences;
        }
        const parsedPreferences: unknown = JSON.parse(storedPreferences);
        return isVideoPlaybackPreferences(parsedPreferences)
            ? parsedPreferences
            : defaultVideoPlaybackPreferences;
    } catch {
        return defaultVideoPlaybackPreferences;
    }
};

export const saveVideoPlaybackPreferences = (preferences: VideoPlaybackPreferences): void => {
    try {
        window.localStorage.setItem(
            videoPlaybackPreferencesStorageKey,
            JSON.stringify(preferences)
        );
    } catch {
        return;
    }
};

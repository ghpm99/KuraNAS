import { getPrimaryLanguage } from '@/features/videos/hooks/useVideoSubtitles/subtitleSelection';

export const videoAudioPreferencesStorageKey = 'kuranas.videos.audioPreferences.v1';

export type VideoAudioPreferences = {
    language: string;
};

export const loadVideoAudioPreferences = (): VideoAudioPreferences | null => {
    try {
        const storedPreferences = window.localStorage.getItem(videoAudioPreferencesStorageKey);
        if (!storedPreferences) return null;
        const parsedPreferences: unknown = JSON.parse(storedPreferences);
        if (typeof parsedPreferences !== 'object' || parsedPreferences === null) return null;
        const { language } = parsedPreferences as Record<string, unknown>;
        if (typeof language !== 'string' || getPrimaryLanguage(language) === '') return null;
        return { language: getPrimaryLanguage(language) };
    } catch {
        return null;
    }
};

export const saveVideoAudioPreferences = (preferences: VideoAudioPreferences): void => {
    try {
        window.localStorage.setItem(
            videoAudioPreferencesStorageKey,
            JSON.stringify({ language: getPrimaryLanguage(preferences.language) })
        );
    } catch {
        return;
    }
};

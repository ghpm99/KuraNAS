export const videoSubtitlePreferencesStorageKey = 'kuranas.videos.subtitlePreferences.v1';

export type VideoSubtitlePreferences = {
    isEnabled: boolean;
    language: string;
};

const parseVideoSubtitlePreferences = (candidate: unknown): VideoSubtitlePreferences | null => {
    if (typeof candidate !== 'object' || candidate === null) {
        return null;
    }
    const { isEnabled, language } = candidate as Record<string, unknown>;
    if (typeof language !== 'string') {
        return null;
    }
    const hasLanguage = language.trim().length > 0;
    if (isEnabled === undefined) {
        return hasLanguage ? { isEnabled: true, language } : null;
    }
    if (typeof isEnabled !== 'boolean' || (isEnabled && !hasLanguage)) {
        return null;
    }
    return { isEnabled, language };
};

export const loadVideoSubtitlePreferences = (): VideoSubtitlePreferences | null => {
    try {
        const storedPreferences = window.localStorage.getItem(videoSubtitlePreferencesStorageKey);
        if (!storedPreferences) {
            return null;
        }
        const parsedPreferences: unknown = JSON.parse(storedPreferences);
        return parseVideoSubtitlePreferences(parsedPreferences);
    } catch {
        return null;
    }
};

export const saveVideoSubtitlePreferences = (preferences: VideoSubtitlePreferences): void => {
    try {
        window.localStorage.setItem(
            videoSubtitlePreferencesStorageKey,
            JSON.stringify(preferences)
        );
    } catch {
        return;
    }
};

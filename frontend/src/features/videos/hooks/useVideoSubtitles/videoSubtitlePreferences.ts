export const videoSubtitlePreferencesStorageKey = 'kuranas.videos.subtitlePreferences.v1';

export type VideoSubtitlePreferences = {
    language: string;
};

const isVideoSubtitlePreferences = (candidate: unknown): candidate is VideoSubtitlePreferences => {
    if (typeof candidate !== 'object' || candidate === null) {
        return false;
    }
    const { language } = candidate as Partial<VideoSubtitlePreferences>;
    return typeof language === 'string' && language.trim().length > 0;
};

export const loadVideoSubtitlePreferences = (): VideoSubtitlePreferences | null => {
    try {
        const storedPreferences = window.localStorage.getItem(videoSubtitlePreferencesStorageKey);
        if (!storedPreferences) {
            return null;
        }
        const parsedPreferences: unknown = JSON.parse(storedPreferences);
        return isVideoSubtitlePreferences(parsedPreferences) ? parsedPreferences : null;
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

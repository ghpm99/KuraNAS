export const autoplayNextPreferenceStorageKey = 'kuranas.videos.autoplayNext.v1';

export const loadAutoplayNextPreference = (fallbackValue: boolean): boolean => {
    try {
        const storedPreference = window.localStorage.getItem(autoplayNextPreferenceStorageKey);
        if (storedPreference === null) {
            return fallbackValue;
        }
        const parsedPreference: unknown = JSON.parse(storedPreference);
        return typeof parsedPreference === 'boolean' ? parsedPreference : fallbackValue;
    } catch {
        return fallbackValue;
    }
};

export const saveAutoplayNextPreference = (isAutoplayEnabled: boolean): void => {
    try {
        window.localStorage.setItem(
            autoplayNextPreferenceStorageKey,
            JSON.stringify(isAutoplayEnabled)
        );
    } catch {
        return;
    }
};

const starredOverridesByTrackId = new Map<number, boolean>();
const changeListeners = new Set<() => void>();

const notifyChangeListeners = () => changeListeners.forEach((listener) => listener());

export const getStarredOverride = (trackId: number): boolean | undefined =>
    starredOverridesByTrackId.get(trackId);

export const setStarredOverride = (trackId: number, isStarred: boolean) => {
    starredOverridesByTrackId.set(trackId, isStarred);
    notifyChangeListeners();
};

export const subscribeToStarredOverrides = (listener: () => void) => {
    changeListeners.add(listener);
    return () => {
        changeListeners.delete(listener);
    };
};

export const clearAllStarredOverrides = () => {
    starredOverridesByTrackId.clear();
    notifyChangeListeners();
};

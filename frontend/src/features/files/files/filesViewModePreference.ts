export type FilesViewMode = 'grid' | 'list';

export const defaultFilesViewMode: FilesViewMode = 'grid';

export const filesViewModeStorageKey = 'kuranas.files.viewMode';

const isFilesViewMode = (candidate: unknown): candidate is FilesViewMode =>
    candidate === 'grid' || candidate === 'list';

export const loadFilesViewMode = (): FilesViewMode => {
    try {
        const storedViewMode = window.localStorage.getItem(filesViewModeStorageKey);
        return isFilesViewMode(storedViewMode) ? storedViewMode : defaultFilesViewMode;
    } catch {
        return defaultFilesViewMode;
    }
};

export const saveFilesViewMode = (viewMode: FilesViewMode): void => {
    try {
        window.localStorage.setItem(filesViewModeStorageKey, viewMode);
    } catch {
        return;
    }
};

export type DocumentsViewMode = 'grid' | 'list';

export const defaultDocumentsViewMode: DocumentsViewMode = 'list';

export const documentsViewModeStorageKey = 'kuranas.documents.viewMode';

const isDocumentsViewMode = (candidate: unknown): candidate is DocumentsViewMode =>
    candidate === 'grid' || candidate === 'list';

export const loadDocumentsViewMode = (): DocumentsViewMode => {
    try {
        const storedViewMode = window.localStorage.getItem(documentsViewModeStorageKey);
        return isDocumentsViewMode(storedViewMode) ? storedViewMode : defaultDocumentsViewMode;
    } catch {
        return defaultDocumentsViewMode;
    }
};

export const saveDocumentsViewMode = (viewMode: DocumentsViewMode): void => {
    try {
        window.localStorage.setItem(documentsViewModeStorageKey, viewMode);
    } catch {
        return;
    }
};

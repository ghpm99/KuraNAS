import {
    defaultDocumentsViewMode,
    documentsViewModeStorageKey,
    loadDocumentsViewMode,
    saveDocumentsViewMode,
} from './documentsViewModePreference';

describe('documentsViewModePreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('defaults to list when nothing is stored', () => {
        expect(defaultDocumentsViewMode).toBe('list');
        expect(loadDocumentsViewMode()).toBe('list');
    });

    it('round-trips a saved view mode', () => {
        saveDocumentsViewMode('grid');

        expect(window.localStorage.getItem(documentsViewModeStorageKey)).toBe('grid');
        expect(loadDocumentsViewMode()).toBe('grid');
    });

    it('ignores invalid stored values', () => {
        window.localStorage.setItem(documentsViewModeStorageKey, 'carousel');

        expect(loadDocumentsViewMode()).toBe('list');
    });

    it('survives a storage that throws', () => {
        const getItem = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setItem = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadDocumentsViewMode()).toBe('list');
        expect(() => saveDocumentsViewMode('grid')).not.toThrow();

        getItem.mockRestore();
        setItem.mockRestore();
    });
});

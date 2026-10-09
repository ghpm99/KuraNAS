import {
    defaultFilesViewMode,
    filesViewModeStorageKey,
    loadFilesViewMode,
    saveFilesViewMode,
} from './filesViewModePreference';

describe('filesViewModePreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('falls back to the default when nothing is stored', () => {
        expect(loadFilesViewMode()).toBe(defaultFilesViewMode);
    });

    it('round-trips a saved view mode', () => {
        saveFilesViewMode('list');

        expect(window.localStorage.getItem(filesViewModeStorageKey)).toBe('list');
        expect(loadFilesViewMode()).toBe('list');
    });

    it('ignores invalid stored values', () => {
        window.localStorage.setItem(filesViewModeStorageKey, 'carousel');

        expect(loadFilesViewMode()).toBe(defaultFilesViewMode);
    });

    it('survives a storage that throws', () => {
        const getItemSpy = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setItemSpy = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadFilesViewMode()).toBe(defaultFilesViewMode);
        expect(() => saveFilesViewMode('list')).not.toThrow();

        getItemSpy.mockRestore();
        setItemSpy.mockRestore();
    });
});

import {
    defaultVideoLibrarySort,
    loadVideoLibrarySort,
    saveVideoLibrarySort,
    videoLibrarySortStorageKey,
} from './videoLibrarySortPreference';

describe('videoLibrarySortPreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('returns the default sort when nothing is stored', () => {
        expect(loadVideoLibrarySort()).toEqual(defaultVideoLibrarySort);
    });

    it('round-trips a saved sort', () => {
        saveVideoLibrarySort({ key: 'duration', order: 'asc' });

        expect(loadVideoLibrarySort()).toEqual({ key: 'duration', order: 'asc' });
    });

    it('falls back to default for corrupted or unknown stored values', () => {
        window.localStorage.setItem(videoLibrarySortStorageKey, '{not json');
        expect(loadVideoLibrarySort()).toEqual(defaultVideoLibrarySort);

        window.localStorage.setItem(
            videoLibrarySortStorageKey,
            JSON.stringify({ key: 'bogus', order: 'asc' })
        );
        expect(loadVideoLibrarySort()).toEqual(defaultVideoLibrarySort);

        window.localStorage.setItem(videoLibrarySortStorageKey, 'null');
        expect(loadVideoLibrarySort()).toEqual(defaultVideoLibrarySort);
    });

    it('survives a storage that throws', () => {
        const getItem = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setItem = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadVideoLibrarySort()).toEqual(defaultVideoLibrarySort);
        expect(() => saveVideoLibrarySort({ key: 'name', order: 'desc' })).not.toThrow();

        getItem.mockRestore();
        setItem.mockRestore();
    });
});

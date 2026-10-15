import {
    defaultStarredSort,
    loadStarredSort,
    saveStarredSort,
    starredSortStorageKey,
} from './starredSortPreference';

describe('starredSortPreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
        jest.restoreAllMocks();
    });

    it('returns the default sort when nothing is stored', () => {
        expect(loadStarredSort()).toBe(defaultStarredSort);
    });

    it('round-trips a saved sort', () => {
        saveStarredSort('starred_at');

        expect(window.localStorage.getItem(starredSortStorageKey)).toBe('starred_at');
        expect(loadStarredSort()).toBe('starred_at');
    });

    it('falls back to default for unknown stored values', () => {
        window.localStorage.setItem(starredSortStorageKey, 'bogus');

        expect(loadStarredSort()).toBe(defaultStarredSort);
    });

    it('survives a storage that throws', () => {
        jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadStarredSort()).toBe(defaultStarredSort);
        expect(() => saveStarredSort('starred_at')).not.toThrow();
    });
});

import {
    addSearchHistoryEntry,
    clearSearchHistory,
    readSearchHistory,
    removeSearchHistoryEntry,
    searchHistoryMaxEntries,
    searchHistoryStorageKey,
} from './searchHistoryStorage';

describe('searchHistoryStorage', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('reads an empty history when storage has nothing', () => {
        expect(readSearchHistory()).toEqual([]);
    });

    it('adds the newest query first and persists it', () => {
        const afterFirst = addSearchHistoryEntry([], 'ferias');
        const afterSecond = addSearchHistoryEntry(afterFirst, 'contrato');

        expect(afterSecond).toEqual(['contrato', 'ferias']);
        expect(readSearchHistory()).toEqual(['contrato', 'ferias']);
    });

    it('moves a repeated query to the front ignoring case and spaces', () => {
        const entries = addSearchHistoryEntry(['contrato', 'ferias'], '  FERIAS ');

        expect(entries).toEqual(['FERIAS', 'contrato']);
    });

    it('ignores queries shorter than two characters', () => {
        expect(addSearchHistoryEntry(['ab'], ' a ')).toEqual(['ab']);
    });

    it('caps the history at the maximum number of entries', () => {
        let entries: string[] = [];
        for (let index = 0; index < searchHistoryMaxEntries + 5; index += 1) {
            entries = addSearchHistoryEntry(entries, `query ${index}`);
        }

        expect(entries).toHaveLength(searchHistoryMaxEntries);
        expect(entries[0]).toBe(`query ${searchHistoryMaxEntries + 4}`);
        expect(readSearchHistory()).toHaveLength(searchHistoryMaxEntries);
    });

    it('removes one entry', () => {
        const entries = removeSearchHistoryEntry(['contrato', 'ferias'], 'contrato');

        expect(entries).toEqual(['ferias']);
        expect(readSearchHistory()).toEqual(['ferias']);
    });

    it('clears the history', () => {
        addSearchHistoryEntry([], 'ferias');

        expect(clearSearchHistory()).toEqual([]);
        expect(readSearchHistory()).toEqual([]);
    });

    it('discards corrupted or malformed stored values', () => {
        window.localStorage.setItem(searchHistoryStorageKey, '{not json');
        expect(readSearchHistory()).toEqual([]);

        window.localStorage.setItem(searchHistoryStorageKey, JSON.stringify({ a: 1 }));
        expect(readSearchHistory()).toEqual([]);

        window.localStorage.setItem(searchHistoryStorageKey, JSON.stringify(['ok query', 3, 'x']));
        expect(readSearchHistory()).toEqual(['ok query']);
    });

    describe('when storage throws', () => {
        beforeEach(() => {
            jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
                throw new Error('denied');
            });
            jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
                throw new Error('quota');
            });
        });

        afterEach(() => {
            jest.restoreAllMocks();
        });

        it('reads an empty history', () => {
            expect(readSearchHistory()).toEqual([]);
        });

        it('keeps the in-memory result for add, remove and clear', () => {
            expect(addSearchHistoryEntry([], 'ferias')).toEqual(['ferias']);
            expect(removeSearchHistoryEntry(['ferias'], 'ferias')).toEqual([]);
            expect(clearSearchHistory()).toEqual([]);
        });
    });
});

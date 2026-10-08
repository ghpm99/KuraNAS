import {
    defaultFilesSort,
    filesSortStorageKey,
    loadFilesSort,
    saveFilesSort,
} from './filesSortPreference';

describe('filesSortPreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('returns the default sort when nothing is stored', () => {
        expect(loadFilesSort()).toEqual(defaultFilesSort);
    });

    it('round-trips a saved sort', () => {
        saveFilesSort({ key: 'size', order: 'desc' });

        expect(loadFilesSort()).toEqual({ key: 'size', order: 'desc' });
    });

    it('falls back to default for corrupted or unknown stored values', () => {
        window.localStorage.setItem(filesSortStorageKey, '{not json');
        expect(loadFilesSort()).toEqual(defaultFilesSort);

        window.localStorage.setItem(
            filesSortStorageKey,
            JSON.stringify({ key: 'bogus', order: 'asc' })
        );
        expect(loadFilesSort()).toEqual(defaultFilesSort);

        window.localStorage.setItem(filesSortStorageKey, 'null');
        expect(loadFilesSort()).toEqual(defaultFilesSort);
    });

    it('survives a storage that throws', () => {
        const getItem = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setItem = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadFilesSort()).toEqual(defaultFilesSort);
        expect(() => saveFilesSort({ key: 'name', order: 'desc' })).not.toThrow();

        getItem.mockRestore();
        setItem.mockRestore();
    });
});

import {
    documentsSortStorageKey,
    loadDocumentsSort,
    saveDocumentsSort,
} from './documentsSortPreference';

describe('documentsSortPreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('has nothing stored by default', () => {
        expect(documentsSortStorageKey).toBe('kuranas.documents.sort');
        expect(loadDocumentsSort()).toBeNull();
    });

    it('round-trips a saved sort', () => {
        saveDocumentsSort({ sort: 'name', order: 'desc' });

        expect(loadDocumentsSort()).toEqual({ sort: 'name', order: 'desc' });
    });

    it.each(['not json', '"name"', 'null', '{"sort":"color","order":"asc"}', '{"sort":"name"}'])(
        'ignores the invalid stored value %s',
        (storedValue) => {
            window.localStorage.setItem(documentsSortStorageKey, storedValue);

            expect(loadDocumentsSort()).toBeNull();
        }
    );

    it('survives a storage that throws', () => {
        const getItem = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setItem = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadDocumentsSort()).toBeNull();
        expect(() => saveDocumentsSort({ sort: 'size', order: 'asc' })).not.toThrow();

        getItem.mockRestore();
        setItem.mockRestore();
    });
});

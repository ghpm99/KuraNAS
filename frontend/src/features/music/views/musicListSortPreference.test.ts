import {
    defaultMusicListSort,
    getMusicListSortStorageKey,
    loadMusicListSort,
    saveMusicListSort,
} from './musicListSortPreference';

describe('musicListSortPreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('returns the default sort when nothing is stored', () => {
        expect(loadMusicListSort('artists')).toEqual(defaultMusicListSort);
    });

    it('round-trips a saved sort per view independently', () => {
        saveMusicListSort('albums', { sort: 'year', order: 'asc' });
        saveMusicListSort('artists', { sort: 'name', order: 'asc' });

        expect(loadMusicListSort('albums')).toEqual({ sort: 'year', order: 'asc' });
        expect(loadMusicListSort('artists')).toEqual({ sort: 'name', order: 'asc' });
        expect(loadMusicListSort('genres')).toEqual(defaultMusicListSort);
    });

    it('falls back to default for corrupted, unknown or view-incompatible values', () => {
        const key = getMusicListSortStorageKey('genres');
        window.localStorage.setItem(key, '{not json');
        expect(loadMusicListSort('genres')).toEqual(defaultMusicListSort);

        window.localStorage.setItem(key, JSON.stringify({ sort: 'bogus', order: 'asc' }));
        expect(loadMusicListSort('genres')).toEqual(defaultMusicListSort);

        window.localStorage.setItem(key, JSON.stringify({ sort: 'year', order: 'asc' }));
        expect(loadMusicListSort('genres')).toEqual(defaultMusicListSort);

        window.localStorage.setItem(key, JSON.stringify({ sort: 'name', order: 'up' }));
        expect(loadMusicListSort('genres')).toEqual(defaultMusicListSort);

        window.localStorage.setItem(key, 'null');
        expect(loadMusicListSort('genres')).toEqual(defaultMusicListSort);
    });

    it('survives a storage that throws', () => {
        const getItem = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setItem = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadMusicListSort('folders')).toEqual(defaultMusicListSort);
        expect(() => saveMusicListSort('folders', defaultMusicListSort)).not.toThrow();

        getItem.mockRestore();
        setItem.mockRestore();
    });
});

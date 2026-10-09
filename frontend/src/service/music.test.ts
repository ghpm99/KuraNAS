jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    getMusicAlbumSummary,
    getMusicAlbums,
    getMusicAlbumsByArtist,
    getMusicArtistSummary,
    getMusicFolderSummary,
    getMusicGenreSummary,
    getMusicArtists,
    getMusicByAlbum,
    getMusicByArtist,
    getMusicByGenre,
    getMusicFolders,
    getMusicGenres,
    getMusicHomeCatalog,
    getMusicQueueByAlbum,
    getMusicQueueByArtist,
    getMusicQueueByFolder,
    getMusicQueueByGenre,
    searchMusicTracks,
} from './music';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
};

describe('service/music', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: { items: [], total: 0 } });
    });

    it('searches tracks with the term and pagination', async () => {
        await searchMusicTracks('queen', 2, 50);
        expect(mockedApi.get).toHaveBeenCalledWith('/music/search', {
            params: { q: 'queen', page: 2, page_size: 50 },
        });
    });

    it('gets artists list', async () => {
        await getMusicArtists(1, 20);
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/artists', {
            params: { page: 1, page_size: 20 },
        });
    });

    it('gets music by encoded artist', async () => {
        await getMusicByArtist('AC/DC', 2, 10);
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/artists/AC%2FDC/tracks', {
            params: { page: 2, page_size: 10 },
        });
    });

    it('gets albums and album details', async () => {
        await getMusicAlbums(1, 5);
        await getMusicByAlbum('Album X', 3, 15);

        expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/music/library/albums', {
            params: { page: 1, page_size: 5 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/music/library/albums/Album%20X/tracks', {
            params: { page: 3, page_size: 15 },
        });
    });

    it('gets genres and genre details', async () => {
        await getMusicGenres(1, 8);
        await getMusicByGenre('R&B/Soul', 4, 12);

        expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/music/library/genres', {
            params: { page: 1, page_size: 8 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(
            2,
            '/music/library/genres/R%26B%2FSoul/tracks',
            {
                params: { page: 4, page_size: 12 },
            }
        );
    });

    it('gets folders list', async () => {
        await getMusicFolders(9, 30);
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/folders', {
            params: { page: 9, page_size: 30 },
        });
    });

    it('gets music home catalog', async () => {
        await getMusicHomeCatalog(4);
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/home', {
            params: { limit: 4 },
        });
    });

    it.each([
        ['artists', getMusicArtists, '/music/library/artists'],
        ['albums', getMusicAlbums, '/music/library/albums'],
        ['genres', getMusicGenres, '/music/library/genres'],
        ['folders', getMusicFolders, '/music/library/folders'],
    ])('sends sort and order to the %s list when given', async (_name, fetchList, path) => {
        await fetchList(2, 50, { sort: 'recent', order: 'asc' });
        expect(mockedApi.get).toHaveBeenCalledWith(path, {
            params: { page: 2, page_size: 50, sort: 'recent', order: 'asc' },
        });
    });

    it('sends the sort to the home catalog when given', async () => {
        await getMusicHomeCatalog(4, { sort: 'recent', order: 'desc' });
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/home', {
            params: { limit: 4, sort: 'recent', order: 'desc' },
        });
    });

    it.each([
        ['artist', getMusicQueueByArtist, 'AC/DC', '/music/library/artists/AC%2FDC/queue'],
        [
            'album',
            getMusicQueueByAlbum,
            'ac/dc::back',
            '/music/library/albums/ac%2Fdc%3A%3Aback/queue',
        ],
        ['genre', getMusicQueueByGenre, 'r&b', '/music/library/genres/r%26b/queue'],
        [
            'folder',
            getMusicQueueByFolder,
            '/data/Rock',
            '/music/library/folders/%2Fdata%2FRock/queue',
        ],
    ])('fetches the %s queue without pagination params', async (_, fetchQueue, key, path) => {
        const queue = { items: [{ file_id: 1 }], truncated: false };
        mockedApi.get.mockResolvedValueOnce({ data: queue });

        const result = await fetchQueue(key);

        expect(mockedApi.get).toHaveBeenCalledWith(path);
        expect(result).toEqual(queue);
    });

    it.each([
        ['album', getMusicAlbumSummary, 'ac/dc::back', '/music/library/albums/ac%2Fdc%3A%3Aback'],
        ['artist', getMusicArtistSummary, 'AC/DC', '/music/library/artists/AC%2FDC'],
        ['genre', getMusicGenreSummary, 'r&b', '/music/library/genres/r%26b'],
        ['folder', getMusicFolderSummary, '/data/Rock', '/music/library/folders/%2Fdata%2FRock'],
    ])('fetches the %s summary without params', async (_, fetchSummary, key, path) => {
        const summary = { key, name: 'Name', track_count: 3, total_length_seconds: 120 };
        mockedApi.get.mockResolvedValueOnce({ data: summary });

        const result = await fetchSummary(key);

        expect(mockedApi.get).toHaveBeenCalledWith(path);
        expect(result).toEqual(summary);
    });

    it('gets the albums of an encoded artist with pagination', async () => {
        await getMusicAlbumsByArtist('AC/DC', 2, 24);
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/artists/AC%2FDC/albums', {
            params: { page: 2, page_size: 24 },
        });
    });
});

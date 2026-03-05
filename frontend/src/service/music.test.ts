jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	getMusicAlbums,
	getMusicArtists,
	getMusicByAlbum,
	getMusicByArtist,
	getMusicByGenre,
	getMusicFolders,
	getMusicGenres,
} from './music';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
};

describe('service/music', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: { items: [], total: 0 } });
	});

	it('gets artists list', async () => {
		await getMusicArtists(1, 20);
		expect(mockedApi.get).toHaveBeenCalledWith('/files/music/artists', {
			params: { page: 1, page_size: 20 },
		});
	});

	it('gets music by encoded artist', async () => {
		await getMusicByArtist('AC/DC', 2, 10);
		expect(mockedApi.get).toHaveBeenCalledWith('/files/music/artists/AC%2FDC', {
			params: { page: 2, page_size: 10 },
		});
	});

	it('gets albums and album details', async () => {
		await getMusicAlbums(1, 5);
		await getMusicByAlbum('Album X', 3, 15);

		expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/files/music/albums', {
			params: { page: 1, page_size: 5 },
		});
		expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/files/music/albums/Album%20X', {
			params: { page: 3, page_size: 15 },
		});
	});

	it('gets genres and genre details', async () => {
		await getMusicGenres(1, 8);
		await getMusicByGenre('R&B/Soul', 4, 12);

		expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/files/music/genres', {
			params: { page: 1, page_size: 8 },
		});
		expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/files/music/genres/R%26B%2FSoul', {
			params: { page: 4, page_size: 12 },
		});
	});

	it('gets folders list', async () => {
		await getMusicFolders(9, 30);
		expect(mockedApi.get).toHaveBeenCalledWith('/files/music/folders', {
			params: { page: 9, page_size: 30 },
		});
	});
});


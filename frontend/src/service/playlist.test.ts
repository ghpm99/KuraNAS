jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
		put: jest.fn(),
		delete: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	addTrackToPlaylist,
	createPlaylist,
	deletePlaylist,
	getNowPlayingPlaylist,
	getPlaylistById,
	getPlaylists,
	getPlaylistTracks,
	removeTrackFromPlaylist,
	updatePlaylist,
} from './playlist';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	put: jest.Mock;
	delete: jest.Mock;
};

describe('service/playlist', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets playlists with pagination params', async () => {
		const payload = { items: [], total: 0 };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getPlaylists(2, 25);

		expect(mockedApi.get).toHaveBeenCalledWith('/music/playlists/', {
			params: { page: 2, page_size: 25 },
		});
		expect(result).toEqual(payload);
	});

	it('gets now playing playlist', async () => {
		const payload = { id: 9, name: 'Now Playing' };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getNowPlayingPlaylist();

		expect(mockedApi.get).toHaveBeenCalledWith('/music/playlists/now-playing');
		expect(result).toEqual(payload);
	});

	it('gets playlist by id', async () => {
		const payload = { id: 11, name: 'Mix' };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getPlaylistById(11);

		expect(mockedApi.get).toHaveBeenCalledWith('/music/playlists/11');
		expect(result).toEqual(payload);
	});

	it('creates and updates a playlist', async () => {
		mockedApi.post.mockResolvedValue({ data: { id: 1, name: 'Nova' } });
		mockedApi.put.mockResolvedValue({ data: { id: 1, name: 'Atualizada' } });

		const created = await createPlaylist({ name: 'Nova' });
		const updated = await updatePlaylist(1, { name: 'Atualizada' });

		expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/', { name: 'Nova' });
		expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/1', { name: 'Atualizada' });
		expect(created).toEqual({ id: 1, name: 'Nova' });
		expect(updated).toEqual({ id: 1, name: 'Atualizada' });
	});

	it('deletes playlist', async () => {
		mockedApi.delete.mockResolvedValue({});
		await deletePlaylist(7);
		expect(mockedApi.delete).toHaveBeenCalledWith('/music/playlists/7');
	});

	it('gets playlist tracks with pagination', async () => {
		const payload = { items: [{ id: 1 }], total: 1 };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getPlaylistTracks(4, 1, 10);

		expect(mockedApi.get).toHaveBeenCalledWith('/music/playlists/4/tracks', {
			params: { page: 1, page_size: 10 },
		});
		expect(result).toEqual(payload);
	});

	it('adds and removes track from playlist', async () => {
		const payload = { id: 2, file_id: 10 };
		mockedApi.post.mockResolvedValue({ data: payload });
		mockedApi.delete.mockResolvedValue({});

		const created = await addTrackToPlaylist(2, 10);
		await removeTrackFromPlaylist(2, 10);

		expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/2/tracks', { file_id: 10 });
		expect(mockedApi.delete).toHaveBeenCalledWith('/music/playlists/2/tracks/10');
		expect(created).toEqual(payload);
	});
});

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
    addVideoToPlaylist,
    getAllVideoFiles,
    getVideoHomeCatalog,
    getVideoPlaybackState,
    getVideoPlaylistById,
    getVideoPlaylists,
    getVideosWithoutPlaylist,
    nextVideoPlayback,
    previousVideoPlayback,
    removeVideoFromPlaylist,
    reorderVideoPlaylist,
    setVideoPlaylistHidden,
    startVideoPlayback,
    updateVideoPlaybackState,
    updateVideoPlaylistName,
} from './videoPlayback';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
    put: jest.Mock;
    delete: jest.Mock;
};

describe('service/videoPlayback', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('starts playback with and without playlist id', async () => {
        mockedApi.post.mockResolvedValue({ data: { ok: true } });

        await startVideoPlayback(10);
        await startVideoPlayback(11, 99);

        expect(mockedApi.post).toHaveBeenNthCalledWith(1, '/video/playback/start', {
            video_id: 10,
            playlist_id: null,
        });
        expect(mockedApi.post).toHaveBeenNthCalledWith(2, '/video/playback/start', {
            video_id: 11,
            playlist_id: 99,
        });
    });

    it('gets and updates playback state', async () => {
        mockedApi.get.mockResolvedValue({
            data: { playback_state: { current_time: 12 } },
        });
        mockedApi.put.mockResolvedValue({ data: { current_time: 25 } });

        const state = await getVideoPlaybackState();
        const updated = await updateVideoPlaybackState({
            current_time: 25,
            is_paused: true,
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/video/playback/state');
        expect(mockedApi.put).toHaveBeenCalledWith('/video/playback/state', {
            current_time: 25,
            is_paused: true,
        });
        expect(state).toEqual({ playback_state: { current_time: 12 } });
        expect(updated).toEqual({ current_time: 25 });
    });

    it('moves playback to next and previous', async () => {
        mockedApi.post.mockResolvedValue({ data: { ok: true } });

        await nextVideoPlayback();
        await previousVideoPlayback();

        expect(mockedApi.post).toHaveBeenNthCalledWith(1, '/video/playback/next');
        expect(mockedApi.post).toHaveBeenNthCalledWith(2, '/video/playback/previous');
    });

    it('gets catalog and playlists with default and explicit params', async () => {
        mockedApi.get.mockResolvedValue({ data: [] });

        await getVideoHomeCatalog();
        await getVideoHomeCatalog(50);
        await getVideoPlaylists();
        await getVideoPlaylists(true);
        await getVideoPlaylistById(33);

        expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/video/catalog/home', {
            params: { limit: 24 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/video/catalog/home', {
            params: { limit: 50 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(3, '/video/playlists/', {
            params: { include_hidden: false },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(4, '/video/playlists/', {
            params: { include_hidden: true },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(5, '/video/playlists/33');
    });

    it('updates playlist visibility and order, add and remove video, and rename playlist', async () => {
        mockedApi.put.mockResolvedValue({});
        mockedApi.post.mockResolvedValue({});
        mockedApi.delete.mockResolvedValue({});

        await setVideoPlaylistHidden(3, true);
        await addVideoToPlaylist(3, 101);
        await removeVideoFromPlaylist(3, 101);
        await reorderVideoPlaylist(3, [{ video_id: 101, order_index: 0 }]);
        await updateVideoPlaylistName(3, 'Minha lista');

        expect(mockedApi.put).toHaveBeenNthCalledWith(1, '/video/playlists/3/hidden', {
            hidden: true,
        });
        expect(mockedApi.post).toHaveBeenCalledWith('/video/playlists/3/videos', {
            video_id: 101,
        });
        expect(mockedApi.delete).toHaveBeenCalledWith('/video/playlists/3/videos/101');
        expect(mockedApi.put).toHaveBeenNthCalledWith(2, '/video/playlists/3/reorder', {
            items: [{ video_id: 101, order_index: 0 }],
        });
        expect(mockedApi.put).toHaveBeenNthCalledWith(3, '/video/playlists/3', {
            name: 'Minha lista',
        });
    });

    it('gets unassigned videos and all videos list with items fallback', async () => {
        mockedApi.get
            .mockResolvedValueOnce({ data: [{ id: 1 }] })
            .mockResolvedValueOnce({ data: { items: [{ id: 2 }] } })
            .mockResolvedValueOnce({ data: {} });

        const unassigned = await getVideosWithoutPlaylist();
        const allVideos = await getAllVideoFiles();
        const emptyFallback = await getAllVideoFiles(10);

        expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/video/playlists/unassigned', {
            params: { limit: 2000 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/files/videos', {
            params: { page: 1, page_size: 2000 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(3, '/files/videos', {
            params: { page: 1, page_size: 10 },
        });
        expect(unassigned).toEqual([{ id: 1 }]);
        expect(allVideos).toEqual([{ id: 2 }]);
        expect(emptyFallback).toEqual([]);
    });
});

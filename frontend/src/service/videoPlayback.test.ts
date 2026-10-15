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
    getVideoContinueWatching,
    getVideoHomeCatalog,
    getVideoLibraryFiles,
    getVideoLibraryFolders,
    getVideoLibraryFolderVideos,
    getVideoLibraryMovies,
    getVideoPlaybackState,
    getVideoPlaylistById,
    getHiddenVideoPlaylists,
    getVideoPlaylists,
    getVideoPlaylistsBySection,
    searchVideoPlaylists,
    getVideoPlaylistsByVideo,
    nextVideoPlayback,
    previousVideoPlayback,
    removeVideoFromPlaylist,
    reorderVideoPlaylist,
    setVideoPlaylistHidden,
    setVideoWatched,
    startVideoPlayback,
    updateVideoPlaybackState,
    updateVideoPlaylistName,
    createVideoPlaylist,
    deleteVideoPlaylist,
    getVideoMoviePosterUrl,
    getVideoPlaylistPosterUrl,
    getVideoSubtitles,
    getVideoSubtitleKey,
    type VideoSubtitleDto,
    getVideoSubtitleVttUrl,
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

    it('requests the hidden playlists page', async () => {
        const hiddenPage = {
            items: [{ id: 4, name: 'Oculta', is_hidden: true }],
            pagination: { page: 2, page_size: 24, has_next: false, has_prev: true },
        };
        mockedApi.get.mockResolvedValue({ data: hiddenPage });

        await expect(getHiddenVideoPlaylists(2, 24)).resolves.toEqual(hiddenPage);

        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/hidden', {
            params: { page: 2, page_size: 24 },
        });
    });

    it('requests the playlists that contain one video', async () => {
        const playlistsOfVideo = [{ id: 3, name: 'Favorites', type: 'custom' }];
        mockedApi.get.mockResolvedValue({ data: playlistsOfVideo });

        await expect(getVideoPlaylistsByVideo(7)).resolves.toEqual(playlistsOfVideo);

        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/by-video/7');
    });

    it('requests one page of the playlists of a section', async () => {
        const sectionPage = {
            items: [],
            pagination: { page: 2, page_size: 24, has_next: false, has_prev: true },
        };
        mockedApi.get.mockResolvedValue({ data: sectionPage });

        await expect(getVideoPlaylistsBySection('clips', 2, 24)).resolves.toEqual(sectionPage);

        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/section/clips', {
            params: { page: 2, page_size: 24 },
        });
    });

    it('searches playlists by name with page and query', async () => {
        const searchPage = {
            items: [],
            pagination: { page: 1, page_size: 30, has_next: false, has_prev: false },
        };
        mockedApi.get.mockResolvedValue({ data: searchPage });

        await expect(searchVideoPlaylists('ação', 1, 30)).resolves.toEqual(searchPage);

        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/search', {
            params: { q: 'ação', page: 1, page_size: 30 },
        });
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

    it('gets the in-progress videos with default and explicit limit', async () => {
        const continueItems = [
            {
                video: { id: 5 },
                position_seconds: 12,
                duration_seconds: 60,
                updated_at: '2026-03-01T00:00:00Z',
            },
        ];
        mockedApi.get.mockResolvedValue({ data: continueItems });

        await expect(getVideoContinueWatching()).resolves.toEqual(continueItems);
        await getVideoContinueWatching(8);

        expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/video/continue', {
            params: { limit: 24 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/video/continue', {
            params: { limit: 8 },
        });
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

    it('marks a video watched or unwatched through the progress endpoint', async () => {
        mockedApi.put.mockResolvedValue({});

        await setVideoWatched(7, true);
        await setVideoWatched(7, false);

        expect(mockedApi.put).toHaveBeenNthCalledWith(1, '/video/progress/7/watched', {
            watched: true,
        });
        expect(mockedApi.put).toHaveBeenNthCalledWith(2, '/video/progress/7/watched', {
            watched: false,
        });
    });

    it('requests the folders under a parent and omits an empty parent', async () => {
        const payload = { items: [], pagination: { page: 2, page_size: 48, has_next: false } };
        mockedApi.get.mockResolvedValue({ data: payload });

        const nested = await getVideoLibraryFolders('/Series', 2, 48);
        await getVideoLibraryFolders('', 1, 48);

        expect(mockedApi.get).toHaveBeenNthCalledWith(1, '/video/library/folders', {
            params: { parent: '/Series', page: 2, page_size: 48 },
        });
        expect(mockedApi.get).toHaveBeenNthCalledWith(2, '/video/library/folders', {
            params: { parent: undefined, page: 1, page_size: 48 },
        });
        expect(nested).toEqual(payload);
    });

    it('requests the direct videos of a folder with pagination', async () => {
        const payload = { items: [{ id: 3 }], pagination: { page: 1, page_size: 24 } };
        mockedApi.get.mockResolvedValue({ data: payload });

        const folderVideos = await getVideoLibraryFolderVideos('/Series/S1', 1, 24);

        expect(mockedApi.get).toHaveBeenCalledWith('/video/library/folders/videos', {
            params: { path: '/Series/S1', page: 1, page_size: 24 },
        });
        expect(folderVideos).toEqual(payload);
    });

    it('requests the paginated movie library with the chosen sort', async () => {
        const payload = { items: [{ id: 9 }], pagination: { page: 2, page_size: 24 } };
        mockedApi.get.mockResolvedValue({ data: payload });

        const movies = await getVideoLibraryMovies('recent', 2, 24);

        expect(mockedApi.get).toHaveBeenCalledWith('/video/library/movies', {
            params: { sort: 'recent', page: 2, page_size: 24 },
        });
        expect(movies).toEqual(payload);
    });

    it('requests the library files with the chosen sort key and order', async () => {
        const payload = { items: [{ id: 3 }], pagination: { page: 1, page_size: 60 } };
        mockedApi.get.mockResolvedValue({ data: payload });

        const files = await getVideoLibraryFiles(1, 60, 'praia', { key: 'duration', order: 'asc' });

        expect(mockedApi.get).toHaveBeenCalledWith('/video/library/files', {
            params: { page: 1, page_size: 60, query: 'praia', sort: 'duration', order: 'asc' },
        });
        expect(files).toEqual(payload);
    });

    it('sends the format, resolution and duration filters only when defined', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [], pagination: {} } });

        await getVideoLibraryFiles(2, 60, '', undefined, {
            formats: ['.mp4', '.mkv'],
            resolution: 'fhd',
            duration: 'long',
        });

        expect(mockedApi.get).toHaveBeenLastCalledWith('/video/library/files', {
            params: {
                page: 2,
                page_size: 60,
                query: '',
                sort: undefined,
                order: undefined,
                format: '.mp4,.mkv',
                resolution: 'fhd',
                duration: 'long',
            },
        });
    });

    it('omits every filter param when no filter is active', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [], pagination: {} } });

        await getVideoLibraryFiles(1, 60, '', undefined, { formats: [] });

        expect(mockedApi.get).toHaveBeenLastCalledWith('/video/library/files', {
            params: {
                page: 1,
                page_size: 60,
                query: '',
                sort: undefined,
                order: undefined,
                format: undefined,
                resolution: undefined,
                duration: undefined,
            },
        });
    });

    it('creates a custom playlist posting only the name, or the name and the video id', async () => {
        mockedApi.post.mockResolvedValue({ data: { id: 8, name: 'Favoritos' } });

        const emptyPlaylist = await createVideoPlaylist('Favoritos');
        await createVideoPlaylist('Favoritos', 5);

        expect(emptyPlaylist).toEqual({ id: 8, name: 'Favoritos' });
        expect(mockedApi.post).toHaveBeenNthCalledWith(1, '/video/playlists', {
            name: 'Favoritos',
        });
        expect(mockedApi.post).toHaveBeenNthCalledWith(2, '/video/playlists', {
            name: 'Favoritos',
            video_id: 5,
        });
    });

    it('deletes a playlist by id', async () => {
        await deleteVideoPlaylist(8);

        expect(mockedApi.delete).toHaveBeenCalledWith('/video/playlists/8');
    });
});

describe('video poster urls', () => {
    it('builds the playlist poster url with the default and a custom size', () => {
        expect(getVideoPlaylistPosterUrl(7)).toMatch(
            /\/video\/playlists\/7\/poster\?width=400&height=600$/
        );
        expect(getVideoPlaylistPosterUrl(7, 200, 300)).toMatch(
            /\/video\/playlists\/7\/poster\?width=200&height=300$/
        );
    });

    it('builds the movie poster url', () => {
        expect(getVideoMoviePosterUrl(9)).toMatch(
            /\/video\/movies\/9\/poster\?width=400&height=600$/
        );
    });

    it('lists the subtitles of a video through the api base', async () => {
        const subtitles = [{ id: 9, label: 'en', language: 'en', format: 'srt', is_forced: false }];
        mockedApi.get.mockResolvedValueOnce({ data: subtitles });

        await expect(getVideoSubtitles(30)).resolves.toEqual(subtitles);
        expect(mockedApi.get).toHaveBeenCalledWith('/video/subtitles/30');
    });

    describe('subtitle identity and vtt url', () => {
        const externalSubtitle: VideoSubtitleDto = {
            id: 9,
            source: 'external',
            label: 'en',
            language: 'en',
            format: 'srt',
            is_forced: false,
        };
        const embeddedSubtitle: VideoSubtitleDto = {
            id: 0,
            source: 'embedded',
            key: '30:4',
            stream_index: 4,
            label: 'en',
            language: 'en',
            format: 'vtt',
            is_forced: false,
        };
        const legacySubtitle: VideoSubtitleDto = {
            id: 7,
            label: 'pt',
            language: 'pt',
            format: 'vtt',
            is_forced: false,
        };

        it('builds the external subtitle vtt url from the file id', () => {
            expect(getVideoSubtitleVttUrl(externalSubtitle, 30)).toMatch(
                /\/video\/subtitles\/9\/vtt$/
            );
        });

        it('builds the embedded subtitle vtt url from the video and stream index', () => {
            expect(getVideoSubtitleVttUrl(embeddedSubtitle, 30)).toMatch(
                /\/video\/subtitles\/30\/embedded\/4\/vtt$/
            );
        });

        it('treats a payload without source as external', () => {
            expect(getVideoSubtitleVttUrl(legacySubtitle, 30)).toMatch(
                /\/video\/subtitles\/7\/vtt$/
            );
            expect(getVideoSubtitleKey(legacySubtitle)).toBe('external:7');
        });

        it('gives embedded subtitles with id 0 distinct keys', () => {
            const otherEmbeddedSubtitle = { ...embeddedSubtitle, key: '30:5', stream_index: 5 };

            expect(getVideoSubtitleKey(embeddedSubtitle)).toBe('30:4');
            expect(getVideoSubtitleKey(otherEmbeddedSubtitle)).toBe('30:5');
            expect(getVideoSubtitleKey(externalSubtitle)).toBe('external:9');
        });

        it('falls back to the external key when an embedded subtitle has no key', () => {
            expect(getVideoSubtitleKey({ ...embeddedSubtitle, key: undefined })).toBe('external:0');
        });
    });
});

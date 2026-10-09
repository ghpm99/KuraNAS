import { act, renderHook, waitFor } from '@testing-library/react';
import { useMusicHomeScreen } from './useMusicHomeScreen';

const mockUseGlobalMusic = jest.fn();
const mockUseQuery = jest.fn();
const queueOf = (...fileIds: number[]) => ({
    items: fileIds.map((fileId) => ({
        file_id: fileId,
        name: `song-${fileId}`,
        path: `/music/song-${fileId}.mp3`,
        format: '.mp3',
        title: `Song ${fileId}`,
        artist: 'Artist A',
        album: 'Album A',
        length: 180,
    })),
    truncated: false,
});
const emptyQueue = { items: [], truncated: false };
const mockGetPlaylistQueue = jest.fn();
const mockGetMusicQueueByArtist = jest.fn();
const mockGetMusicQueueByAlbum = jest.fn();
const mockGetMusicHomeCatalog = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
    getMusicTitle: (m: any) => m.name ?? m.metadata?.title ?? `title-${m.id}`,
    getMusicArtist: (m: any) => m.metadata?.artist ?? `artist-${m.id}`,
    musicMetadata: () => 'meta',
    getTrackDurationSeconds: (metadata?: any) => metadata?.length ?? 0,
    formatMusicDuration: (s: number) =>
        `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`,
}));

jest.mock('@tanstack/react-query', () => ({
    useQuery: (...args: any[]) => mockUseQuery(...args),
}));

jest.mock('@/service/playlist', () => ({
    getPlaylistQueue: (...args: any[]) => mockGetPlaylistQueue(...args),
}));

jest.mock('@/service/music', () => ({
    getMusicQueueByArtist: (...args: any[]) => mockGetMusicQueueByArtist(...args),
    getMusicQueueByAlbum: (...args: any[]) => mockGetMusicQueueByAlbum(...args),
    getMusicHomeCatalog: (...args: any[]) => mockGetMusicHomeCatalog(...args),
}));

describe('useMusicHomeScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseGlobalMusic.mockReturnValue({
            currentIndex: 0,
            currentTrack: { id: 1 },
            hasQueue: true,
            playbackContext: { href: '/music/albums' },
            queue: [{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }],
            replaceQueue: jest.fn(),
            toggleQueue: jest.fn(),
        });
        mockUseQuery.mockReturnValue({
            data: {
                summary: {
                    total_tracks: 2,
                    total_artists: 2,
                    total_albums: 2,
                    total_genres: 1,
                    total_folders: 1,
                },
                playlists: [
                    {
                        id: 5,
                        name: 'Mix',
                        description: '',
                        track_count: 3,
                        is_system: false,
                        is_auto: false,
                        kind: 'manual',
                        source_key: '',
                    },
                ],
                artists: [
                    {
                        key: 'artist-a',
                        artist: 'Artist A',
                        track_count: 1,
                        album_count: 1,
                    },
                ],
                albums: [
                    {
                        key: 'artist-a::album-a',
                        album: 'Album A',
                        artist: 'Artist A',
                        year: '2024',
                        track_count: 1,
                    },
                ],
            },
            isLoading: false,
            status: 'success',
        });
        mockGetMusicHomeCatalog.mockResolvedValue({});
        mockGetPlaylistQueue.mockResolvedValue(queueOf(10));
        mockGetMusicQueueByArtist.mockResolvedValue(queueOf(20));
        mockGetMusicQueueByAlbum.mockResolvedValue(queueOf(30));
    });

    it('derives home metrics, next tracks, and featured cards', () => {
        const { result } = renderHook(() => useMusicHomeScreen());

        expect(result.current.totalTracks).toBe(2);
        expect(result.current.totalArtists).toBe(2);
        expect(result.current.totalAlbums).toBe(2);
        expect(result.current.totalPlaylists).toBe(1);
        expect(result.current.currentTrackTitle).toBe('title-1');
        expect(result.current.currentTrackArtist).toBe('artist-1');
        expect(result.current.nextTracks).toEqual([
            { id: 2, title: 'title-2', artist: 'artist-2' },
            { id: 3, title: 'title-3', artist: 'artist-3' },
            { id: 4, title: 'title-4', artist: 'artist-4' },
        ]);
        expect(result.current.returnToContextHref).toBe('/music/albums');
        expect(result.current.featuredPlaylists[0]).toMatchObject({
            href: '/music/playlists',
            actionKey: 'playlist-5',
        });
        expect(result.current.artistHighlights[0]).toMatchObject({
            href: '/music/artists',
        });
        expect(result.current.albumHighlights[0]).toMatchObject({
            href: '/music/albums',
        });
    });

    it('handles empty playback fetches and pending action state', async () => {
        const replaceQueue = jest.fn();
        let resolvePlaylistQueue: ((value: typeof emptyQueue) => void) | undefined;
        mockUseGlobalMusic.mockReturnValue({
            currentIndex: undefined,
            currentTrack: undefined,
            hasQueue: false,
            playbackContext: undefined,
            queue: [],
            replaceQueue,
            toggleQueue: jest.fn(),
        });
        mockUseQuery.mockReturnValue({
            data: undefined,
            isLoading: true,
            status: 'pending',
        });
        mockGetPlaylistQueue.mockImplementationOnce(
            () =>
                new Promise((resolve) => {
                    resolvePlaylistQueue = resolve;
                })
        );
        mockGetMusicQueueByArtist.mockResolvedValueOnce(emptyQueue);
        mockGetMusicQueueByAlbum.mockResolvedValueOnce(emptyQueue);

        const { result } = renderHook(() => useMusicHomeScreen());

        expect(result.current.isLoadingPlaylists).toBe(true);
        expect(result.current.featuredPlaylists).toEqual([]);
        expect(result.current.nextTracks).toEqual([]);

        let playlistPromise: Promise<void> | undefined;
        await act(async () => {
            playlistPromise = result.current.playPlaylist(5, 'Mix');
        });
        await waitFor(() => {
            expect(result.current.isActionPending('playlist-5')).toBe(true);
        });
        await act(async () => {
            resolvePlaylistQueue?.(emptyQueue);
            await playlistPromise;
        });
        expect(replaceQueue).not.toHaveBeenCalled();
        expect(result.current.isActionPending('playlist-5')).toBe(false);

        await act(async () => {
            await result.current.playArtist({
                key: 'artist-a',
                artist: 'Artist A',
                track_count: 1,
                album_count: 1,
            });
            await result.current.playAlbum({
                key: 'artist-a::album-a',
                album: 'Album A',
                artist: 'Artist A',
                year: '2024',
                track_count: 1,
            });
        });
        expect(replaceQueue).not.toHaveBeenCalled();
    });

    it('starts playback with the complete queue returned for each home card', async () => {
        const replaceQueue = jest.fn();
        mockUseGlobalMusic.mockReturnValue({
            currentIndex: undefined,
            currentTrack: undefined,
            hasQueue: false,
            playbackContext: undefined,
            queue: [],
            replaceQueue,
            toggleQueue: jest.fn(),
        });
        const manyTrackIds = Array.from({ length: 500 }, (_, index) => index + 1);
        mockGetPlaylistQueue.mockResolvedValueOnce(queueOf(...manyTrackIds));
        mockGetMusicQueueByArtist.mockResolvedValueOnce(queueOf(...manyTrackIds));
        mockGetMusicQueueByAlbum.mockResolvedValueOnce(queueOf(...manyTrackIds));

        const { result } = renderHook(() => useMusicHomeScreen());

        await act(async () => {
            await result.current.playPlaylist(5, 'Mix');
            await result.current.playArtist({
                key: 'artist-a',
                artist: 'Artist A',
                track_count: 500,
                album_count: 1,
            });
            await result.current.playAlbum({
                key: 'album-a',
                album: 'Album A',
                artist: 'Artist A',
                year: '2024',
                track_count: 500,
            });
        });

        expect(mockGetPlaylistQueue).toHaveBeenCalledWith(5);
        expect(mockGetMusicQueueByArtist).toHaveBeenCalledWith('artist-a');
        expect(mockGetMusicQueueByAlbum).toHaveBeenCalledWith('album-a');
        expect(replaceQueue).toHaveBeenCalledTimes(3);
        for (const [tracks, startIndex] of replaceQueue.mock.calls) {
            expect(tracks).toHaveLength(500);
            expect(startIndex).toBe(0);
            expect(tracks[0]).toMatchObject({
                id: 1,
                name: 'song-1',
                metadata: { title: 'Song 1', artist: 'Artist A', album: 'Album A', length: 180 },
            });
        }
    });

    it('requests the artist and album highlights ordered by most recently added', async () => {
        renderHook(() => useMusicHomeScreen());

        const { queryFn } = mockUseQuery.mock.calls[0][0];
        await queryFn();

        expect(mockGetMusicHomeCatalog).toHaveBeenCalledWith(4, { sort: 'recent', order: 'desc' });
    });
});

import { act, renderHook, waitFor } from '@testing-library/react';
import { useMusicHomeScreen } from './useMusicHomeScreen';

const mockUseGlobalMusic = jest.fn();
const mockUseQuery = jest.fn();
const mockGetPlaylistTracks = jest.fn();
const mockGetMusicByArtist = jest.fn();
const mockGetMusicByAlbum = jest.fn();
const mockGetMusicHomeCatalog = jest.fn();

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
    getMusicTitle: (m: any) => m.name ?? m.metadata?.title ?? `title-${m.id}`,
    getMusicArtist: (m: any) => m.metadata?.artist ?? `artist-${m.id}`,
    musicMetadata: () => 'meta',
    formatMusicDuration: (s: number) =>
        `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`,
}));

jest.mock('@tanstack/react-query', () => ({
    useQuery: (...args: any[]) => mockUseQuery(...args),
}));

jest.mock('@/service/playlist', () => ({
    getPlaylistTracks: (...args: any[]) => mockGetPlaylistTracks(...args),
}));

jest.mock('@/service/music', () => ({
    getMusicByArtist: (...args: any[]) => mockGetMusicByArtist(...args),
    getMusicByAlbum: (...args: any[]) => mockGetMusicByAlbum(...args),
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
        mockGetPlaylistTracks.mockResolvedValue({
            items: [{ file: { id: 10 } }],
        });
        mockGetMusicByArtist.mockResolvedValue({
            items: [{ id: 20 }],
        });
        mockGetMusicByAlbum.mockResolvedValue({
            items: [{ id: 30 }],
        });
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
        let resolvePlaylistTracks: ((value: { items: never[] }) => void) | undefined;
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
        mockGetPlaylistTracks.mockImplementationOnce(
            () =>
                new Promise((resolve) => {
                    resolvePlaylistTracks = resolve;
                })
        );
        mockGetMusicByArtist.mockResolvedValueOnce({ items: [] });
        mockGetMusicByAlbum.mockResolvedValueOnce({ items: [] });

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
            resolvePlaylistTracks?.({ items: [] });
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
});

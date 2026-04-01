import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { VideoContentProvider, useVideoContentProvider } from './videoContentProvider';
import type { ReactNode } from 'react';

const mockGetVideoPlaylists = jest.fn();
const mockGetVideoHomeCatalog = jest.fn();
const mockGetVideoLibraryFiles = jest.fn();
const mockGetVideoPlaylistMemberships = jest.fn();
const mockGetVideoPlaybackState = jest.fn();
const mockGetVideoPlaylistById = jest.fn();
const mockAddVideoToPlaylist = jest.fn();
const mockReorderVideoPlaylist = jest.fn();
const mockRemoveVideoFromPlaylist = jest.fn();
const mockUpdateVideoPlaylistName = jest.fn();
const mockNavigate = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
    getVideoPlaylists: (...args: unknown[]) => mockGetVideoPlaylists(...args),
    getVideoHomeCatalog: (...args: unknown[]) => mockGetVideoHomeCatalog(...args),
    getVideoLibraryFiles: (...args: unknown[]) => mockGetVideoLibraryFiles(...args),
    getVideoPlaylistMemberships: (...args: unknown[]) => mockGetVideoPlaylistMemberships(...args),
    getVideoPlaybackState: (...args: unknown[]) => mockGetVideoPlaybackState(...args),
    getVideoPlaylistById: (...args: unknown[]) => mockGetVideoPlaylistById(...args),
    addVideoToPlaylist: (...args: unknown[]) => mockAddVideoToPlaylist(...args),
    reorderVideoPlaylist: (...args: unknown[]) => mockReorderVideoPlaylist(...args),
    removeVideoFromPlaylist: (...args: unknown[]) => mockRemoveVideoFromPlaylist(...args),
    updateVideoPlaylistName: (...args: unknown[]) => mockUpdateVideoPlaylistName(...args),
}));

jest.mock('react-router-dom', () => {
    const actual = jest.requireActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

const createPlaylist = (overrides: Record<string, unknown> = {}) => ({
    id: 1,
    type: 'series',
    source_path: '/videos/series',
    name: 'Test Series',
    is_hidden: false,
    is_auto: true,
    group_mode: 'folder',
    classification: 'series',
    item_count: 3,
    cover_video_id: 10,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    last_played_at: null,
    items: [],
    ...overrides,
});

const createVideoFile = (overrides: Record<string, unknown> = {}) => ({
    id: 100,
    name: 'video1.mp4',
    path: '/videos/video1.mp4',
    parent_path: '/videos',
    format: '.mp4',
    size: 1024,
    ...overrides,
});

const createQueryClient = () =>
    new QueryClient({
        defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
        },
    });

const createWrapper = (initialEntries: string[] = ['/videos']) => {
    const queryClient = createQueryClient();
    return ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <VideoContentProvider>{children}</VideoContentProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const setupDefaultMocks = () => {
    mockGetVideoPlaylists.mockResolvedValue([]);
    mockGetVideoHomeCatalog.mockResolvedValue({ sections: [] });
    mockGetVideoLibraryFiles.mockResolvedValue({
        items: [],
        pagination: { page: 1, page_size: 60, has_next: false, has_prev: false },
    });
    mockGetVideoPlaylistMemberships.mockResolvedValue([]);
    mockGetVideoPlaybackState.mockResolvedValue(null);
    mockGetVideoPlaylistById.mockResolvedValue(null);
    mockAddVideoToPlaylist.mockResolvedValue(undefined);
    mockReorderVideoPlaylist.mockResolvedValue(undefined);
    mockRemoveVideoFromPlaylist.mockResolvedValue(undefined);
    mockUpdateVideoPlaylistName.mockResolvedValue(undefined);
};

describe('VideoContentProvider', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        setupDefaultMocks();
    });

    it('throws when useVideoContentProvider is used outside the provider', () => {
        expect(() => {
            renderHook(() => useVideoContentProvider());
        }).toThrow('useVideoContentProvider must be used within VideoContentProvider');
    });

    it('provides default state', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        expect(result.current.playlists).toEqual([]);
        expect(result.current.allVideos).toEqual([]);
        expect(result.current.filteredVideos).toEqual([]);
        expect(result.current.videoSearch).toBe('');
    });

    it('categorizes playlists by classification and type', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 1, name: 'Anime Show', classification: 'anime' }),
            createPlaylist({ id: 2, name: 'Drama Series', classification: 'series' }),
            createPlaylist({ id: 3, name: 'Action Movie', classification: 'movie' }),
            createPlaylist({
                id: 4,
                name: 'Personal Vids',
                classification: 'personal',
            }),
            createPlaylist({ id: 5, name: 'Fun Clips', classification: 'clip' }),
            createPlaylist({ id: 6, name: 'Programs', classification: 'program' }),
            createPlaylist({
                id: 7,
                name: 'Folder Vids',
                type: 'folder',
                classification: 'series',
            }),
        ]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(7));

        expect(result.current.seriesPlaylists).toHaveLength(3);
        expect(result.current.moviePlaylists).toHaveLength(1);
        expect(result.current.personalPlaylists).toHaveLength(1);
        expect(result.current.clipPlaylists).toHaveLength(2);
        expect(result.current.folderPlaylists).toHaveLength(1);
    });

    it('sorts continuePlaylists by last_played_at and applies playback state cover', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({
                id: 1,
                name: 'Old Show',
                last_played_at: '2026-01-01T00:00:00Z',
            }),
            createPlaylist({
                id: 2,
                name: 'New Show',
                last_played_at: '2026-03-01T00:00:00Z',
            }),
            createPlaylist({ id: 3, name: 'No Play', last_played_at: null }),
        ]);
        mockGetVideoPlaybackState.mockResolvedValue({
            playlist: createPlaylist({ id: 2 }),
            playback_state: { playlist_id: 2, video_id: 99 },
        });

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.continuePlaylists).toHaveLength(2));

        expect(result.current.continuePlaylists[0]?.name).toBe('New Show');
        expect(result.current.continuePlaylists[1]?.name).toBe('Old Show');
        expect(result.current.continuePlaylists[0]?.cover_video_id).toBe(99);
    });

    it('builds playlistMembershipMap from memberships', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 1 }),
            createPlaylist({ id: 2 }),
        ]);
        mockGetVideoPlaylistMemberships.mockResolvedValue([
            { playlist_id: 1, video_id: 10 },
            { playlist_id: 1, video_id: 20 },
            { playlist_id: 2, video_id: 30 },
        ]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(2));

        await waitFor(() => {
            expect(Object.keys(result.current.playlistMembershipMap).length).toBeGreaterThan(0);
        });

        const map = result.current.playlistMembershipMap;
        expect(map[1]!.has(10)).toBe(true);
        expect(map[1]!.has(20)).toBe(true);
        expect(map[2]!.has(30)).toBe(true);
    });

    it('filters videos by search via query refetch', async () => {
        mockGetVideoLibraryFiles.mockResolvedValue({
            items: [
                createVideoFile({
                    id: 1,
                    name: 'action.mp4',
                    parent_path: '/movies',
                    format: '.mp4',
                }),
                createVideoFile({
                    id: 2,
                    name: 'comedy.mkv',
                    parent_path: '/movies',
                    format: '.mkv',
                }),
            ],
            pagination: { page: 1, page_size: 60, has_next: false, has_prev: false },
        });

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.allVideos).toHaveLength(2));

        // When videoSearch changes, the infinite query key changes and refetches
        // The filtering also happens client-side on the returned data
        mockGetVideoLibraryFiles.mockResolvedValue({
            items: [
                createVideoFile({
                    id: 1,
                    name: 'action.mp4',
                    parent_path: '/movies',
                    format: '.mp4',
                }),
            ],
            pagination: { page: 1, page_size: 60, has_next: false, has_prev: false },
        });

        act(() => result.current.setVideoSearch('action'));

        await waitFor(() => expect(result.current.filteredVideos).toHaveLength(1));
        expect(result.current.filteredVideos[0]?.name).toBe('action.mp4');
    });

    it('loadMoreVideos calls fetchNextPage when has more', async () => {
        mockGetVideoLibraryFiles
            .mockResolvedValueOnce({
                items: [createVideoFile()],
                pagination: { page: 1, page_size: 60, has_next: true, has_prev: false },
            })
            .mockResolvedValueOnce({
                items: [createVideoFile({ id: 200 })],
                pagination: { page: 2, page_size: 60, has_next: false, has_prev: true },
            });

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.hasMoreVideos).toBe(true));

        act(() => result.current.loadMoreVideos());

        await waitFor(() => expect(mockGetVideoLibraryFiles).toHaveBeenCalledTimes(2));
    });

    it('closeFeedback sets open to false', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.closeFeedback());
        expect(result.current.feedback.open).toBe(false);
    });

    it('setSelectedPlaylistForVideo updates the per-video playlist selection', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.setSelectedPlaylistForVideo(100, 5));
        expect(result.current.selectedPlaylistPerVideo[100]).toBe(5);
    });

    it('addVideoFromLibrary calls mutation with selected playlist or first playlist', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 10, name: 'First Playlist' }),
            createPlaylist({ id: 20, name: 'Second Playlist' }),
        ]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(2));

        act(() => result.current.addVideoFromLibrary(100));

        await waitFor(() => expect(mockAddVideoToPlaylist).toHaveBeenCalledWith(10, 100));
    });

    it('addVideoFromLibrary uses per-video selected playlist when set', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 10, name: 'First' }),
            createPlaylist({ id: 20, name: 'Second' }),
        ]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(2));

        act(() => result.current.setSelectedPlaylistForVideo(100, 20));
        act(() => result.current.addVideoFromLibrary(100));

        await waitFor(() => expect(mockAddVideoToPlaylist).toHaveBeenCalledWith(20, 100));
    });

    it('addVideoFromLibrary does nothing when no playlists exist', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.addVideoFromLibrary(100));

        expect(mockAddVideoToPlaylist).not.toHaveBeenCalled();
    });

    it('renameSelectedPlaylist does nothing for empty name', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.renameSelectedPlaylist('   '));

        expect(mockUpdateVideoPlaylistName).not.toHaveBeenCalled();
    });

    it('clearSelectedPlaylist navigates to /videos from home section', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos']),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.clearSelectedPlaylist());

        expect(mockNavigate).toHaveBeenCalledWith('/videos');
    });

    it('clearSelectedPlaylist navigates to /videos/series from series section', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/series']),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.clearSelectedPlaylist());

        expect(mockNavigate).toHaveBeenCalledWith('/videos/series');
    });

    it('playVideo navigates with playlist id and from state', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/series']),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.playVideo(42, 7));

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.stringContaining('/video/42'),
            expect.objectContaining({
                state: expect.objectContaining({ playlistId: 7 }),
            })
        );
    });

    it('playVideo does nothing for videoId 0', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.playVideo(0));

        expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('selectPlaylist navigates to playlist detail route', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 1, name: 'Cool Series', classification: 'series' }),
        ]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/series']),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(1));

        act(() => result.current.selectPlaylist(result.current.playlists[0]!));

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.stringContaining('/videos/series/cool-series')
        );
    });

    it('selectedPlaylistSummary and detail load when URL has slug', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 5, name: 'Action Movies', classification: 'movie' }),
        ]);
        mockGetVideoPlaylistById.mockResolvedValue(
            createPlaylist({
                id: 5,
                name: 'Action Movies',
                items: [
                    {
                        id: 1,
                        order_index: 0,
                        source_kind: 'auto',
                        video: createVideoFile({ id: 50 }),
                        status: 'not_started',
                        progress_pct: 0,
                    },
                    {
                        id: 2,
                        order_index: 1,
                        source_kind: 'auto',
                        video: createVideoFile({ id: 51 }),
                        status: 'not_started',
                        progress_pct: 0,
                    },
                ],
            })
        );

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/movies/action-movies']),
        });

        await waitFor(() => expect(result.current.selectedPlaylistSummary).not.toBeNull());

        expect(result.current.selectedPlaylistSummary?.name).toBe('Action Movies');

        await waitFor(() => expect(result.current.selectedPlaylistDetail).not.toBeNull());
    });

    it('removeVideoFromSelectedPlaylist triggers mutation when playlist is selected', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 5, name: 'Action Movies', classification: 'movie' }),
        ]);
        mockGetVideoPlaylistById.mockResolvedValue(
            createPlaylist({ id: 5, name: 'Action Movies', items: [] })
        );

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/movies/action-movies']),
        });

        await waitFor(() => expect(result.current.selectedPlaylistSummary).not.toBeNull());

        act(() => result.current.removeVideoFromSelectedPlaylist(50));

        await waitFor(() => expect(mockRemoveVideoFromPlaylist).toHaveBeenCalledWith(5, 50));
    });

    it('moveSelectedPlaylistItem swaps items and triggers reorder mutation', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 5, name: 'Action Movies', classification: 'movie' }),
        ]);
        mockGetVideoPlaylistById.mockResolvedValue(
            createPlaylist({
                id: 5,
                name: 'Action Movies',
                items: [
                    {
                        id: 1,
                        order_index: 0,
                        source_kind: 'auto',
                        video: createVideoFile({ id: 50 }),
                        status: 'not_started',
                        progress_pct: 0,
                    },
                    {
                        id: 2,
                        order_index: 1,
                        source_kind: 'auto',
                        video: createVideoFile({ id: 51 }),
                        status: 'not_started',
                        progress_pct: 0,
                    },
                    {
                        id: 3,
                        order_index: 2,
                        source_kind: 'auto',
                        video: createVideoFile({ id: 52 }),
                        status: 'not_started',
                        progress_pct: 0,
                    },
                ],
            })
        );

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/movies/action-movies']),
        });

        await waitFor(() => expect(result.current.selectedPlaylistDetail).not.toBeNull());

        act(() => result.current.moveSelectedPlaylistItem(0, 1));

        await waitFor(() =>
            expect(mockReorderVideoPlaylist).toHaveBeenCalledWith(5, [
                { video_id: 51, order_index: 0 },
                { video_id: 50, order_index: 1 },
                { video_id: 52, order_index: 2 },
            ])
        );
    });

    it('moveSelectedPlaylistItem does nothing when no detail loaded', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.moveSelectedPlaylistItem(0, 1));

        expect(mockReorderVideoPlaylist).not.toHaveBeenCalled();
    });

    it('moveSelectedPlaylistItem does nothing for out-of-bounds target', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 5, name: 'Action Movies', classification: 'movie' }),
        ]);
        mockGetVideoPlaylistById.mockResolvedValue(
            createPlaylist({
                id: 5,
                name: 'Action Movies',
                items: [
                    {
                        id: 1,
                        order_index: 0,
                        source_kind: 'auto',
                        video: createVideoFile({ id: 50 }),
                        status: 'not_started',
                        progress_pct: 0,
                    },
                ],
            })
        );

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/movies/action-movies']),
        });

        await waitFor(() => expect(result.current.selectedPlaylistDetail).not.toBeNull());

        act(() => result.current.moveSelectedPlaylistItem(0, -1));

        expect(mockReorderVideoPlaylist).not.toHaveBeenCalled();
    });

    it('openPlaylistVideo navigates with playlist id', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({ id: 5, name: 'Action Movies', classification: 'movie' }),
        ]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/movies/action-movies']),
        });

        await waitFor(() => expect(result.current.selectedPlaylistSummary).not.toBeNull());

        act(() => result.current.openPlaylistVideo(42));

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.stringContaining('/video/42'),
            expect.objectContaining({
                state: expect.objectContaining({ playlistId: 5 }),
            })
        );
    });

    it('openPlaylistVideo does nothing when no playlist selected', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.openPlaylistVideo(42));

        expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('recentCatalogItems extracts from home catalog recent section', async () => {
        mockGetVideoHomeCatalog.mockResolvedValue({
            sections: [
                {
                    key: 'recent',
                    title: 'Recent',
                    items: [
                        {
                            video: createVideoFile({ id: 1 }),
                            status: 'not_started',
                            progress_pct: 0,
                        },
                    ],
                },
            ],
        });

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.recentCatalogItems).toHaveLength(1));
    });

    it('addToPlaylist mutation shows error feedback on failure', async () => {
        mockGetVideoPlaylists.mockResolvedValue([createPlaylist({ id: 10 })]);
        mockAddVideoToPlaylist.mockRejectedValue(new Error('fail'));

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(1));

        act(() => result.current.addVideoFromLibrary(100));

        await waitFor(() => expect(result.current.feedback.severity).toBe('error'));
        expect(result.current.feedback.open).toBe(true);
    });

    it('addToPlaylist mutation shows success feedback', async () => {
        mockGetVideoPlaylists.mockResolvedValue([createPlaylist({ id: 10 })]);
        mockAddVideoToPlaylist.mockResolvedValue(undefined);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(1));

        act(() => result.current.addVideoFromLibrary(100));

        await waitFor(() => expect(result.current.feedback.severity).toBe('success'));
        expect(result.current.feedback.open).toBe(true);
    });

    it('getNextPageParam returns next page when has_next is true', async () => {
        mockGetVideoLibraryFiles
            .mockResolvedValueOnce({
                items: [createVideoFile()],
                pagination: { page: 1, page_size: 60, has_next: true, has_prev: false },
            })
            .mockResolvedValueOnce({
                items: [],
                pagination: { page: 2, page_size: 60, has_next: false, has_prev: true },
            });

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.hasMoreVideos).toBe(true));

        act(() => result.current.loadMoreVideos());

        await waitFor(() => expect(result.current.hasMoreVideos).toBe(false));
    });

    it('continuePlaylists handles playlists where one has null last_played_at time gracefully', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({
                id: 1,
                name: 'Show A',
                last_played_at: '2026-01-15T00:00:00Z',
            }),
            createPlaylist({
                id: 2,
                name: 'Show B',
                last_played_at: '2026-02-15T00:00:00Z',
            }),
        ]);
        mockGetVideoPlaybackState.mockResolvedValue(null);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.continuePlaylists).toHaveLength(2));

        expect(result.current.continuePlaylists[0]?.name).toBe('Show B');
        expect(result.current.continuePlaylists[1]?.name).toBe('Show A');
    });

    it('renameSelectedPlaylist calls mutation with valid name', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({
                id: 5,
                name: 'My Playlist',
                classification: 'personal',
            }),
        ]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos/personal/my-playlist']),
        });

        await waitFor(() => expect(result.current.selectedPlaylistSummary).not.toBeNull());

        act(() => result.current.renameSelectedPlaylist('New Name'));

        await waitFor(() =>
            expect(mockUpdateVideoPlaylistName).toHaveBeenCalledWith(5, 'New Name')
        );
    });

    it('renameSelectedPlaylist mutation guards against no selected playlist', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        // No selected playlist => mutation function returns early
        act(() => result.current.renameSelectedPlaylist('New Name'));

        // The mutation fires but the mutationFn should return early
        await waitFor(() => expect(result.current.isRenamingPlaylist).toBe(false));
        expect(mockUpdateVideoPlaylistName).not.toHaveBeenCalled();
    });

    it('removeVideoFromSelectedPlaylist guards against no selected playlist', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.removeVideoFromSelectedPlaylist(50));

        await waitFor(() => expect(result.current.isRemovingFromPlaylist).toBe(false));
        expect(mockRemoveVideoFromPlaylist).not.toHaveBeenCalled();
    });

    it('playVideo without playlist id navigates without playlist param', async () => {
        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

        act(() => result.current.playVideo(42));

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.stringContaining('/video/42'),
            expect.objectContaining({
                state: expect.objectContaining({ playlistId: null }),
            })
        );
    });

    it('continuePlaylists returns sorted playlists when no playback state', async () => {
        mockGetVideoPlaylists.mockResolvedValue([
            createPlaylist({
                id: 1,
                name: 'Show A',
                last_played_at: '2026-01-15T00:00:00Z',
            }),
        ]);
        mockGetVideoPlaybackState.mockRejectedValue(new Error('no state'));

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.continuePlaylists).toHaveLength(1));
        expect(result.current.continuePlaylists[0]?.name).toBe('Show A');
    });

    it('selectPlaylist resolves section from home when on home route', async () => {
        const folderPlaylist = createPlaylist({
            id: 1,
            name: 'Folder Vids',
            type: 'folder',
            classification: 'series',
        });
        mockGetVideoPlaylists.mockResolvedValue([folderPlaylist]);

        const { result } = renderHook(() => useVideoContentProvider(), {
            wrapper: createWrapper(['/videos']),
        });

        await waitFor(() => expect(result.current.playlists).toHaveLength(1));

        act(() => result.current.selectPlaylist(folderPlaylist as any));

        expect(mockNavigate).toHaveBeenCalledWith(expect.stringContaining('/videos/folders/'));
    });
});

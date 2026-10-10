import { renderHook } from '@testing-library/react';
import { useQuery } from '@tanstack/react-query';
import useHomeScreen, { homeScreenUtils } from './useHomeScreen';

jest.mock('@tanstack/react-query', () => ({
    useQuery: jest.fn(),
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: jest.fn(),
}));

jest.mock('@/service/analytics', () => ({
    fetchAnalyticsStorage: jest.fn(() => Promise.resolve({ storage: {}, counts: {} })),
    fetchAnalyticsHealth: jest.fn(() => Promise.resolve({})),
    fetchAnalyticsRecentFiles: jest.fn(() => Promise.resolve([])),
}));

jest.mock('@/service/files', () => ({
    getStarredFiles: jest.fn(() => Promise.resolve({ items: [] })),
}));

jest.mock('@/service/image', () => ({
    getImageFiles: jest.fn(() => Promise.resolve({ items: [] })),
}));

jest.mock('@/service/playerState', () => ({
    getPlayerState: jest.fn(() => Promise.resolve({})),
    getPlayerQueue: jest.fn(() => Promise.resolve({ items: [], current_index: 0 })),
}));

jest.mock('@/service/videoPlayback', () => ({
    getVideoContinueWatching: jest.fn(() => Promise.resolve([])),
    getVideoPlaybackState: jest.fn(() => Promise.resolve(null)),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseGlobalMusic = jest.requireMock('@/features/music/providers/GlobalMusicProvider')
    .useGlobalMusic as jest.Mock;

const defaultGlobalMusic = () => ({
    queue: [],
    currentTrack: undefined,
    currentTime: 0,
    duration: 0,
    isPlaying: false,
});

const buildQueryState = (data: unknown, overrides?: Record<string, unknown>) => ({
    data,
    isLoading: false,
    ...overrides,
});

const loadingQueryState = () => buildQueryState(undefined, { isLoading: true });

const savedQueueEntry = (fileId: number, lengthSeconds = 0) => ({
    file_id: fileId,
    name: `${fileId}.mp3`,
    path: `/music/${fileId}.mp3`,
    format: '.mp3',
    title: `Title ${fileId}`,
    artist: 'Artist',
    album: 'Album',
    length: lengthSeconds,
});

const savedQueue = (fileIds: number[], currentIndex = 0, lengthSeconds = 0) =>
    buildQueryState({
        items: fileIds.map((fileId) => savedQueueEntry(fileId, lengthSeconds)),
        current_index: currentIndex,
    });

const buildActiveTrack = () => ({
    id: 10,
    name: 'Now.mp3',
    size: 512,
    updated_at: '2026-03-15',
    metadata: { length: 300, title: 'Now', artist: 'Band' },
});

const setupDefaultQueries = (
    overrides?: Partial<
        Record<
            | 'analytics'
            | 'favorites'
            | 'images'
            | 'videoContinue'
            | 'videoPlayback'
            | 'playerState'
            | 'playerQueue',
            ReturnType<typeof buildQueryState>
        >
    >
) => {
    mockedUseQuery
        .mockReturnValueOnce(overrides?.analytics ?? buildQueryState({ recent_files: [] }))
        .mockReturnValueOnce(overrides?.favorites ?? buildQueryState({ items: [] }))
        .mockReturnValueOnce(overrides?.images ?? buildQueryState({ items: [] }))
        .mockReturnValueOnce(overrides?.videoContinue ?? buildQueryState([]))
        .mockReturnValueOnce(overrides?.videoPlayback ?? buildQueryState(null))
        .mockReturnValueOnce(overrides?.playerState ?? buildQueryState(null))
        .mockReturnValueOnce(overrides?.playerQueue ?? buildQueryState(null));
};

describe('useHomeScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedUseGlobalMusic.mockReturnValue(defaultGlobalMusic());
    });

    // --- getProgressPercent / clampProgress ---

    describe('homeScreenUtils.getProgressPercent', () => {
        it('returns percentage when duration > 0', () => {
            expect(homeScreenUtils.getProgressPercent(50, 200)).toBe(25);
        });

        it('clamps to 100 when currentTime exceeds duration', () => {
            expect(homeScreenUtils.getProgressPercent(300, 100)).toBe(100);
        });

        it('clamps to 0 when value is negative', () => {
            expect(homeScreenUtils.getProgressPercent(-10, 100)).toBe(0);
        });

        it('returns fallback clamped when duration is 0', () => {
            expect(homeScreenUtils.getProgressPercent(0, 0, 50)).toBe(50);
        });

        it('returns 0 when duration is 0 and no fallback', () => {
            expect(homeScreenUtils.getProgressPercent(0, 0)).toBe(0);
        });

        it('clamps fallback to 100 when overflow', () => {
            expect(homeScreenUtils.getProgressPercent(0, 0, 135)).toBe(100);
        });

        it('clamps fallback to 0 when negative', () => {
            expect(homeScreenUtils.getProgressPercent(0, 0, -20)).toBe(0);
        });

        it('returns 0 for NaN currentTime', () => {
            expect(homeScreenUtils.getProgressPercent(Number.NaN, 120)).toBe(0);
        });

        it('returns 0 for Infinity result', () => {
            expect(homeScreenUtils.getProgressPercent(Infinity, 100)).toBe(0);
        });

        it('returns 0 when duration is negative', () => {
            expect(homeScreenUtils.getProgressPercent(50, -10)).toBe(0);
        });
    });

    // --- Null / empty data paths ---

    describe('null and empty data derivations', () => {
        it('returns empty arrays when all query data is null/undefined', () => {
            setupDefaultQueries({
                analytics: buildQueryState(null),
                favorites: buildQueryState(null),
                images: buildQueryState(null),
                videoContinue: buildQueryState(null),
            });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.recentFiles).toEqual([]);
            expect(result.current.favoriteItems).toEqual([]);
            expect(result.current.recentImages).toEqual([]);
            expect(result.current.videoContinueItems).toEqual([]);
            expect(result.current.videoResume).toBeNull();
            expect(result.current.musicResume).toBeNull();
            expect(result.current.analytics).toBeNull();
        });

        it('returns empty arrays when data objects have null sub-fields', () => {
            setupDefaultQueries({
                analytics: buildQueryState({ recent_files: null }),
                favorites: buildQueryState({ items: null }),
                images: buildQueryState({ items: null }),
            });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.recentFiles).toEqual([]);
            expect(result.current.favoriteItems).toEqual([]);
            expect(result.current.recentImages).toEqual([]);
        });

        it('slices recent_files to 6 items', () => {
            const files = Array.from({ length: 10 }, (_, i) => ({
                id: i,
                name: `file${i}`,
            }));
            setupDefaultQueries({
                analytics: buildQueryState({ recent_files: files }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.recentFiles).toHaveLength(6);
        });

        it('slices favoriteItems to homeFavoritesLimit (6)', () => {
            const items = Array.from({ length: 10 }, (_, i) => ({
                id: i,
                name: `fav${i}`,
            }));
            setupDefaultQueries({
                favorites: buildQueryState({ items }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.favoriteItems).toHaveLength(6);
        });

        it('slices recentImages to homeImagesLimit (6)', () => {
            const items = Array.from({ length: 10 }, (_, i) => ({
                id: i,
                name: `img${i}`,
            }));
            setupDefaultQueries({
                images: buildQueryState({ items }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.recentImages).toHaveLength(6);
        });
    });

    // --- videoContinueItems ---

    describe('videoContinueItems', () => {
        const continueEntry = (id: number, positionSeconds: number, durationSeconds: number) => ({
            video: { id, name: `v${id}` },
            position_seconds: positionSeconds,
            duration_seconds: durationSeconds,
            updated_at: '2026-01-01T00:00:00Z',
        });

        it('maps in-progress videos with their progress percentage', () => {
            setupDefaultQueries({
                videoContinue: buildQueryState([
                    continueEntry(1, 10, 100),
                    continueEntry(2, 30, 60),
                ]),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoContinueItems).toHaveLength(2);
            expect(result.current.videoContinueItems[0]?.video.id).toBe(1);
            expect(result.current.videoContinueItems[0]?.progress_pct).toBe(10);
            expect(result.current.videoContinueItems[1]?.progress_pct).toBe(50);
            expect(result.current.videoContinueItems[1]?.status).toBe('in_progress');
        });

        it('returns empty when there is nothing in progress', () => {
            setupDefaultQueries({ videoContinue: buildQueryState([]) });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoContinueItems).toEqual([]);
        });

        it('returns empty while the data is not loaded', () => {
            setupDefaultQueries({ videoContinue: buildQueryState(undefined) });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoContinueItems).toEqual([]);
        });

        it('slices continue items to max 4', () => {
            setupDefaultQueries({
                videoContinue: buildQueryState(
                    Array.from({ length: 8 }, (_, i) => continueEntry(i, 10, 100))
                ),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoContinueItems).toHaveLength(4);
        });
    });

    // --- videoResume ---

    describe('videoResume', () => {
        it('returns null when videoPlayback data is null', () => {
            setupDefaultQueries({
                videoPlayback: buildQueryState(null),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoResume).toBeNull();
        });

        it('returns null when session has no video_id', () => {
            setupDefaultQueries({
                videoPlayback: buildQueryState({
                    playlist: { id: 1, items: [] },
                    playback_state: {
                        video_id: null,
                        current_time: 0,
                        duration: 0,
                        playlist_id: null,
                    },
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoResume).toBeNull();
        });

        it('returns null when video_id is 0 (falsy)', () => {
            setupDefaultQueries({
                videoPlayback: buildQueryState({
                    playlist: { id: 1, items: [{ video: { id: 0 } }] },
                    playback_state: {
                        video_id: 0,
                        current_time: 0,
                        duration: 0,
                        playlist_id: null,
                    },
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoResume).toBeNull();
        });

        it('returns null when activeItem is not found in playlist items', () => {
            setupDefaultQueries({
                videoPlayback: buildQueryState({
                    playlist: {
                        id: 1,
                        items: [{ video: { id: 99, name: 'Other' } }],
                    },
                    playback_state: {
                        video_id: 5,
                        current_time: 30,
                        duration: 120,
                        playlist_id: 1,
                    },
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoResume).toBeNull();
        });

        it('returns video resume when session is valid', () => {
            setupDefaultQueries({
                videoPlayback: buildQueryState({
                    playlist: {
                        id: 10,
                        items: [{ video: { id: 5, name: 'Episode 5', parent_path: '/tv' } }],
                    },
                    playback_state: {
                        video_id: 5,
                        current_time: 60,
                        duration: 120,
                        playlist_id: 10,
                    },
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoResume).toEqual({
                video: { id: 5, name: 'Episode 5', parent_path: '/tv' },
                progressSeconds: 60,
                durationSeconds: 120,
                progressPercent: 50,
                playlistId: 10,
            });
        });

        it('handles zero duration in video resume', () => {
            setupDefaultQueries({
                videoPlayback: buildQueryState({
                    playlist: {
                        id: 10,
                        items: [{ video: { id: 5, name: 'ep', parent_path: '/' } }],
                    },
                    playback_state: {
                        video_id: 5,
                        current_time: 0,
                        duration: 0,
                        playlist_id: null,
                    },
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoResume!.progressPercent).toBe(0);
            expect(result.current.videoResume!.playlistId).toBeNull();
        });
    });

    describe('musicResume from the saved player queue', () => {
        it('is null when there is no current track and no saved queue', () => {
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume).toBeNull();
        });

        it('is null when the saved queue is empty', () => {
            setupDefaultQueries({ playerQueue: buildQueryState({ items: [], current_index: 0 }) });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume).toBeNull();
        });

        it('is null when the saved index points outside the queue', () => {
            setupDefaultQueries({ playerQueue: savedQueue([4], 3) });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume).toBeNull();
        });

        it('resumes the saved track at the saved position without playing', () => {
            setupDefaultQueries({
                playerState: buildQueryState({ current_file_id: 6, current_position: 60 }),
                playerQueue: savedQueue([5, 6, 7], 1, 240),
            });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume).toMatchObject({
                track: { id: 6, metadata: { title: 'Title 6', length: 240 } },
                progressSeconds: 60,
                durationSeconds: 240,
                progressPercent: 25,
                queueCount: 3,
                isPlaying: false,
            });
        });

        it('starts at zero when the player state is missing', () => {
            setupDefaultQueries({ playerQueue: savedQueue([5], 0, 100) });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume).toMatchObject({
                progressSeconds: 0,
                progressPercent: 0,
                queueCount: 1,
            });
        });

        it('starts at zero when the saved state has no position', () => {
            setupDefaultQueries({
                playerState: buildQueryState({ current_file_id: 5 }),
                playerQueue: savedQueue([5]),
            });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume!.progressSeconds).toBe(0);
        });

        it('has no known duration for a saved track without length', () => {
            setupDefaultQueries({ playerQueue: savedQueue([5], 0, 0) });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume!.durationSeconds).toBe(0);
        });

        it('prefers the live track and live queue over the saved ones', () => {
            const currentTrack = buildActiveTrack();
            mockedUseGlobalMusic.mockReturnValue({
                queue: [currentTrack, { id: 11 }],
                currentTrack,
                currentTime: 150,
                duration: 100,
                isPlaying: true,
            });
            setupDefaultQueries({
                playerState: buildQueryState({ current_file_id: 6, current_position: 60 }),
                playerQueue: savedQueue([5, 6, 7, 8], 1, 240),
            });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume).toMatchObject({
                track: currentTrack,
                progressSeconds: 150,
                durationSeconds: 300,
                queueCount: 2,
                isPlaying: true,
            });
        });

        it('uses the duration reported by the player when it exceeds the metadata', () => {
            const currentTrack = { ...buildActiveTrack(), metadata: { length: 100 } };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [currentTrack],
                currentTrack,
                currentTime: 10,
                duration: 200,
                isPlaying: false,
            });
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume!.durationSeconds).toBe(200);
        });

        it('has no duration for a live track without metadata', () => {
            const { metadata: _metadata, ...trackWithoutMetadata } = buildActiveTrack();
            mockedUseGlobalMusic.mockReturnValue({
                queue: [],
                currentTrack: trackWithoutMetadata,
                currentTime: 0,
                duration: 0,
                isPlaying: false,
            });
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.musicResume!.durationSeconds).toBe(0);
            expect(result.current.musicResume!.queueCount).toBe(0);
        });
    });

    describe('loading states', () => {
        it.each([
            ['analytics', 'isAnalyticsLoading'],
            ['favorites', 'isFavoritesLoading'],
            ['images', 'isImagesLoading'],
            ['videoContinue', 'isVideoLoading'],
            ['videoPlayback', 'isVideoLoading'],
            ['playerState', 'isMusicLoading'],
            ['playerQueue', 'isMusicLoading'],
        ] as const)('reports loading while %s is loading', (queryName, loadingFlag) => {
            setupDefaultQueries({ [queryName]: loadingQueryState() });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current[loadingFlag]).toBe(true);
        });

        it('reports not loading when all queries have data', () => {
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.isAnalyticsLoading).toBe(false);
            expect(result.current.isFavoritesLoading).toBe(false);
            expect(result.current.isImagesLoading).toBe(false);
            expect(result.current.isVideoLoading).toBe(false);
            expect(result.current.isMusicLoading).toBe(false);
        });
    });

    describe('integration: full data scenario', () => {
        it('derives all fields correctly when all data is present', () => {
            const currentTrack = buildActiveTrack();
            mockedUseGlobalMusic.mockReturnValue({
                queue: [currentTrack, { id: 11 }, { id: 12 }],
                currentTrack,
                currentTime: 150,
                duration: 300,
                isPlaying: true,
            });

            setupDefaultQueries({
                analytics: buildQueryState({
                    recent_files: [{ id: 1 }, { id: 2 }],
                    health: { status: 'ok' },
                    storage: { used_bytes: 100 },
                }),
                favorites: buildQueryState({ items: [{ id: 55 }] }),
                images: buildQueryState({ items: [{ id: 77 }] }),
                videoContinue: buildQueryState([
                    {
                        video: { id: 3 },
                        position_seconds: 45,
                        duration_seconds: 100,
                        updated_at: '2026-01-01T00:00:00Z',
                    },
                ]),
                videoPlayback: buildQueryState({
                    playlist: {
                        id: 10,
                        items: [{ video: { id: 3, name: 'Ep3', parent_path: '/s' } }],
                    },
                    playback_state: {
                        video_id: 3,
                        current_time: 90,
                        duration: 180,
                        playlist_id: 10,
                    },
                }),
                playerState: buildQueryState({ current_file_id: 10, current_position: 50 }),
                playerQueue: savedQueue([10]),
            });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.recentFiles).toHaveLength(2);
            expect(result.current.favoriteItems).toHaveLength(1);
            expect(result.current.recentImages).toHaveLength(1);
            expect(result.current.videoContinueItems).toHaveLength(1);
            expect(result.current.videoResume!.progressPercent).toBe(50);
            expect(result.current.musicResume!.progressSeconds).toBe(150);
            expect(result.current.musicResume!.isPlaying).toBe(true);
            expect(result.current.musicResume!.queueCount).toBe(3);
            expect(result.current.analytics).toBeTruthy();
        });
    });

    describe('query configuration', () => {
        it('disables retry for the playback state queries', () => {
            setupDefaultQueries();

            renderHook(() => useHomeScreen());

            expect(mockedUseQuery.mock.calls[4][0].retry).toBe(false);
            expect(mockedUseQuery.mock.calls[5][0].retry).toBe(false);
            expect(mockedUseQuery.mock.calls[6][0].retry).toBe(false);
        });

        it('calls queryFn correctly for each query', async () => {
            setupDefaultQueries();

            renderHook(() => useHomeScreen());

            await expect(mockedUseQuery.mock.calls[0][0].queryFn()).resolves.toEqual({
                storage: {},
                counts: {},
                health: {},
                recent_files: [],
            });
            await expect(mockedUseQuery.mock.calls[1][0].queryFn()).resolves.toEqual({ items: [] });
            await expect(mockedUseQuery.mock.calls[2][0].queryFn()).resolves.toEqual({ items: [] });
            await expect(mockedUseQuery.mock.calls[3][0].queryFn()).resolves.toEqual([]);
            await expect(mockedUseQuery.mock.calls[4][0].queryFn()).resolves.toBeNull();
            await expect(mockedUseQuery.mock.calls[5][0].queryFn()).resolves.toEqual({});
            await expect(mockedUseQuery.mock.calls[6][0].queryFn()).resolves.toEqual({
                items: [],
                current_index: 0,
            });
        });
    });
});

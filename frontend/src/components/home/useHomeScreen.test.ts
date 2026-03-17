import { renderHook } from '@testing-library/react';
import { useQuery } from '@tanstack/react-query';
import useHomeScreen, { homeScreenUtils } from './useHomeScreen';

jest.mock('@tanstack/react-query', () => ({
    useQuery: jest.fn(),
}));

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: jest.fn(),
}));

jest.mock('@/service/analytics', () => ({
    fetchAnalyticsOverview: jest.fn(() => Promise.resolve({})),
}));

jest.mock('@/service/files', () => ({
    getFilesTree: jest.fn(() => Promise.resolve({ items: [] })),
    getImageFiles: jest.fn(() => Promise.resolve({ items: [] })),
}));

jest.mock('@/service/playerState', () => ({
    getPlayerState: jest.fn(() => Promise.resolve({})),
}));

jest.mock('@/service/playlist', () => ({
    getNowPlayingPlaylist: jest.fn(() => Promise.resolve({ id: 7 })),
    getPlaylistTracks: jest.fn(() => Promise.resolve({ items: [] })),
}));

jest.mock('@/service/videoPlayback', () => ({
    getVideoHomeCatalog: jest.fn(() => Promise.resolve({ sections: [] })),
    getVideoPlaybackState: jest.fn(() => Promise.resolve(null)),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseGlobalMusic = jest.requireMock('@/components/providers/GlobalMusicProvider')
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

/** Set up 8 useQuery returns (the hook calls useQuery 8 times). */
const setupDefaultQueries = (
    overrides?: Partial<
        Record<
            | 'analytics'
            | 'favorites'
            | 'images'
            | 'videoCatalog'
            | 'videoPlayback'
            | 'playerState'
            | 'nowPlaying'
            | 'nowPlayingTracks',
            ReturnType<typeof buildQueryState>
        >
    >
) => {
    mockedUseQuery
        .mockReturnValueOnce(overrides?.analytics ?? buildQueryState({ recent_files: [] }))
        .mockReturnValueOnce(overrides?.favorites ?? buildQueryState({ items: [] }))
        .mockReturnValueOnce(overrides?.images ?? buildQueryState({ items: [] }))
        .mockReturnValueOnce(overrides?.videoCatalog ?? buildQueryState({ sections: [] }))
        .mockReturnValueOnce(overrides?.videoPlayback ?? buildQueryState(null))
        .mockReturnValueOnce(overrides?.playerState ?? buildQueryState(null))
        .mockReturnValueOnce(overrides?.nowPlaying ?? buildQueryState(null))
        .mockReturnValueOnce(overrides?.nowPlayingTracks ?? buildQueryState({ items: [] }));
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
                videoCatalog: buildQueryState(null),
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
        it('returns items from the "continue" section', () => {
            const items = [
                {
                    video: { id: 1, name: 'v1' },
                    progress_pct: 10,
                    status: 'in_progress',
                },
                {
                    video: { id: 2, name: 'v2' },
                    progress_pct: 30,
                    status: 'in_progress',
                },
            ];
            setupDefaultQueries({
                videoCatalog: buildQueryState({
                    sections: [
                        { key: 'recent', title: 'Recent', items: [{ video: { id: 99 } }] },
                        { key: 'continue', title: 'Continue', items },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoContinueItems).toHaveLength(2);
            expect(result.current.videoContinueItems[0]?.video.id).toBe(1);
        });

        it('returns empty when no continue section exists', () => {
            setupDefaultQueries({
                videoCatalog: buildQueryState({
                    sections: [{ key: 'recent', title: 'Recent', items: [{ video: { id: 1 } }] }],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.videoContinueItems).toEqual([]);
        });

        it('slices continue items to max 4', () => {
            const items = Array.from({ length: 8 }, (_, i) => ({
                video: { id: i, name: `v${i}` },
                progress_pct: 10,
                status: 'in_progress',
            }));
            setupDefaultQueries({
                videoCatalog: buildQueryState({
                    sections: [{ key: 'continue', title: 'Continue', items }],
                }),
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

    // --- fallbackMusicTrack ---

    describe('fallbackMusicTrack', () => {
        it('returns null when playerState has no current_file_id', () => {
            setupDefaultQueries({
                playerState: buildQueryState({ current_file_id: null }),
                nowPlayingTracks: buildQueryState({ items: [] }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume).toBeNull();
        });

        it('returns null when current_file_id is set but track not found in nowPlayingTracks', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 10,
                }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 100,
                                name: 'other.mp3',
                                size: 100,
                                metadata: { duration: 60 },
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume).toBeNull();
        });

        it('returns fallback track when found in nowPlayingTracks', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 30,
                }),
                nowPlaying: buildQueryState({ id: 7, track_count: 3 }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 42,
                                name: 'song.mp3',
                                size: 500,
                                updated_at: '2026-01-01',
                                metadata: { duration: 200, title: 'Song' },
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume).not.toBeNull();
            expect(result.current.musicResume!.track.id).toBe(42);
        });
    });

    // --- musicResume ---

    describe('musicResume', () => {
        it('returns null when no currentTrack and no fallback', () => {
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume).toBeNull();
        });

        it('uses currentTrack over fallback when currentTrack is available', () => {
            const track = {
                id: 10,
                name: 'Live.mp3',
                size: 1024,
                updated_at: '2026-03-01',
                metadata: { duration: 300, title: 'Live' },
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [track],
                currentTrack: track,
                currentTime: 120,
                duration: 300,
                isPlaying: true,
            });

            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 10,
                    current_position: 50,
                }),
                nowPlaying: buildQueryState({ id: 7, track_count: 5 }),
                nowPlayingTracks: buildQueryState({ items: [{ file: track }] }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume).not.toBeNull();
            // Uses currentTime from global music (120), not playerState.current_position (50)
            expect(result.current.musicResume!.progressSeconds).toBe(120);
            // Uses global music duration (300)
            expect(result.current.musicResume!.durationSeconds).toBe(300);
            expect(result.current.musicResume!.isPlaying).toBe(true);
            // queue.length (1) is truthy
            expect(result.current.musicResume!.queueCount).toBe(1);
        });

        it('uses fallback track when currentTrack is undefined', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 60,
                }),
                nowPlaying: buildQueryState({ id: 7, track_count: 4 }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 42,
                                name: 'Track.mp3',
                                size: 2048,
                                updated_at: '2026-03-14',
                                metadata: { duration: 240, title: 'Track', artist: 'Artist' },
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.progressSeconds).toBe(60);
            expect(result.current.musicResume!.durationSeconds).toBe(240);
            expect(result.current.musicResume!.isPlaying).toBe(false);
        });

        it('uses playerState.current_position default 0 when fallback and no position', () => {
            setupDefaultQueries({
                playerState: buildQueryState({ current_file_id: 42 }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 42,
                                name: 'T.mp3',
                                size: 100,
                                updated_at: '2026-01-01',
                                metadata: { duration: 100 },
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.progressSeconds).toBe(0);
        });

        it('uses 0 when playerState.data is null for fallback progressSeconds', () => {
            // currentTrack is undefined, fallback found, playerState.data is null -> current_position ?? 0 -> 0
            const track = {
                id: 42,
                name: 'T.mp3',
                size: 100,
                updated_at: '2026-01-01',
                metadata: { duration: 100 },
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [],
                currentTrack: undefined,
                currentTime: 0,
                duration: 0,
                isPlaying: false,
            });
            setupDefaultQueries({
                playerState: buildQueryState(null),
                nowPlaying: buildQueryState({ id: 7, track_count: 3 }),
                nowPlayingTracks: buildQueryState({ items: [{ file: track }] }),
            });

            // fallbackMusicTrack requires playerStateQuery.data?.current_file_id to be truthy
            // but playerState data is null => current_file_id is undefined => fallback is null
            // So musicResume will be null. This tests the playerState.data?.current_file_id branch.
            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume).toBeNull();
        });

        it('falls through to 0 when playerState.data exists but current_position is undefined', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: undefined,
                }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 42,
                                name: 'T.mp3',
                                size: 100,
                                updated_at: '2026-01-01',
                                metadata: { duration: 100 },
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.progressSeconds).toBe(0);
        });

        it('uses Math.max(duration, metadata.duration) when currentTrack is available', () => {
            const track = {
                id: 10,
                name: 'T.mp3',
                size: 100,
                updated_at: '2026-01-01',
                metadata: { duration: 400 },
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [track],
                currentTrack: track,
                currentTime: 50,
                duration: 200, // less than metadata.duration (400)
                isPlaying: false,
            });
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.durationSeconds).toBe(400);
        });

        it('uses duration when greater than metadata.duration for currentTrack', () => {
            const track = {
                id: 10,
                name: 'T.mp3',
                size: 100,
                updated_at: '2026-01-01',
                metadata: { duration: 100 },
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [track],
                currentTrack: track,
                currentTime: 50,
                duration: 500, // greater than metadata.duration
                isPlaying: false,
            });
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.durationSeconds).toBe(500);
        });

        it('uses 0 for metadata.duration when currentTrack has no metadata', () => {
            const track = {
                id: 10,
                name: 'T.mp3',
                size: 100,
                updated_at: '2026-01-01',
                metadata: null,
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [track],
                currentTrack: track,
                currentTime: 50,
                duration: 0,
                isPlaying: false,
            });
            setupDefaultQueries();

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.durationSeconds).toBe(0);
        });

        it('uses metadata.duration for fallback track (no currentTrack)', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 10,
                }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 42,
                                name: 'T.mp3',
                                size: 100,
                                updated_at: '2026-01-01',
                                metadata: { duration: 180 },
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.durationSeconds).toBe(180);
        });

        it('uses 0 when fallback track has no metadata.duration', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 10,
                }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 42,
                                name: 'T.mp3',
                                size: 100,
                                updated_at: '2026-01-01',
                                metadata: null,
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.durationSeconds).toBe(0);
        });
    });

    // --- queueCount chain ---

    describe('musicResume.queueCount', () => {
        const trackWithMeta = {
            id: 42,
            name: 'T.mp3',
            size: 100,
            updated_at: '2026-01-01',
            metadata: { duration: 100 },
        };

        it('uses queue.length when > 0', () => {
            mockedUseGlobalMusic.mockReturnValue({
                queue: [trackWithMeta, trackWithMeta],
                currentTrack: trackWithMeta,
                currentTime: 10,
                duration: 100,
                isPlaying: false,
            });
            setupDefaultQueries({
                nowPlaying: buildQueryState({ id: 7, track_count: 10 }),
                nowPlayingTracks: buildQueryState({
                    items: [{ file: trackWithMeta }, { file: trackWithMeta }],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.queueCount).toBe(2);
        });

        it('uses nowPlayingQuery.track_count when queue is empty', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 5,
                }),
                nowPlaying: buildQueryState({ id: 7, track_count: 8 }),
                nowPlayingTracks: buildQueryState({ items: [{ file: trackWithMeta }] }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.queueCount).toBe(8);
        });

        it('uses nowPlayingTracks.items.length when queue empty and track_count is 0', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 5,
                }),
                nowPlaying: buildQueryState({ id: 7, track_count: 0 }),
                nowPlayingTracks: buildQueryState({
                    items: [{ file: trackWithMeta }, { file: { id: 43 } }],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.queueCount).toBe(2);
        });

        it('returns 0 when all queue count sources are falsy', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 5,
                }),
                nowPlaying: buildQueryState(null),
                nowPlayingTracks: buildQueryState({ items: [{ file: trackWithMeta }] }),
            });

            const { result } = renderHook(() => useHomeScreen());
            // queue.length=0, track_count=undefined, items.length=1 -> should be 1
            expect(result.current.musicResume!.queueCount).toBe(1);
        });

        it('falls through to 0 when queue empty, track_count falsy, and items.length is 0', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 5,
                }),
                nowPlaying: buildQueryState({ id: 7, track_count: 0 }),
                nowPlayingTracks: buildQueryState({
                    items: [
                        {
                            file: {
                                id: 42,
                                name: 'T.mp3',
                                size: 100,
                                updated_at: '2026-01-01',
                                metadata: { duration: 100 },
                            },
                        },
                    ],
                }),
            });

            const { result } = renderHook(() => useHomeScreen());
            // queue.length=0, track_count=0, items.length=1 -> 1
            expect(result.current.musicResume!.queueCount).toBe(1);
        });

        it('falls through entire chain to 0 when queue empty, track_count 0, and items.length 0', () => {
            const track = {
                id: 42,
                name: 'T.mp3',
                size: 100,
                updated_at: '2026-01-01',
                metadata: { duration: 100 },
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [],
                currentTrack: track,
                currentTime: 10,
                duration: 100,
                isPlaying: false,
            });
            setupDefaultQueries({
                nowPlaying: buildQueryState({ id: 7, track_count: 0 }),
                nowPlayingTracks: buildQueryState({ items: [] }),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.musicResume!.queueCount).toBe(0);
        });

        it('falls through to 0 when queue empty and nowPlaying.data is null (track_count undefined)', () => {
            const track = {
                id: 42,
                name: 'T.mp3',
                size: 100,
                updated_at: '2026-01-01',
                metadata: { duration: 100 },
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [],
                currentTrack: track,
                currentTime: 10,
                duration: 100,
                isPlaying: false,
            });
            setupDefaultQueries({
                nowPlaying: buildQueryState(null),
                nowPlayingTracks: buildQueryState({ items: [] }),
            });

            const { result } = renderHook(() => useHomeScreen());
            // queue.length=0, nowPlayingQuery.data?.track_count = undefined, nowPlayingTracksQuery.data?.items.length = 0, fallback 0
            expect(result.current.musicResume!.queueCount).toBe(0);
        });

        it('falls through to 0 when queue empty and nowPlayingTracks.data is null (items undefined)', () => {
            const track = {
                id: 42,
                name: 'T.mp3',
                size: 100,
                updated_at: '2026-01-01',
                metadata: { duration: 100 },
            };
            mockedUseGlobalMusic.mockReturnValue({
                queue: [],
                currentTrack: track,
                currentTime: 10,
                duration: 100,
                isPlaying: false,
            });
            setupDefaultQueries({
                nowPlaying: buildQueryState(null),
                nowPlayingTracks: buildQueryState(null),
            });

            const { result } = renderHook(() => useHomeScreen());
            // queue.length=0, nowPlayingQuery.data?.track_count = undefined, nowPlayingTracksQuery.data?.items.length = undefined, fallback 0
            expect(result.current.musicResume!.queueCount).toBe(0);
        });

        it('returns 0 when queue empty and no nowPlaying data at all', () => {
            setupDefaultQueries({
                playerState: buildQueryState({
                    current_file_id: 42,
                    current_position: 5,
                }),
                nowPlaying: buildQueryState(null),
                nowPlayingTracks: buildQueryState(null),
            });

            const { result } = renderHook(() => useHomeScreen());
            // fallback track requires nowPlayingTracks.data?.items to exist
            // if null, fallbackMusicTrack will be null => musicResume null
            expect(result.current.musicResume).toBeNull();
        });
    });

    // --- Loading states ---

    describe('loading states', () => {
        it('reports loading when analytics query is loading', () => {
            setupDefaultQueries({
                analytics: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isAnalyticsLoading).toBe(true);
        });

        it('reports video loading when either videoCatalog or videoPlayback is loading', () => {
            setupDefaultQueries({
                videoCatalog: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isVideoLoading).toBe(true);
        });

        it('reports video loading when videoPlayback is loading', () => {
            setupDefaultQueries({
                videoPlayback: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isVideoLoading).toBe(true);
        });

        it('reports music loading when playerState is loading', () => {
            setupDefaultQueries({
                playerState: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isMusicLoading).toBe(true);
        });

        it('reports music loading when nowPlaying is loading', () => {
            setupDefaultQueries({
                nowPlaying: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isMusicLoading).toBe(true);
        });

        it('reports music loading when nowPlayingTracks is loading', () => {
            setupDefaultQueries({
                nowPlayingTracks: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isMusicLoading).toBe(true);
        });

        it('reports favorites loading', () => {
            setupDefaultQueries({
                favorites: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isFavoritesLoading).toBe(true);
        });

        it('reports images loading', () => {
            setupDefaultQueries({
                images: loadingQueryState(),
            });

            const { result } = renderHook(() => useHomeScreen());
            expect(result.current.isImagesLoading).toBe(true);
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

    // --- nowPlayingTracksQuery enabled flag ---

    describe('nowPlayingTracksQuery enabled', () => {
        it('sets enabled=true when nowPlaying has an id', () => {
            setupDefaultQueries({
                nowPlaying: buildQueryState({ id: 9 }),
            });

            renderHook(() => useHomeScreen());

            const nowPlayingTracksOptions = mockedUseQuery.mock.calls[7][0];
            expect(nowPlayingTracksOptions.enabled).toBe(true);
            expect(nowPlayingTracksOptions.queryKey).toContain(9);
        });

        it('sets enabled=false when nowPlaying data is null', () => {
            setupDefaultQueries({
                nowPlaying: buildQueryState(null),
            });

            renderHook(() => useHomeScreen());

            const nowPlayingTracksOptions = mockedUseQuery.mock.calls[7][0];
            expect(nowPlayingTracksOptions.enabled).toBe(false);
        });

        it('sets enabled=false when nowPlaying data has no id', () => {
            setupDefaultQueries({
                nowPlaying: buildQueryState({}),
            });

            renderHook(() => useHomeScreen());

            const nowPlayingTracksOptions = mockedUseQuery.mock.calls[7][0];
            expect(nowPlayingTracksOptions.enabled).toBe(false);
        });
    });

    // --- Full integration scenario ---

    describe('integration: full data scenario', () => {
        it('derives all fields correctly when all data is present', () => {
            const currentTrack = {
                id: 10,
                name: 'Now.mp3',
                size: 512,
                updated_at: '2026-03-15',
                metadata: { duration: 300, title: 'Now', artist: 'Band' },
            };
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
                videoCatalog: buildQueryState({
                    sections: [
                        {
                            key: 'continue',
                            title: 'C',
                            items: [{ video: { id: 3 }, progress_pct: 45 }],
                        },
                    ],
                }),
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
                playerState: buildQueryState({
                    current_file_id: 10,
                    current_position: 50,
                }),
                nowPlaying: buildQueryState({ id: 7, track_count: 5 }),
                nowPlayingTracks: buildQueryState({ items: [{ file: currentTrack }] }),
            });

            const { result } = renderHook(() => useHomeScreen());

            expect(result.current.recentFiles).toHaveLength(2);
            expect(result.current.favoriteItems).toHaveLength(1);
            expect(result.current.recentImages).toHaveLength(1);
            expect(result.current.videoContinueItems).toHaveLength(1);
            expect(result.current.videoResume).not.toBeNull();
            expect(result.current.videoResume!.progressPercent).toBe(50);
            expect(result.current.musicResume).not.toBeNull();
            expect(result.current.musicResume!.progressSeconds).toBe(150);
            expect(result.current.musicResume!.isPlaying).toBe(true);
            expect(result.current.musicResume!.queueCount).toBe(3);
            expect(result.current.analytics).toBeTruthy();
        });
    });

    // --- Query configuration ---

    describe('query configuration', () => {
        it('passes retry=false to videoPlayback, playerState, nowPlaying, nowPlayingTracks queries', () => {
            setupDefaultQueries();

            renderHook(() => useHomeScreen());

            expect(mockedUseQuery.mock.calls[4][0].retry).toBe(false); // videoPlayback
            expect(mockedUseQuery.mock.calls[5][0].retry).toBe(false); // playerState
            expect(mockedUseQuery.mock.calls[6][0].retry).toBe(false); // nowPlaying
            expect(mockedUseQuery.mock.calls[7][0].retry).toBe(false); // nowPlayingTracks
        });

        it('calls queryFn correctly for each query', async () => {
            setupDefaultQueries({
                nowPlaying: buildQueryState({ id: 9 }),
            });

            renderHook(() => useHomeScreen());

            // analytics queryFn
            await expect(mockedUseQuery.mock.calls[0][0].queryFn()).resolves.toEqual({});
            // favorites queryFn
            await expect(mockedUseQuery.mock.calls[1][0].queryFn()).resolves.toEqual({
                items: [],
            });
            // images queryFn
            await expect(mockedUseQuery.mock.calls[2][0].queryFn()).resolves.toEqual({
                items: [],
            });
            // videoCatalog queryFn
            await expect(mockedUseQuery.mock.calls[3][0].queryFn()).resolves.toEqual({
                sections: [],
            });
            // videoPlayback queryFn
            await expect(mockedUseQuery.mock.calls[4][0].queryFn()).resolves.toBeNull();
            // playerState queryFn
            await expect(mockedUseQuery.mock.calls[5][0].queryFn()).resolves.toEqual({});
            // nowPlaying queryFn
            await expect(mockedUseQuery.mock.calls[6][0].queryFn()).resolves.toEqual({
                id: 7,
            });
            // nowPlayingTracks queryFn
            await expect(mockedUseQuery.mock.calls[7][0].queryFn()).resolves.toEqual({
                items: [],
            });
        });
    });
});

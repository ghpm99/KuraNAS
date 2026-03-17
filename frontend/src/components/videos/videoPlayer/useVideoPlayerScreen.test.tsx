import { act, renderHook } from '@testing-library/react';
import useVideoPlayerScreen from './useVideoPlayerScreen';

const mockUseVideoPlayer = jest.fn();
const mockNavigate = jest.fn();
const mockUseParams = jest.fn();
const mockUseLocation = jest.fn();
const mockPlayVideo = jest.fn();
const mockNextVideo = jest.fn();
const mockPreviousVideo = jest.fn();
const mockOnVideoEnded = jest.fn();

jest.mock('@/components/hooks/useVideoPlayer/useVideoPlayer', () => ({
    __esModule: true,
    default: (...args: any[]) => mockUseVideoPlayer(...args),
}));

const mockSearchParams = new URLSearchParams();
jest.mock('react-router-dom', () => ({
    useNavigate: () => mockNavigate,
    useParams: () => mockUseParams(),
    useLocation: () => mockUseLocation(),
    useSearchParams: () => [mockSearchParams],
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) => {
            const translations: Record<string, string> = {
                HOME: 'Home',
                FILES: 'Files',
                FAVORITES_PAGE_TITLE: 'Favorites',
                NAV_VIDEOS: 'Videos',
                VIDEO_BACK: 'Back',
                VIDEO_INVALID_ID: 'Invalid video ID',
                VIDEO_NO_VIDEO_PLAYING: 'No video playing',
                VIDEO_SECTION_SERIES: 'Series',
                VIDEO_SECTION_MOVIES: 'Movies',
                VIDEO_SECTION_PERSONAL: 'Personal',
                VIDEO_SECTION_CLIPS: 'Clips',
                VIDEO_SECTION_FOLDERS: 'Folders',
                VIDEO_PLAYER_FROM_CONTEXT: 'From {{context}}',
                VIDEO_PLAYER_POSITION: 'Item {{current}} of {{total}}',
                VIDEO_PLAYER_RESUME_POSITION: 'Resume at {{time}}',
                VIDEO_PLAYER_NEXT_EPISODES: 'Next episodes',
                VIDEO_PLAYER_RELATED_VIDEOS: 'Related videos',
            };
            const template = translations[key] ?? key;
            return Object.entries(params ?? {}).reduce(
                (result, [name, value]) => result.replace(`{{${name}}}`, value),
                template
            );
        },
    }),
}));

jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
    useSettings: () => ({
        settings: {
            players: {
                remember_video_progress: true,
                autoplay_next_video: true,
            },
        },
    }),
}));

const buildPlaylist = (overrides?: Partial<any>) => ({
    id: 7,
    type: 'series',
    source_path: '/library/my-show',
    name: 'My Show',
    is_hidden: false,
    is_auto: true,
    group_mode: 'prefix',
    classification: 'series',
    item_count: 3,
    cover_video_id: 30,
    created_at: '',
    updated_at: '',
    last_played_at: null,
    items: [
        {
            id: 130,
            order_index: 0,
            source_kind: 'auto',
            status: 'in_progress',
            progress_pct: 50,
            video: {
                id: 30,
                name: 'My.Show.S01E01.mkv',
                format: 'mkv',
                path: '',
                parent_path: '/library/my-show',
                size: 1,
            },
        },
        {
            id: 131,
            order_index: 1,
            source_kind: 'auto',
            status: 'not_started',
            progress_pct: 0,
            video: {
                id: 31,
                name: 'My.Show.S01E02.mkv',
                format: 'mkv',
                path: '',
                parent_path: '/library/my-show',
                size: 1,
            },
        },
        {
            id: 132,
            order_index: 2,
            source_kind: 'auto',
            status: 'completed',
            progress_pct: 100,
            video: {
                id: 32,
                name: 'My.Show.S01E03.mkv',
                format: 'mkv',
                path: '',
                parent_path: '/library/my-show',
                size: 1,
            },
        },
    ],
    ...overrides,
});

const buildPlayerState = (overrides?: Partial<any>) => ({
    videoRef: { current: null },
    playVideo: mockPlayVideo,
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    setPlaybackRate: jest.fn(),
    toggleFullscreen: jest.fn(),
    togglePlayPause: jest.fn(),
    nextVideo: mockNextVideo,
    previousVideo: mockPreviousVideo,
    status: 'playing',
    currentTime: 120,
    duration: 1800,
    volume: 0.8,
    playbackRate: 1,
    isFullscreen: false,
    setCurrentTime: jest.fn(),
    setDuration: jest.fn(),
    currentVideo: {
        id: 30,
        name: 'My.Show.S01E01.mkv',
        format: 'mkv',
        path: '',
        parent_path: '/library/my-show',
        size: 1,
    },
    playlist: buildPlaylist(),
    playbackState: {
        id: 1,
        client_id: 'client',
        playlist_id: 7,
        video_id: 30,
        current_time: 120,
        duration: 1800,
        is_paused: false,
        completed: false,
        last_update: '',
    },
    onVideoEnded: mockOnVideoEnded,
    ...overrides,
});

describe('components/videos/videoPlayer/useVideoPlayerScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockSearchParams.delete('playlist');
        mockSearchParams.delete('from');
        mockUseParams.mockReturnValue({ id: '30' });
        mockUseLocation.mockReturnValue({
            state: { from: '/videos/series/my-show', playlistId: 7 },
        });
        mockUseVideoPlayer.mockReturnValue(buildPlayerState());
    });

    it('loads playback, exposes context metadata and navigates back to the source route', () => {
        const { result } = renderHook(() => useVideoPlayerScreen());

        expect(mockPlayVideo).toHaveBeenCalledTimes(1);
        expect(result.current.originBadgeLabel).toBe('Series');
        expect(result.current.contextDescription).toBe('From My Show');
        expect(result.current.metadataLine).toBe('My Show • Item 1 of 3 • Resume at 2:00');
        expect(result.current.nextItem?.video.id).toBe(31);
        expect(result.current.relatedItems.map((item) => item.video.id)).toEqual([32]);
        expect(result.current.relatedTitle).toBe('Next episodes');

        act(() => {
            result.current.handleBack();
        });

        expect(mockNavigate).toHaveBeenCalledWith('/videos/series/my-show');
    });

    it('keeps the route synced with the active video without restarting playback after replace navigation', () => {
        let state = buildPlayerState();
        mockUseVideoPlayer.mockImplementation(() => state);

        const { rerender } = renderHook(() => useVideoPlayerScreen());

        expect(mockPlayVideo).toHaveBeenCalledTimes(1);

        state = buildPlayerState({
            currentVideo: {
                id: 31,
                name: 'My.Show.S01E02.mkv',
                format: 'mkv',
                path: '',
                parent_path: '/library/my-show',
                size: 1,
            },
            playbackState: {
                ...buildPlayerState().playbackState,
                video_id: 31,
            },
        });

        rerender();

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.stringContaining('/video/31'),
            expect.objectContaining({
                replace: true,
                state: expect.objectContaining({
                    from: '/videos/series/my-show',
                    playlistId: 7,
                }),
            })
        );

        mockUseParams.mockReturnValue({ id: '31' });
        rerender();

        expect(mockPlayVideo).toHaveBeenCalledTimes(1);
    });

    it('marks the current video as completed and only advances automatically when another item exists', async () => {
        const { result, rerender } = renderHook(() => useVideoPlayerScreen());

        await act(async () => {
            await result.current.handlePlaybackEnded();
        });

        expect(mockOnVideoEnded).toHaveBeenCalledTimes(1);
        expect(mockNextVideo).toHaveBeenCalledTimes(1);

        mockUseVideoPlayer.mockReturnValue(
            buildPlayerState({
                playlist: buildPlaylist({
                    item_count: 1,
                    items: [
                        {
                            id: 130,
                            order_index: 0,
                            source_kind: 'auto',
                            status: 'in_progress',
                            progress_pct: 50,
                            video: {
                                id: 30,
                                name: 'Movie.mkv',
                                format: 'mkv',
                                path: '',
                                parent_path: '/library/movies',
                                size: 1,
                            },
                        },
                    ],
                }),
            })
        );

        rerender();

        await act(async () => {
            await result.current.handlePlaybackEnded();
        });

        expect(mockOnVideoEnded).toHaveBeenCalledTimes(2);
        expect(mockNextVideo).toHaveBeenCalledTimes(1);
    });

    it('covers route origin variants and generic fallbacks without a loaded playlist', () => {
        const scenarios = [
            { from: '/home', expectedBadge: 'Home', expectedContext: 'From Home' },
            { from: '/files', expectedBadge: 'Files', expectedContext: 'From Files' },
            {
                from: '/starred',
                expectedBadge: 'Favorites',
                expectedContext: 'From Favorites',
            },
            {
                from: '/videos/movies',
                expectedBadge: 'Videos',
                expectedContext: 'From Movies',
            },
            {
                from: '/external',
                expectedBadge: 'Videos',
                expectedContext: 'From Videos',
            },
        ];

        scenarios.forEach(({ from, expectedBadge, expectedContext }) => {
            mockUseLocation.mockReturnValue({ state: { from, playlistId: null } });
            mockUseVideoPlayer.mockReturnValue(
                buildPlayerState({
                    playlist: null,
                    playbackState: null,
                    currentVideo: null,
                })
            );

            const { result, unmount } = renderHook(() => useVideoPlayerScreen());

            expect(result.current.originBadgeLabel).toBe(expectedBadge);
            expect(result.current.contextDescription).toBe(expectedContext);
            expect(result.current.relatedItems).toEqual([]);

            act(() => {
                result.current.openVideo(55);
            });

            expect(mockNavigate).toHaveBeenCalledWith(
                expect.stringContaining('/video/55'),
                expect.objectContaining({
                    state: expect.objectContaining({ from, playlistId: null }),
                })
            );

            unmount();
            jest.clearAllMocks();
        });
    });

    it('skips autoplay when the route id is missing and formats long resume durations', () => {
        mockUseParams.mockReturnValue({ id: undefined });
        mockUseLocation.mockReturnValue({ state: { from: '', playlistId: null } });
        mockUseVideoPlayer.mockReturnValue(
            buildPlayerState({
                playlist: null,
                playbackState: {
                    id: 1,
                    client_id: 'client',
                    playlist_id: null,
                    video_id: null,
                    current_time: 3661,
                    duration: 7200,
                    is_paused: false,
                    completed: false,
                    last_update: '',
                },
                currentVideo: {
                    id: 30,
                    name: 'Long Movie.mkv',
                    format: 'mkv',
                    path: '',
                    parent_path: '/movies',
                    size: 1,
                },
            })
        );

        const { result } = renderHook(() => useVideoPlayerScreen());

        expect(result.current.isInvalidVideoId).toBe(true);
        expect(result.current.metadataLine).toBe('Resume at 1:01:01');
        expect(mockPlayVideo).not.toHaveBeenCalled();
        expect(result.current.originBadgeLabel).toBe('Videos');
        expect(result.current.contextDescription).toBe('From Videos');
    });
});

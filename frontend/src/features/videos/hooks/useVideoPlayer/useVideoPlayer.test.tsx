import { act, renderHook } from '@testing-library/react';
import useVideoPlayer from './useVideoPlayer';

const mockStartVideoPlayback = jest.fn();
const mockNextVideoPlayback = jest.fn();
const mockPreviousVideoPlayback = jest.fn();
const mockUpdateVideoPlaybackState = jest.fn();
const mockGetApiV1BaseUrl = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
    startVideoPlayback: (...args: any[]) => mockStartVideoPlayback(...args),
    nextVideoPlayback: (...args: any[]) => mockNextVideoPlayback(...args),
    previousVideoPlayback: (...args: any[]) => mockPreviousVideoPlayback(...args),
    updateVideoPlaybackState: (...args: any[]) => mockUpdateVideoPlaybackState(...args),
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => mockGetApiV1BaseUrl(),
}));

const makeSession = (videoId: number, playlistId = 11) => ({
    playlist: {
        id: playlistId,
        type: 'custom',
        source_path: '/',
        name: 'playlist',
        is_hidden: false,
        is_auto: false,
        group_mode: 'single',
        classification: 'personal',
        item_count: 1,
        cover_video_id: videoId,
        created_at: '',
        updated_at: '',
        last_played_at: null,
        items: [
            {
                id: 100 + videoId,
                order_index: 0,
                source_kind: 'manual',
                status: 'in_progress',
                video: {
                    id: videoId,
                    name: `video-${videoId}.mp4`,
                    format: 'mp4',
                    path: '',
                    parent_path: '',
                    size: 1,
                },
            },
        ],
    },
    playback_state: {
        id: 1,
        client_id: 'client',
        playlist_id: playlistId,
        video_id: videoId,
        current_time: 12,
        duration: 120,
        is_paused: false,
        completed: false,
        last_update: '',
    },
});

const createFakeVideo = () =>
    ({
        src: '',
        currentTime: 0,
        volume: 1,
        playbackRate: 1,
        play: jest.fn(() => Promise.resolve()),
        pause: jest.fn(),
        requestFullscreen: jest.fn(),
    }) as any;

describe('hooks/useVideoPlayer', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        jest.useFakeTimers();
        mockGetApiV1BaseUrl.mockReturnValue('http://localhost:8000/v1');
        mockStartVideoPlayback.mockResolvedValue(makeSession(7));
        mockNextVideoPlayback.mockResolvedValue(makeSession(8));
        mockPreviousVideoPlayback.mockResolvedValue(makeSession(6));
        mockUpdateVideoPlaybackState.mockResolvedValue({});
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('starts playback and sets video source from session', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        expect(mockStartVideoPlayback).toHaveBeenCalledWith(7, 11);
        expect(fakeVideo.src).toContain('/files/video-stream/7');
        expect(result.current.status).toBe('playing');
        expect(result.current.currentVideo?.id).toBe(7);
        expect(result.current.playlist?.id).toBe(11);
        expect(result.current.playbackState?.video_id).toBe(7);
    });

    it('clamps volume and applies playback rate', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        act(() => {
            result.current.setVolume(2);
        });
        expect(result.current.volume).toBe(1);

        act(() => {
            result.current.setPlaybackRate(1.5);
        });
        expect(result.current.playbackRate).toBe(1.5);
    });

    it('seeks to specific time', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        act(() => {
            result.current.seekTo(34);
        });
        expect(result.current.currentTime).toBe(34);
    });

    it('pauses and resumes video', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        await act(async () => {
            result.current.pause();
            result.current.resume();
        });

        expect(fakeVideo.pause).toHaveBeenCalled();
        expect(fakeVideo.play).toHaveBeenCalled();
    });

    it('toggles fullscreen on and off', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        Object.defineProperty(document, 'fullscreenElement', {
            value: null,
            configurable: true,
        });
        act(() => {
            result.current.toggleFullscreen();
        });
        expect(fakeVideo.requestFullscreen).toHaveBeenCalled();
        expect(result.current.isFullscreen).toBe(true);

        (document as any).exitFullscreen = jest.fn();
        Object.defineProperty(document, 'fullscreenElement', {
            value: {},
            configurable: true,
        });
        act(() => {
            result.current.toggleFullscreen();
        });
        expect((document as any).exitFullscreen).toHaveBeenCalled();
        expect(result.current.isFullscreen).toBe(false);
    });

    it('navigates to next and previous video', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        await act(async () => {
            await result.current.nextVideo();
        });
        expect(mockNextVideoPlayback).toHaveBeenCalled();

        await act(async () => {
            await result.current.previousVideo();
        });
        expect(mockPreviousVideoPlayback).toHaveBeenCalled();
    });

    it('syncs playback state periodically', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        act(() => {
            result.current.setDuration(120);
        });

        await act(async () => {
            jest.advanceTimersByTime(5000);
        });
        expect(mockUpdateVideoPlaybackState).toHaveBeenCalled();
    });

    it('handles missing video in session gracefully', async () => {
        mockStartVideoPlayback.mockResolvedValue({
            ...makeSession(0),
            playback_state: {
                ...makeSession(0).playback_state,
                video_id: 0,
                playlist_id: null,
            },
        });

        const { result } = renderHook(() => useVideoPlayer({ videoId: '0', playlistId: null }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        expect(result.current.currentVideo).toBeNull();
        expect(result.current.playbackState?.video_id).toBe(0);
    });

    it('sets paused status when play rejects', async () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        mockNextVideoPlayback.mockResolvedValue({
            ...makeSession(9),
            playback_state: {
                ...makeSession(9).playback_state,
                current_time: 0,
            },
        });
        fakeVideo.play.mockImplementationOnce(() => Promise.reject(new Error('play failed')));

        await act(async () => {
            await result.current.nextVideo();
        });
        expect(result.current.status).toBe('paused');
    });

    it('tolerates null videoRef for control operations', async () => {
        mockStartVideoPlayback.mockResolvedValue({
            ...makeSession(0),
            playback_state: {
                ...makeSession(0).playback_state,
                video_id: 0,
                playlist_id: null,
            },
        });

        const { result } = renderHook(() => useVideoPlayer({ videoId: '0', playlistId: null }));
        const fakeVideo = createFakeVideo();

        act(() => {
            result.current.videoRef.current = fakeVideo;
        });

        await act(async () => {
            await result.current.playVideo();
        });

        act(() => {
            result.current.videoRef.current = null;
            result.current.pause();
            result.current.resume();
            result.current.seekTo(5);
            result.current.setVolume(0.4);
            result.current.setPlaybackRate(2);
        });
    });
});

import { act, renderHook } from '@testing-library/react';
import useVideoPlayer from './useVideoPlayer';
import { videoPlaybackPreferencesStorageKey } from './videoPlaybackPreferences';

const mockStartVideoPlayback = jest.fn();
const mockNextVideoPlayback = jest.fn();
const mockPreviousVideoPlayback = jest.fn();

jest.mock('@/service/playerStateFlush', () => ({
    flushVideoPlaybackState: jest.fn(),
}));

jest.mock('@/service/videoPlayback', () => ({
    startVideoPlayback: (...args: any[]) => mockStartVideoPlayback(...args),
    nextVideoPlayback: (...args: any[]) => mockNextVideoPlayback(...args),
    previousVideoPlayback: (...args: any[]) => mockPreviousVideoPlayback(...args),
    updateVideoPlaybackState: jest.fn(() => Promise.resolve({})),
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => 'http://localhost:8000/v1',
}));

const makeSession = (videoId: number) => ({
    playlist: {
        id: 11,
        items: [{ id: 100 + videoId, video: { id: videoId, name: 'v.mp4', format: 'mp4' } }],
    },
    playback_state: { playlist_id: 11, video_id: videoId, current_time: 0, duration: 60 },
});

const createFakeVideo = () =>
    ({
        src: '',
        currentTime: 0,
        volume: 1,
        muted: false,
        playbackRate: 1,
        play: jest.fn(() => Promise.resolve()),
        pause: jest.fn(),
    }) as any;

const readStoredPreferences = () =>
    JSON.parse(window.localStorage.getItem(videoPlaybackPreferencesStorageKey) ?? 'null');

const mountPlayer = () => {
    const hook = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
    const fakeVideo = createFakeVideo();
    act(() => {
        hook.result.current.videoRef.current = fakeVideo;
    });
    return { ...hook, fakeVideo };
};

describe('hooks/useVideoPlayer playback preferences', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        jest.useFakeTimers();
        window.localStorage.clear();
        mockStartVideoPlayback.mockResolvedValue(makeSession(7));
        mockNextVideoPlayback.mockResolvedValue(makeSession(8));
        mockPreviousVideoPlayback.mockResolvedValue(makeSession(6));
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('mounts with defaults when nothing is stored and storage throws', () => {
        jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });

        const { result } = renderHook(() => useVideoPlayer({ videoId: '7' }));

        expect(result.current.volume).toBe(1);
        expect(result.current.playbackRate).toBe(1);
        jest.restoreAllMocks();
    });

    it('starts with the saved preferences', () => {
        window.localStorage.setItem(
            videoPlaybackPreferencesStorageKey,
            JSON.stringify({ volume: 0.3, isMuted: false, playbackRate: 1.5 })
        );

        const { result } = renderHook(() => useVideoPlayer({ videoId: '7' }));

        expect(result.current.volume).toBe(0.3);
        expect(result.current.playbackRate).toBe(1.5);
    });

    it('applies the saved preferences to the video on every loaded video', async () => {
        window.localStorage.setItem(
            videoPlaybackPreferencesStorageKey,
            JSON.stringify({ volume: 0.3, isMuted: true, playbackRate: 1.5 })
        );
        const { result, fakeVideo } = mountPlayer();

        await act(async () => {
            await result.current.playVideo();
        });
        expect(fakeVideo).toMatchObject({ volume: 0.3, muted: true, playbackRate: 1.5 });

        fakeVideo.volume = 1;
        fakeVideo.muted = false;
        fakeVideo.playbackRate = 1;
        await act(async () => {
            await result.current.nextVideo();
        });
        expect(fakeVideo).toMatchObject({ volume: 0.3, muted: true, playbackRate: 1.5 });

        fakeVideo.volume = 1;
        fakeVideo.muted = false;
        fakeVideo.playbackRate = 1;
        await act(async () => {
            await result.current.previousVideo();
        });
        expect(fakeVideo).toMatchObject({ volume: 0.3, muted: true, playbackRate: 1.5 });
    });

    it('applies preferences changed during playback to the next video', async () => {
        const { result, fakeVideo } = mountPlayer();
        await act(async () => {
            await result.current.playVideo();
        });

        act(() => {
            result.current.setVolume(0.6);
            result.current.setPlaybackRate(2);
        });
        fakeVideo.volume = 1;
        fakeVideo.playbackRate = 1;
        await act(async () => {
            await result.current.nextVideo();
        });

        expect(fakeVideo).toMatchObject({ volume: 0.6, muted: false, playbackRate: 2 });
    });

    it('saves the playback rate immediately', () => {
        const { result } = mountPlayer();

        act(() => {
            result.current.setPlaybackRate(1.25);
        });

        expect(readStoredPreferences()).toEqual({ volume: 1, isMuted: false, playbackRate: 1.25 });
    });

    it('saves muted state when volume reaches zero and clears it when restored', () => {
        const { result } = mountPlayer();

        act(() => {
            result.current.setVolume(0);
            jest.advanceTimersByTime(300);
        });
        expect(readStoredPreferences()).toEqual({ volume: 0, isMuted: true, playbackRate: 1 });

        act(() => {
            result.current.setVolume(0.5);
            jest.advanceTimersByTime(300);
        });
        expect(readStoredPreferences()).toEqual({ volume: 0.5, isMuted: false, playbackRate: 1 });
    });

    it('debounces volume writes while the slider is dragged', () => {
        const setItem = jest.spyOn(Storage.prototype, 'setItem');
        const { result } = mountPlayer();

        act(() => {
            result.current.setVolume(0.9);
            jest.advanceTimersByTime(100);
            result.current.setVolume(0.7);
            jest.advanceTimersByTime(100);
            result.current.setVolume(0.4);
        });
        expect(setItem).not.toHaveBeenCalled();

        act(() => {
            jest.advanceTimersByTime(300);
        });
        expect(setItem).toHaveBeenCalledTimes(1);
        expect(readStoredPreferences().volume).toBe(0.4);
        setItem.mockRestore();
    });

    it('flushes a pending volume write on unmount', () => {
        const { result, unmount } = mountPlayer();

        act(() => {
            result.current.setVolume(0.2);
        });
        unmount();

        expect(readStoredPreferences().volume).toBe(0.2);
    });

    it('does not save anything when the browser changes the video on its own', async () => {
        const { result, fakeVideo } = mountPlayer();
        await act(async () => {
            await result.current.playVideo();
        });

        fakeVideo.muted = true;
        act(() => {
            jest.advanceTimersByTime(1000);
        });

        expect(window.localStorage.getItem(videoPlaybackPreferencesStorageKey)).toBeNull();
    });
});

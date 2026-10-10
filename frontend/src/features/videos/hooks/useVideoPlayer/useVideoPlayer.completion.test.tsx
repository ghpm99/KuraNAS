import { act, cleanup, renderHook } from '@testing-library/react';
import useVideoPlayer from './useVideoPlayer';
import { COMPLETED_PROGRESS_RATIO, hasReachedCompletionThreshold } from './playbackCompletion';

const mockStartVideoPlayback = jest.fn();
const mockUpdateVideoPlaybackState = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
    startVideoPlayback: (...args: any[]) => mockStartVideoPlayback(...args),
    nextVideoPlayback: jest.fn(),
    previousVideoPlayback: jest.fn(),
    updateVideoPlaybackState: (...args: any[]) => mockUpdateVideoPlaybackState(...args),
}));

const sessionAtStart = {
    playlist: { id: 11, items: [{ video: { id: 7, format: 'mp4' } }] },
    playback_state: {
        playlist_id: 11,
        video_id: 7,
        current_time: 0,
        duration: 100,
        is_paused: false,
        completed: false,
    },
};

const renderPlayingHook = async () => {
    const hook = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
    act(() => {
        hook.result.current.videoRef.current = {
            src: '',
            currentTime: 0,
            volume: 1,
            playbackRate: 1,
            play: jest.fn(() => Promise.resolve()),
            pause: jest.fn(),
        } as any;
    });
    await act(async () => {
        await hook.result.current.playVideo();
    });
    return hook;
};

const pauseAtPosition = async (positionSeconds: number) => {
    const hook = await renderPlayingHook();
    act(() => hook.result.current.setCurrentTime(positionSeconds));
    mockUpdateVideoPlaybackState.mockClear();
    act(() => hook.result.current.pause());
    return mockUpdateVideoPlaybackState.mock.calls[0][0];
};

describe('hooks/useVideoPlayer completion threshold', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockStartVideoPlayback.mockResolvedValue(sessionAtStart);
        mockUpdateVideoPlaybackState.mockResolvedValue({});
        Object.defineProperty(navigator, 'sendBeacon', {
            value: jest.fn().mockReturnValue(true),
            configurable: true,
        });
    });

    afterEach(() => cleanup());

    it('detects the threshold only with a known duration', () => {
        expect(hasReachedCompletionThreshold(89, 100)).toBe(false);
        expect(hasReachedCompletionThreshold(100 * COMPLETED_PROGRESS_RATIO, 100)).toBe(true);
        expect(hasReachedCompletionThreshold(500, 0)).toBe(false);
    });

    it('does not send completed below the threshold', async () => {
        const request = await pauseAtPosition(89);
        expect(request.completed).toBe(false);
    });

    it('sends completed when playback crosses the threshold', async () => {
        const request = await pauseAtPosition(90);
        expect(request.completed).toBe(true);
        expect(request.current_time).toBe(90);
    });
});

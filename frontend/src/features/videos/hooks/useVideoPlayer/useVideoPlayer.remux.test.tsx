import { act, renderHook } from '@testing-library/react';
import useVideoPlayer from './useVideoPlayer';

const mockStartVideoPlayback = jest.fn();
const mockUpdateVideoPlaybackState = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
    ...jest.requireActual('@/service/videoPlayback'),
    startVideoPlayback: (...args: any[]) => mockStartVideoPlayback(...args),
    updateVideoPlaybackState: (...args: any[]) => mockUpdateVideoPlaybackState(...args),
}));

const makeSession = (format: string, duration = 600, currentTime = 0) => ({
    playlist: {
        id: 11,
        type: 'custom',
        source_path: '/',
        name: 'playlist',
        is_hidden: false,
        is_auto: false,
        group_mode: 'single',
        classification: 'personal',
        item_count: 1,
        cover_video_id: 7,
        created_at: '',
        updated_at: '',
        last_played_at: null,
        items: [
            {
                id: 107,
                order_index: 0,
                source_kind: 'manual',
                status: 'in_progress',
                video: {
                    id: 7,
                    name: `video.${format}`,
                    format,
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
        playlist_id: 11,
        video_id: 7,
        current_time: currentTime,
        duration,
        is_paused: false,
        completed: false,
        last_update: '',
    },
});

const createFakeVideo = (canPlayType?: (mimeType: string) => string) =>
    ({
        src: '',
        currentTime: 0,
        play: jest.fn(() => Promise.resolve()),
        pause: jest.fn(),
        canPlayType,
    }) as any;

const startPlayback = async (fakeVideo: any) => {
    const rendered = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
    act(() => {
        rendered.result.current.videoRef.current = fakeVideo;
    });
    await act(async () => {
        await rendered.result.current.playVideo();
    });
    return rendered;
};

describe('hooks/useVideoPlayer remux', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        jest.useFakeTimers();
        mockUpdateVideoPlaybackState.mockResolvedValue({});
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('mounts with no service mock and exposes an idle player', () => {
        const { result } = renderHook(() => useVideoPlayer({ videoId: '7' }));
        expect(result.current.status).toBe('waiting');
        expect(result.current.playbackError).toBeNull();
        expect(result.current.duration).toBe(0);
    });

    it('keeps the default stream url when the browser plays the container', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mkv'));
        const fakeVideo = createFakeVideo(() => 'maybe');

        await startPlayback(fakeVideo);

        expect(fakeVideo.src).toContain('/files/video-stream/7');
    });

    it('selects the remux url up front when canPlayType rejects the container', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mkv', 600, 40));
        const fakeVideo = createFakeVideo(() => '');

        const { result } = await startPlayback(fakeVideo);

        expect(fakeVideo.src).toMatch(/\/video\/stream\/7\/remux\?start=40\.000$/);
        expect(result.current.currentTime).toBe(40);
    });

    it('falls back to remux when the element reports an unsupported source', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mp4'));
        const fakeVideo = createFakeVideo(() => 'probably');
        const { result } = await startPlayback(fakeVideo);
        expect(fakeVideo.src).toContain('/files/video-stream/7');

        await act(async () => {
            result.current.reportPlaybackError('unsupported');
        });

        expect(fakeVideo.src).toMatch(/\/video\/stream\/7\/remux$/);
        expect(result.current.playbackError).toBeNull();
    });

    it('shows the error overlay state when remux fails too', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mkv'));
        const fakeVideo = createFakeVideo(() => '');
        const { result } = await startPlayback(fakeVideo);

        act(() => {
            result.current.reportPlaybackError('unsupported');
        });

        expect(result.current.playbackError).toBe('unsupported');
        expect(result.current.status).toBe('paused');
    });

    it('rebuilds the remux url with start on seek and offsets the displayed time', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mkv'));
        const fakeVideo = createFakeVideo(() => '');
        const { result } = await startPlayback(fakeVideo);

        await act(async () => {
            result.current.seekTo(120.5);
        });

        expect(fakeVideo.src).toMatch(/\/video\/stream\/7\/remux\?start=120\.500$/);
        expect(result.current.currentTime).toBe(120.5);

        act(() => {
            result.current.setCurrentTime(3);
        });
        expect(result.current.currentTime).toBe(123.5);
    });

    it('keeps the metadata duration in remux mode and ignores the loaded total', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mkv', 600));
        const { result } = await startPlayback(createFakeVideo(() => ''));

        act(() => {
            result.current.setDuration(Infinity);
        });
        act(() => {
            result.current.setDuration(42);
        });

        expect(result.current.duration).toBe(600);
    });

    it('uses the loaded total plus offset when metadata has no duration', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mkv', 0, 100));
        const { result } = await startPlayback(createFakeVideo(() => ''));

        act(() => {
            result.current.setDuration(Infinity);
        });
        expect(result.current.duration).toBe(0);

        act(() => {
            result.current.setDuration(50);
        });
        expect(result.current.duration).toBe(150);
    });

    it('passes loaded duration and time straight through in direct mode', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mp4'));
        const { result } = await startPlayback(createFakeVideo(() => 'maybe'));

        act(() => {
            result.current.setCurrentTime(9);
            result.current.setDuration(77);
        });

        expect(result.current.currentTime).toBe(9);
        expect(result.current.duration).toBe(77);
    });

    it('seeks the element directly in direct mode', async () => {
        mockStartVideoPlayback.mockResolvedValue(makeSession('mp4'));
        const fakeVideo = createFakeVideo(() => 'maybe');
        const { result } = await startPlayback(fakeVideo);

        act(() => {
            result.current.seekTo(30);
        });

        expect(fakeVideo.currentTime).toBe(30);
        expect(fakeVideo.src).toContain('/files/video-stream/7');
    });
});

import { act, renderHook } from '@testing-library/react';
import useVideoPlayer from './useVideoPlayer';
import {
    loadVideoAudioPreferences,
    saveVideoAudioPreferences,
    videoAudioPreferencesStorageKey,
} from '@/features/videos/hooks/useVideoAudioTracks/videoAudioPreferences';

const mockStartVideoPlayback = jest.fn();
const mockNextVideoPlayback = jest.fn();
const mockGetVideoAudioTracks = jest.fn();

const mockGetVideoRemuxStartSeconds = jest.fn((_videoId: number, atSeconds: number) =>
    Promise.resolve(atSeconds)
);

jest.mock('@/service/playerStateFlush', () => ({
    flushVideoPlaybackState: jest.fn(),
}));

jest.mock('@/service/videoPlayback', () => ({
    ...jest.requireActual('@/service/videoPlayback'),
    startVideoPlayback: (...args: any[]) => mockStartVideoPlayback(...args),
    nextVideoPlayback: (...args: any[]) => mockNextVideoPlayback(...args),
    updateVideoPlaybackState: jest.fn().mockResolvedValue({}),
    getVideoAudioTracks: (...args: any[]) => mockGetVideoAudioTracks(...args),
    getVideoRemuxStartSeconds: (...args: [number, number]) =>
        mockGetVideoRemuxStartSeconds(...args),
}));

const japaneseAndPortugueseTracks = [
    { stream_index: 1, language: 'ja', title: '', codec: 'aac', channels: 2, is_default: true },
    { stream_index: 2, language: 'pt', title: 'Dub', codec: 'aac', channels: 2, is_default: false },
];

const makeSession = (videoId: number, currentTime = 0) => ({
    playlist: {
        id: 11,
        name: 'playlist',
        items: [
            {
                id: 100 + videoId,
                order_index: 0,
                video: { id: videoId, name: `video-${videoId}.mp4`, format: 'mp4' },
            },
        ],
    },
    playback_state: {
        id: 1,
        playlist_id: 11,
        video_id: videoId,
        current_time: currentTime,
        duration: 600,
        is_paused: false,
        completed: false,
    },
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
        canPlayType: () => 'probably',
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

describe('hooks/useVideoPlayer audio track', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
        mockStartVideoPlayback.mockResolvedValue(makeSession(7, 30));
        mockGetVideoAudioTracks.mockResolvedValue(japaneseAndPortugueseTracks);
    });

    it('exposes the tracks of the current video and marks the default one as current', async () => {
        const { result } = await startPlayback(createFakeVideo());

        expect(result.current.audioTracks).toEqual(japaneseAndPortugueseTracks);
        expect(result.current.currentAudioStreamIndex).toBe(1);
    });

    it('keeps the direct stream and exposes no tracks when the tracks request fails', async () => {
        mockGetVideoAudioTracks.mockRejectedValue(new Error('offline'));
        const fakeVideo = createFakeVideo();

        const { result } = await startPlayback(fakeVideo);

        expect(result.current.audioTracks).toEqual([]);
        expect(result.current.currentAudioStreamIndex).toBeNull();
        expect(fakeVideo.src).toContain('/files/video-stream/7');
    });

    it('switches from the direct stream to remux at the current position with the audio index', async () => {
        const fakeVideo = createFakeVideo();
        fakeVideo.volume = 0.4;
        const { result } = await startPlayback(fakeVideo);
        expect(fakeVideo.src).toContain('/files/video-stream/7');

        await act(async () => result.current.selectAudioTrack(2));

        expect(fakeVideo.src).toMatch(
            /\/video\/stream\/7\/remux\?start=30\.000&audio_stream_index=2$/
        );
        expect(result.current.currentAudioStreamIndex).toBe(2);
        expect(fakeVideo.playbackRate).toBe(1);
        expect(fakeVideo.volume).toBe(1);
    });

    it('keeps the audio index when seeking after the switch', async () => {
        const fakeVideo = createFakeVideo();
        const { result } = await startPlayback(fakeVideo);
        await act(async () => result.current.selectAudioTrack(2));

        await act(async () => result.current.seekTo(120));

        expect(fakeVideo.src).toMatch(
            /\/video\/stream\/7\/remux\?start=120\.000&audio_stream_index=2$/
        );
    });

    it('keeps the audio index when falling back from remux to transcode', async () => {
        const fakeVideo = createFakeVideo();
        const { result } = await startPlayback(fakeVideo);
        await act(async () => result.current.selectAudioTrack(2));

        await act(async () => {
            result.current.reportPlaybackError('decode');
        });

        expect(fakeVideo.src).toMatch(/\/video\/stream\/7\/transcode\?.*audio_stream_index=2$/);
    });

    it('remembers the language of the chosen track on the device', async () => {
        const { result } = await startPlayback(createFakeVideo());

        await act(async () => result.current.selectAudioTrack(2));

        expect(loadVideoAudioPreferences()).toEqual({ language: 'pt' });
    });

    it('starts the next video through remux in the remembered language', async () => {
        saveVideoAudioPreferences({ language: 'pt-BR' });
        mockStartVideoPlayback.mockResolvedValue(makeSession(7));
        mockNextVideoPlayback.mockResolvedValue(makeSession(8));
        const fakeVideo = createFakeVideo();
        const { result } = await startPlayback(fakeVideo);
        expect(fakeVideo.src).toMatch(/\/video\/stream\/7\/remux\?audio_stream_index=2$/);

        await act(async () => {
            await result.current.nextVideo();
        });

        expect(fakeVideo.src).toMatch(/\/video\/stream\/8\/remux\?audio_stream_index=2$/);
    });

    it('keeps the direct stream when the remembered language is already the default track', async () => {
        saveVideoAudioPreferences({ language: 'ja' });
        const fakeVideo = createFakeVideo();

        await startPlayback(fakeVideo);

        expect(fakeVideo.src).toContain('/files/video-stream/7');
    });

    it('keeps the direct stream when the video has no track in the remembered language', async () => {
        saveVideoAudioPreferences({ language: 'fr' });
        const fakeVideo = createFakeVideo();

        await startPlayback(fakeVideo);

        expect(fakeVideo.src).toContain('/files/video-stream/7');
    });

    it('does not block playback when the tracks request never answers', async () => {
        jest.useFakeTimers();
        saveVideoAudioPreferences({ language: 'pt' });
        mockGetVideoAudioTracks.mockReturnValue(new Promise(() => undefined));
        const fakeVideo = createFakeVideo();
        const rendered = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));
        act(() => {
            rendered.result.current.videoRef.current = fakeVideo;
        });

        const playing = act(async () => {
            const pendingPlay = rendered.result.current.playVideo();
            await jest.advanceTimersByTimeAsync(2000);
            await pendingPlay;
        });
        await playing;

        expect(fakeVideo.src).toContain('/files/video-stream/7');
        jest.useRealTimers();
    });

    it('persists the language under the documented storage key', () => {
        saveVideoAudioPreferences({ language: 'pt-BR' });

        expect(window.localStorage.getItem(videoAudioPreferencesStorageKey)).toBe(
            '{"language":"pt"}'
        );
        expect(videoAudioPreferencesStorageKey).toBe('kuranas.videos.audioPreferences.v1');
    });
});

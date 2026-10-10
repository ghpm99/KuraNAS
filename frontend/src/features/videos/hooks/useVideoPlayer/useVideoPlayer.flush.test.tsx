import { act, cleanup, renderHook } from '@testing-library/react';
import useVideoPlayer from './useVideoPlayer';
import { getPlayerClientId } from '@/service/playerClientId';

const mockStartVideoPlayback = jest.fn();
const mockUpdateVideoPlaybackState = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
    startVideoPlayback: (...args: any[]) => mockStartVideoPlayback(...args),
    nextVideoPlayback: jest.fn(),
    previousVideoPlayback: jest.fn(),
    updateVideoPlaybackState: (...args: any[]) => mockUpdateVideoPlaybackState(...args),
}));

const readBlobAsText = (blob: Blob) =>
    new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(String(reader.result));
        reader.onerror = () => reject(reader.error);
        reader.readAsText(blob);
    });

const sessionWithPlaylistAndVideo = {
    playlist: { id: 11, items: [{ video: { id: 7, format: 'mp4' } }] },
    playback_state: {
        playlist_id: 11,
        video_id: 7,
        current_time: 12,
        duration: 120,
        is_paused: false,
        completed: false,
    },
};

const expectedStateUrl = () =>
    `/api/v1/video/playback/state?client_id=${encodeURIComponent(getPlayerClientId())}`;

const expectedPlayingState = {
    playlist_id: 11,
    video_id: 7,
    current_time: 12,
    duration: 120,
    is_paused: false,
    completed: false,
};

const stubSendBeacon = (implementation: unknown) =>
    Object.defineProperty(navigator, 'sendBeacon', { value: implementation, configurable: true });

const setVisibilityState = (visibilityState: DocumentVisibilityState) =>
    Object.defineProperty(document, 'visibilityState', {
        value: visibilityState,
        configurable: true,
    });

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

describe('hooks/useVideoPlayer flush on page exit', () => {
    const originalSendBeacon = navigator.sendBeacon;
    const originalFetch = globalThis.fetch;

    beforeEach(() => {
        jest.clearAllMocks();
        mockStartVideoPlayback.mockResolvedValue(sessionWithPlaylistAndVideo);
        mockUpdateVideoPlaybackState.mockResolvedValue({});
    });

    afterEach(() => {
        cleanup();
        stubSendBeacon(originalSendBeacon);
        globalThis.fetch = originalFetch;
        setVisibilityState('visible');
    });

    it('does not flush on pagehide without a playback session', () => {
        const sendBeacon = jest.fn().mockReturnValue(true);
        stubSendBeacon(sendBeacon);
        const { unmount } = renderHook(() => useVideoPlayer({ videoId: '7', playlistId: 11 }));

        window.dispatchEvent(new Event('pagehide'));
        unmount();

        expect(sendBeacon).not.toHaveBeenCalled();
    });

    it('sends the current state by beacon on pagehide', async () => {
        const sendBeacon = jest.fn().mockReturnValue(true);
        stubSendBeacon(sendBeacon);
        globalThis.fetch = jest.fn();
        await renderPlayingHook();

        window.dispatchEvent(new Event('pagehide'));

        const [url, body] = sendBeacon.mock.calls[0];
        expect(url).toBe(expectedStateUrl());
        expect(JSON.parse(await readBlobAsText(body as Blob))).toEqual(expectedPlayingState);
        expect(globalThis.fetch).not.toHaveBeenCalled();
    });

    it('flushes when the page becomes hidden but not when visible', async () => {
        const sendBeacon = jest.fn().mockReturnValue(true);
        stubSendBeacon(sendBeacon);
        await renderPlayingHook();

        setVisibilityState('visible');
        document.dispatchEvent(new Event('visibilitychange'));
        expect(sendBeacon).not.toHaveBeenCalled();

        setVisibilityState('hidden');
        document.dispatchEvent(new Event('visibilitychange'));
        expect(sendBeacon).toHaveBeenCalledTimes(1);
        expect(sendBeacon.mock.calls[0][0]).toBe(expectedStateUrl());
    });

    it('falls back to keepalive fetch when the beacon is rejected', async () => {
        stubSendBeacon(jest.fn().mockReturnValue(false));
        globalThis.fetch = jest.fn().mockResolvedValue({});
        await renderPlayingHook();

        window.dispatchEvent(new Event('pagehide'));

        const [url, init] = (globalThis.fetch as jest.Mock).mock.calls[0];
        expect(url).toBe(expectedStateUrl());
        expect(init.method).toBe('PUT');
        expect(init.keepalive).toBe(true);
        expect(JSON.parse(init.body)).toEqual(expectedPlayingState);
    });

    it('flushes the state when the player unmounts', async () => {
        const sendBeacon = jest.fn().mockReturnValue(true);
        stubSendBeacon(sendBeacon);
        const { unmount } = await renderPlayingHook();

        unmount();

        expect(sendBeacon).toHaveBeenCalledTimes(1);
        const body = sendBeacon.mock.calls[0][1] as Blob;
        expect(JSON.parse(await readBlobAsText(body))).toEqual(expectedPlayingState);
    });

    it('stops listening after unmount', async () => {
        const sendBeacon = jest.fn().mockReturnValue(true);
        stubSendBeacon(sendBeacon);
        const { unmount } = await renderPlayingHook();
        unmount();
        sendBeacon.mockClear();

        window.dispatchEvent(new Event('pagehide'));

        expect(sendBeacon).not.toHaveBeenCalled();
    });
});

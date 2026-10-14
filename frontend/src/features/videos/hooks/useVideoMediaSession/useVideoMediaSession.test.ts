import { renderHook } from '@testing-library/react';
import useVideoMediaSession from './useVideoMediaSession';

jest.mock('@/service/videoPlayback', () => ({
    getVideoThumbnailUrl: (videoId: number) => `/thumbnails/${videoId}`,
}));

type Handler = (details: MediaSessionActionDetails) => void;

const video = {
    id: 7,
    name: 'Episode 1',
    parent_path: '/library/Show',
} as any;

const buildOptions = (overrides: Record<string, unknown> = {}) => ({
    currentVideo: video,
    isPlaying: true,
    currentTime: 30,
    duration: 100,
    hasPreviousVideo: true,
    hasNextVideo: true,
    togglePlayPause: jest.fn(),
    seekTo: jest.fn(),
    nextVideo: jest.fn(),
    previousVideo: jest.fn(),
    ...overrides,
});

describe('hooks/useVideoMediaSession', () => {
    let handlers: Record<string, Handler | null>;
    let mediaSession: any;
    const originalMediaMetadata = (global as any).MediaMetadata;

    beforeEach(() => {
        handlers = {};
        mediaSession = {
            metadata: null,
            playbackState: 'none',
            setActionHandler: jest.fn((action: string, handler: Handler | null) => {
                handlers[action] = handler;
            }),
            setPositionState: jest.fn(),
        };
        Object.defineProperty(navigator, 'mediaSession', {
            value: mediaSession,
            configurable: true,
        });
        (global as any).MediaMetadata = class {
            constructor(init: Record<string, unknown>) {
                Object.assign(this, init);
            }
        };
    });

    afterEach(() => {
        delete (navigator as any).mediaSession;
        (global as any).MediaMetadata = originalMediaMetadata;
    });

    it('publishes title, folder and thumbnail artwork', () => {
        renderHook(() => useVideoMediaSession(buildOptions()));

        expect(mediaSession.metadata).toMatchObject({
            title: 'Episode 1',
            artist: 'Show',
            artwork: [{ src: '/thumbnails/7', sizes: '640x640' }],
        });
    });

    it('routes play and pause only when the state differs', () => {
        const paused = buildOptions({ isPlaying: false });
        const { unmount } = renderHook(() => useVideoMediaSession(paused));
        handlers.play?.({ action: 'play' });
        handlers.pause?.({ action: 'pause' });
        expect(paused.togglePlayPause).toHaveBeenCalledTimes(1);
        unmount();

        const playing = buildOptions();
        renderHook(() => useVideoMediaSession(playing));
        handlers.pause?.({ action: 'pause' });
        handlers.play?.({ action: 'play' });
        expect(playing.togglePlayPause).toHaveBeenCalledTimes(1);
    });

    it('seeks relatively with default and custom offsets, clamped', () => {
        const options = buildOptions();
        renderHook(() => useVideoMediaSession(options));

        handlers.seekbackward?.({ action: 'seekbackward' });
        handlers.seekforward?.({ action: 'seekforward', seekOffset: 5 });
        handlers.seekforward?.({ action: 'seekforward', seekOffset: 500 });
        handlers.seekbackward?.({ action: 'seekbackward', seekOffset: 500 });

        expect(options.seekTo.mock.calls).toEqual([[20], [35], [100], [0]]);
    });

    it('seeks to an absolute time and ignores missing seekTime', () => {
        const options = buildOptions();
        renderHook(() => useVideoMediaSession(options));

        handlers.seekto?.({ action: 'seekto', seekTime: 42 });
        handlers.seekto?.({ action: 'seekto' });

        expect(options.seekTo.mock.calls).toEqual([[42]]);
    });

    it('wires previous and next when available', () => {
        const options = buildOptions();
        renderHook(() => useVideoMediaSession(options));

        handlers.previoustrack?.({ action: 'previoustrack' });
        handlers.nexttrack?.({ action: 'nexttrack' });

        expect(options.previousVideo).toHaveBeenCalledTimes(1);
        expect(options.nextVideo).toHaveBeenCalledTimes(1);
    });

    it('registers null for previous and next without a playlist neighbour', () => {
        renderHook(() =>
            useVideoMediaSession(buildOptions({ hasPreviousVideo: false, hasNextVideo: false }))
        );

        expect(handlers.previoustrack).toBeNull();
        expect(handlers.nexttrack).toBeNull();
    });

    it('reflects playback state and position', () => {
        const { rerender } = renderHook((props) => useVideoMediaSession(props), {
            initialProps: buildOptions(),
        });
        expect(mediaSession.playbackState).toBe('playing');
        expect(mediaSession.setPositionState).toHaveBeenCalledWith({
            duration: 100,
            playbackRate: 1,
            position: 30,
        });

        rerender(buildOptions({ isPlaying: false }));
        expect(mediaSession.playbackState).toBe('paused');
    });

    it('skips position state for unknown duration', () => {
        renderHook(() => useVideoMediaSession(buildOptions({ duration: NaN })));
        renderHook(() => useVideoMediaSession(buildOptions({ duration: 0 })));

        expect(mediaSession.setPositionState).not.toHaveBeenCalled();
    });

    it('swallows invalid position state errors', () => {
        mediaSession.setPositionState.mockImplementation(() => {
            throw new TypeError('invalid');
        });

        expect(() => renderHook(() => useVideoMediaSession(buildOptions()))).not.toThrow();
    });

    it('clears everything on unmount', () => {
        const { unmount } = renderHook(() => useVideoMediaSession(buildOptions()));
        unmount();

        expect(mediaSession.metadata).toBeNull();
        expect(mediaSession.playbackState).toBe('none');
        [
            'play',
            'pause',
            'seekbackward',
            'seekforward',
            'seekto',
            'previoustrack',
            'nexttrack',
        ].forEach((action) => expect(handlers[action]).toBeNull());
    });

    it('does not publish metadata without a current video', () => {
        renderHook(() => useVideoMediaSession(buildOptions({ currentVideo: null })));

        expect(mediaSession.metadata).toBeNull();
        expect(mediaSession.setPositionState).not.toHaveBeenCalled();
    });

    it('tolerates a missing media session API', () => {
        delete (navigator as any).mediaSession;

        expect(() => renderHook(() => useVideoMediaSession(buildOptions()))).not.toThrow();
    });

    it('tolerates setActionHandler throwing for unsupported actions', () => {
        mediaSession.setActionHandler.mockImplementation(() => {
            throw new TypeError('unsupported');
        });

        const { unmount } = renderHook(() => useVideoMediaSession(buildOptions()));
        expect(() => unmount()).not.toThrow();
    });
});

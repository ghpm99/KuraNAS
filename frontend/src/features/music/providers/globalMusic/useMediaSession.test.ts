import { renderHook } from '@testing-library/react';
import useMediaSession from './useMediaSession';
import type { IMusicData } from '../musicProvider/musicProvider';

const buildOptions = (currentTrack: IMusicData | undefined) => ({
    currentTrack,
    isPlaying: false,
    onPlay: jest.fn(),
    onPause: jest.fn(),
    onNext: jest.fn(),
    onPrevious: jest.fn(),
    onSeekTo: jest.fn(),
    currentTime: 0,
    duration: 0,
});

describe('useMediaSession', () => {
    const originalMediaMetadata = (globalThis as { MediaMetadata?: unknown }).MediaMetadata;

    afterEach(() => {
        (globalThis as { MediaMetadata?: unknown }).MediaMetadata = originalMediaMetadata;
        delete (navigator as { mediaSession?: unknown }).mediaSession;
    });

    it('does nothing when the browser has no media session', () => {
        expect(() => renderHook(() => useMediaSession(buildOptions(undefined)))).not.toThrow();
    });

    it('publishes the track cover as artwork in three sizes', () => {
        const mediaSession = {
            metadata: null as unknown,
            playbackState: 'none',
            setActionHandler: jest.fn(),
        };
        Object.defineProperty(navigator, 'mediaSession', {
            value: mediaSession,
            configurable: true,
        });
        (globalThis as { MediaMetadata?: unknown }).MediaMetadata = class {
            constructor(init: object) {
                Object.assign(this, init);
            }
        };
        const track = { id: 42, name: 'song.mp3', metadata: { title: 'Song' } } as IMusicData;

        const { unmount } = renderHook(() => useMediaSession(buildOptions(track)));

        const publishedMetadata = mediaSession.metadata as { title: string; artwork: MediaImage[] };
        expect(publishedMetadata.title).toBe('Song');
        expect(publishedMetadata.artwork.map((artwork) => artwork.sizes)).toEqual([
            '96x96',
            '256x256',
            '512x512',
        ]);
        expect(publishedMetadata.artwork[0]?.src).toBe('/api/v1/music/tracks/42/cover?size=96');
        unmount();
    });

    describe('action handlers', () => {
        const installMediaSession = () => {
            const handlers: Record<string, ((details: MediaSessionActionDetails) => void) | null> = {};
            const setActionHandler = jest.fn(
                (action: string, handler: ((details: MediaSessionActionDetails) => void) | null) => {
                    handlers[action] = handler;
                }
            );
            Object.defineProperty(navigator, 'mediaSession', {
                value: { metadata: null, playbackState: 'none', setActionHandler },
                configurable: true,
            });
            return { handlers, setActionHandler };
        };

        const trigger = (
            handlers: Record<string, ((details: MediaSessionActionDetails) => void) | null>,
            action: string,
            details: Partial<MediaSessionActionDetails> = {}
        ) => handlers[action]?.({ action, ...details } as MediaSessionActionDetails);

        it('seeks backward by the default 10 seconds, clamped to zero', () => {
            const { handlers } = installMediaSession();
            const options = { ...buildOptions(undefined), currentTime: 25, duration: 100 };
            renderHook(() => useMediaSession(options));

            trigger(handlers, 'seekbackward');
            expect(options.onSeekTo).toHaveBeenLastCalledWith(15);

            trigger(handlers, 'seekbackward', { seekOffset: 60 });
            expect(options.onSeekTo).toHaveBeenLastCalledWith(0);
        });

        it('seeks forward using the provided offset, clamped to the duration', () => {
            const { handlers } = installMediaSession();
            const options = { ...buildOptions(undefined), currentTime: 90, duration: 100 };
            renderHook(() => useMediaSession(options));

            trigger(handlers, 'seekforward', { seekOffset: 5 });
            expect(options.onSeekTo).toHaveBeenLastCalledWith(95);

            trigger(handlers, 'seekforward');
            expect(options.onSeekTo).toHaveBeenLastCalledWith(100);
        });

        it('seeks forward without upper bound while the duration is unknown', () => {
            const { handlers } = installMediaSession();
            const options = { ...buildOptions(undefined), currentTime: 5, duration: 0 };
            renderHook(() => useMediaSession(options));

            trigger(handlers, 'seekforward');
            expect(options.onSeekTo).toHaveBeenLastCalledWith(15);
        });

        it('stop pauses a playing track and rewinds to zero', () => {
            const { handlers } = installMediaSession();
            const options = { ...buildOptions(undefined), isPlaying: true, currentTime: 30, duration: 100 };
            renderHook(() => useMediaSession(options));

            trigger(handlers, 'stop');

            expect(options.onPause).toHaveBeenCalledTimes(1);
            expect(options.onSeekTo).toHaveBeenCalledWith(0);
        });

        it('stop does not toggle playback when already paused', () => {
            const { handlers } = installMediaSession();
            const options = buildOptions(undefined);
            renderHook(() => useMediaSession(options));

            trigger(handlers, 'stop');

            expect(options.onPause).not.toHaveBeenCalled();
            expect(options.onSeekTo).toHaveBeenCalledWith(0);
        });

        it('seekto forwards the requested time', () => {
            const { handlers } = installMediaSession();
            const options = buildOptions(undefined);
            renderHook(() => useMediaSession(options));

            trigger(handlers, 'seekto', { seekTime: 42 });
            trigger(handlers, 'seekto');

            expect(options.onSeekTo).toHaveBeenCalledTimes(1);
            expect(options.onSeekTo).toHaveBeenCalledWith(42);
        });

        it('routes play, pause, next and previous to the callbacks', () => {
            const { handlers } = installMediaSession();
            const options = buildOptions(undefined);
            renderHook(() => useMediaSession(options));

            trigger(handlers, 'play');
            trigger(handlers, 'pause');
            trigger(handlers, 'nexttrack');
            trigger(handlers, 'previoustrack');

            expect(options.onPlay).toHaveBeenCalledTimes(1);
            expect(options.onPause).toHaveBeenCalledTimes(1);
            expect(options.onNext).toHaveBeenCalledTimes(1);
            expect(options.onPrevious).toHaveBeenCalledTimes(1);
        });

        it('tolerates browsers that throw for unsupported actions', () => {
            const { handlers, setActionHandler } = installMediaSession();
            setActionHandler.mockImplementation((action, handler) => {
                if (action === 'seekbackward' || action === 'stop') {
                    throw new TypeError('unsupported action');
                }
                handlers[action] = handler;
            });
            const options = buildOptions(undefined);

            const { unmount } = renderHook(() => useMediaSession(options));
            trigger(handlers, 'play');

            expect(options.onPlay).toHaveBeenCalledTimes(1);
            expect(handlers.seekforward).toBeTruthy();
            expect(() => unmount()).not.toThrow();
        });

        it('clears every handler on unmount', () => {
            const { handlers } = installMediaSession();
            const { unmount } = renderHook(() => useMediaSession(buildOptions(undefined)));
            const registeredActions = Object.keys(handlers);

            unmount();

            expect(registeredActions).toEqual(
                expect.arrayContaining([
                    'play',
                    'pause',
                    'nexttrack',
                    'previoustrack',
                    'seekto',
                    'seekbackward',
                    'seekforward',
                    'stop',
                ])
            );
            registeredActions.forEach((action) => expect(handlers[action]).toBeNull());
        });
    });
});

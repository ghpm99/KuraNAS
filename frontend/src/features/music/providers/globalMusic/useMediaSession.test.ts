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
});

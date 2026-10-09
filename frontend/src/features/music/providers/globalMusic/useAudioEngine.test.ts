import { act, renderHook } from '@testing-library/react';
import useAudioEngine from './useAudioEngine';

class MockAudio {
    static instances: MockAudio[] = [];

    constructor() {
        MockAudio.instances.push(this);
    }

    src = '';
    currentTime = 0;
    duration = 0;
    ended = false;
    preload = 'none';
    paused = true;
    volume = 1;
    load = jest.fn();
    listeners: Record<string, Function[]> = {};

    static reset() {
        MockAudio.instances = [];
    }

    addEventListener(event: string, callback: Function) {
        this.listeners[event] = this.listeners[event] ?? [];
        this.listeners[event].push(callback);
    }

    removeEventListener(event: string, callback: Function) {
        this.listeners[event] = (this.listeners[event] ?? []).filter((fn) => fn !== callback);
    }

    play() {
        this.paused = false;
        this.trigger('play');
        return Promise.resolve();
    }

    pause() {
        this.paused = true;
        this.trigger('pause');
    }

    trigger(event: string) {
        (this.listeners[event] ?? []).forEach((fn) => fn());
    }
}

const getMainAudio = () => MockAudio.instances[0]!;
const getPreloadAudio = () => MockAudio.instances[1]!;

describe('useAudioEngine', () => {
    let originalAudio: typeof Audio;

    beforeAll(() => {
        originalAudio = globalThis.Audio;
        (globalThis as any).Audio = MockAudio;
    });

    beforeEach(() => {
        MockAudio.reset();
    });

    afterAll(() => {
        globalThis.Audio = originalAudio;
    });

    it('exposes control helpers and respects clamped volume', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });
        act(() => {
            result.current.loadAndPlayUrl('http://example.com/test.mp3');
        });
        expect(getMainAudio().src).toBe('http://example.com/test.mp3');

        act(() => {
            result.current.setVolume(1.5);
        });
        expect(result.current.volume).toBe(1);

        act(() => {
            result.current.setVolume(-1);
        });
        expect(result.current.volume).toBe(0);

        act(() => {
            result.current.seek(42);
        });

        act(() => {
            result.current.togglePlayPause();
        });
        await act(async () => {
            await Promise.resolve();
        });
        expect(getMainAudio().paused).toBe(true);

        act(() => {
            result.current.togglePlayPause();
        });
        await act(async () => {
            await Promise.resolve();
        });
        expect(getMainAudio().paused).toBe(false);

        act(() => {
            result.current.stop();
        });
        expect(result.current.currentTime).toBe(0);
        expect(result.current.isPlaying).toBe(false);
    });

    it('stop resets src and pauses the audio element', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        act(() => {
            result.current.loadAndPlayUrl('http://example.com/song.mp3');
        });
        expect(getMainAudio().src).toBe('http://example.com/song.mp3');

        act(() => {
            result.current.stop();
        });
        expect(getMainAudio().src).toBe('');
        expect(getMainAudio().paused).toBe(true);
        expect(result.current.isPlaying).toBe(false);
        expect(result.current.currentTime).toBe(0);
        expect(result.current.duration).toBe(0);
    });

    it('loadAndPlayUrl sets src and calls play', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        act(() => {
            result.current.loadAndPlayUrl('http://example.com/track.mp3');
        });
        await act(async () => {
            await Promise.resolve();
        });

        expect(getMainAudio().src).toBe('http://example.com/track.mp3');
        expect(getMainAudio().paused).toBe(false);
        expect(result.current.isPlaying).toBe(true);
    });

    it('setVolume clamps value and sets it on the audio element', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        act(() => {
            result.current.setVolume(0.5);
        });
        expect(result.current.volume).toBe(0.5);
        expect(getMainAudio().volume).toBe(0.5);

        act(() => {
            result.current.setVolume(2);
        });
        expect(result.current.volume).toBe(1);
        expect(getMainAudio().volume).toBe(1);

        act(() => {
            result.current.setVolume(-0.5);
        });
        expect(result.current.volume).toBe(0);
        expect(getMainAudio().volume).toBe(0);
    });

    it('seek sets currentTime on the audio element', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        act(() => {
            result.current.seek(99);
        });
        expect(getMainAudio().currentTime).toBe(99);
    });

    it('timeupdate event updates currentTime state', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        const audio = getMainAudio();
        act(() => {
            audio.currentTime = 15.5;
            audio.trigger('timeupdate');
        });
        expect(result.current.currentTime).toBe(15.5);
    });

    it('loadedmetadata event updates duration state', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        const audio = getMainAudio();
        act(() => {
            audio.duration = 240;
            audio.trigger('loadedmetadata');
        });
        expect(result.current.duration).toBe(240);
    });

    it('ended event calls onTrackEnded callback', async () => {
        const onEnded = jest.fn();
        renderHook(() => useAudioEngine(onEnded));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        const audio = getMainAudio();
        act(() => {
            audio.trigger('ended');
        });
        expect(onEnded).toHaveBeenCalledTimes(1);
    });

    it('pause event sets isPlaying to false', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        act(() => {
            result.current.loadAndPlayUrl('http://example.com/test.mp3');
        });
        await act(async () => {
            await Promise.resolve();
        });
        expect(result.current.isPlaying).toBe(true);

        const audio = getMainAudio();
        act(() => {
            audio.pause();
        });
        expect(result.current.isPlaying).toBe(false);
    });

    it('play event sets isPlaying to true', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        expect(result.current.isPlaying).toBe(false);
        const audio = getMainAudio();
        act(() => {
            audio.play();
        });
        await act(async () => {
            await Promise.resolve();
        });
        expect(result.current.isPlaying).toBe(true);
    });

    it('handles all operations gracefully when audioRef is null', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        (result.current.audioRef as any).current = null;

        act(() => {
            result.current.togglePlayPause();
            result.current.loadAndPlayUrl('http://example.com/test.mp3');
            result.current.seek(10);
            result.current.setVolume(0.3);
            result.current.stop();
        });

        expect(result.current.volume).toBe(0.3);
        expect(result.current.isPlaying).toBe(false);
        expect(result.current.currentTime).toBe(0);
        expect(result.current.duration).toBe(0);
    });

    it('updates onTrackEnded ref when callback changes', async () => {
        const first = jest.fn();
        const second = jest.fn();

        const { rerender } = renderHook(({ cb }) => useAudioEngine(cb), {
            initialProps: { cb: first },
        });
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        rerender({ cb: second });

        const audio = getMainAudio();
        act(() => {
            audio.trigger('ended');
        });

        expect(first).not.toHaveBeenCalled();
        expect(second).toHaveBeenCalledTimes(1);
    });

    it('preloadUrl preloads once and ignores duplicate url', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        const preloadAudio = getPreloadAudio();
        act(() => {
            result.current.preloadUrl('http://example.com/preload.mp3');
        });

        expect(preloadAudio.src).toBe('http://example.com/preload.mp3');
        expect(preloadAudio.load).toHaveBeenCalledTimes(1);

        act(() => {
            result.current.preloadUrl('http://example.com/preload.mp3');
        });
        expect(preloadAudio.load).toHaveBeenCalledTimes(1);
    });

    it('error event triggers onTrackEnded once when source exists', async () => {
        const onEnded = jest.fn();
        const { result } = renderHook(() => useAudioEngine(onEnded));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        act(() => {
            result.current.loadAndPlayUrl('http://example.com/error.mp3');
        });

        const audio = getMainAudio();
        act(() => {
            audio.trigger('error');
            audio.trigger('error');
        });

        expect(onEnded).toHaveBeenCalledTimes(1);
    });

    it('fallback interval advances when ended is true in background playback', async () => {
        jest.useFakeTimers();
        const onEnded = jest.fn();
        const { result, unmount } = renderHook(() => useAudioEngine(onEnded));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        const audio = getMainAudio();
        act(() => {
            result.current.loadAndPlayUrl('http://example.com/fallback.mp3');
            audio.duration = 60;
            audio.ended = true;
        });

        act(() => {
            jest.advanceTimersByTime(500);
        });

        expect(onEnded).toHaveBeenCalledTimes(1);
        unmount();
        jest.useRealTimers();
    });

    it('loads a url without playing and seeks to the saved position on loadedmetadata', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
        });
        const audio = getMainAudio();
        const playSpy = jest.spyOn(audio, 'play');

        act(() => {
            result.current.loadUrlPaused('http://example.com/saved.mp3', 73);
        });

        expect(audio.src).toBe('http://example.com/saved.mp3');
        expect(playSpy).not.toHaveBeenCalled();
        expect(audio.paused).toBe(true);
        expect(audio.currentTime).toBe(0);
        expect(result.current.currentTime).toBe(73);

        act(() => {
            audio.trigger('loadedmetadata');
        });

        expect(audio.currentTime).toBe(73);
        expect(playSpy).not.toHaveBeenCalled();
    });

    it('does not register a seek when the saved position is zero', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
        });
        const audio = getMainAudio();

        act(() => {
            result.current.loadUrlPaused('http://example.com/start.mp3', 0);
            audio.currentTime = 5;
            audio.trigger('loadedmetadata');
        });

        expect(audio.currentTime).toBe(5);
    });

    it('ignores loadUrlPaused before the audio element exists', () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        const mountedAudio = getMainAudio();
        const originalSrc = mountedAudio.src;

        act(() => {
            result.current.audioRef.current = null;
            result.current.loadUrlPaused('http://example.com/ignored.mp3', 10);
        });

        expect(mountedAudio.src).toBe(originalSrc);
    });

    describe('transcoded streams', () => {
        const transcodedStream = {
            buildUrl: (startSeconds: number) =>
                `http://example.com/transcode?start=${startSeconds}`,
            fallbackUrl: 'http://example.com/raw',
            durationSeconds: 200,
        };

        it('reports metadata duration and offsets currentTime by the restart position', () => {
            const { result } = renderHook(() => useAudioEngine(() => {}));
            act(() => {
                result.current.loadAndPlayUrl('http://example.com/transcode', transcodedStream);
            });
            const audio = getMainAudio();
            act(() => {
                audio.duration = Infinity;
                audio.trigger('loadedmetadata');
            });
            expect(result.current.duration).toBe(200);

            act(() => {
                result.current.seek(60);
            });
            expect(audio.src).toBe('http://example.com/transcode?start=60');
            expect(result.current.currentTime).toBe(60);

            act(() => {
                audio.currentTime = 5;
                audio.trigger('timeupdate');
            });
            expect(result.current.currentTime).toBe(65);
            expect(result.current.getPositionSeconds()).toBe(65);
        });

        it('keeps a paused transcoded track paused when seeking', () => {
            const { result } = renderHook(() => useAudioEngine(() => {}));
            act(() => {
                result.current.loadAndPlayUrl('http://example.com/transcode', transcodedStream);
                getMainAudio().pause();
                result.current.seek(30);
            });
            expect(getMainAudio().paused).toBe(true);
            expect(getMainAudio().src).toBe('http://example.com/transcode?start=30');
        });

        it('resumes playing after seeking a playing transcoded track', async () => {
            const { result } = renderHook(() => useAudioEngine(() => {}));
            act(() => {
                result.current.loadAndPlayUrl('http://example.com/transcode', transcodedStream);
            });
            expect(getMainAudio().paused).toBe(false);
            act(() => {
                result.current.seek(10);
            });
            expect(getMainAudio().paused).toBe(false);
        });

        it('restores a saved position by rebuilding the transcode URL without autoplay', () => {
            const { result } = renderHook(() => useAudioEngine(() => {}));
            act(() => {
                result.current.loadUrlPaused('http://example.com/transcode', 42, transcodedStream);
            });
            expect(getMainAudio().src).toBe('http://example.com/transcode?start=42');
            expect(getMainAudio().paused).toBe(true);
            expect(result.current.currentTime).toBe(42);
            expect(result.current.duration).toBe(200);
        });

        it('falls back to the raw stream when the transcode fails', () => {
            const onTrackEnded = jest.fn();
            const { result } = renderHook(() => useAudioEngine(onTrackEnded));
            act(() => {
                result.current.loadAndPlayUrl('http://example.com/transcode', transcodedStream);
                result.current.seek(20);
            });
            act(() => {
                getMainAudio().trigger('error');
            });
            expect(getMainAudio().src).toBe('http://example.com/raw');
            expect(onTrackEnded).not.toHaveBeenCalled();
            expect(result.current.getPositionSeconds()).toBe(0);
            act(() => {
                getMainAudio().trigger('error');
            });
            expect(onTrackEnded).toHaveBeenCalledTimes(1);
        });

        it('seeks a direct stream by setting currentTime without offset', () => {
            const { result } = renderHook(() => useAudioEngine(() => {}));
            act(() => {
                result.current.loadAndPlayUrl('http://example.com/raw');
                result.current.seek(33);
            });
            expect(getMainAudio().currentTime).toBe(33);
            expect(getMainAudio().src).toBe('http://example.com/raw');
        });

        it('assumes playable when the element cannot answer canPlayType', () => {
            const { result } = renderHook(() => useAudioEngine(() => {}));
            expect(result.current.canPlayType('audio/mpeg')).toBe('maybe');
        });
    });
});

describe('useAudioEngine consecutive playback failures', () => {
    let originalAudio: typeof Audio;

    beforeAll(() => {
        originalAudio = globalThis.Audio;
        (globalThis as any).Audio = MockAudio;
    });

    beforeEach(() => {
        MockAudio.reset();
    });

    afterAll(() => {
        globalThis.Audio = originalAudio;
    });

    const failTrack = (engine: { loadAndPlayUrl: (url: string) => void }, url: string) => {
        act(() => {
            engine.loadAndPlayUrl(url);
        });
        act(() => {
            getMainAudio().trigger('error');
        });
    };

    it('does not crash when no failure callback is provided', () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        for (let attempt = 0; attempt < 4; attempt += 1) {
            failTrack(result.current, `http://example.com/${attempt}.mp3`);
        }
        expect(getMainAudio().src).toBe('http://example.com/3.mp3');
    });

    it('stops advancing and reports failure after three consecutive errors', () => {
        const onTrackEnded = jest.fn();
        const onPlaybackFailure = jest.fn();
        const { result } = renderHook(() => useAudioEngine(onTrackEnded, onPlaybackFailure));

        failTrack(result.current, 'http://example.com/1.mp3');
        failTrack(result.current, 'http://example.com/2.mp3');
        expect(onTrackEnded).toHaveBeenCalledTimes(2);
        expect(onPlaybackFailure).not.toHaveBeenCalled();

        failTrack(result.current, 'http://example.com/3.mp3');

        expect(onTrackEnded).toHaveBeenCalledTimes(2);
        expect(onPlaybackFailure).toHaveBeenCalledTimes(1);
        expect(getMainAudio().paused).toBe(true);
    });

    it('resets the failure counter when a track starts playing', () => {
        const onTrackEnded = jest.fn();
        const onPlaybackFailure = jest.fn();
        const { result } = renderHook(() => useAudioEngine(onTrackEnded, onPlaybackFailure));

        failTrack(result.current, 'http://example.com/1.mp3');
        failTrack(result.current, 'http://example.com/2.mp3');
        act(() => {
            getMainAudio().trigger('playing');
        });
        failTrack(result.current, 'http://example.com/3.mp3');
        failTrack(result.current, 'http://example.com/4.mp3');

        expect(onPlaybackFailure).not.toHaveBeenCalled();
        expect(onTrackEnded).toHaveBeenCalledTimes(4);
    });

    it('allows three more attempts after a failure was reported', () => {
        const onPlaybackFailure = jest.fn();
        const { result } = renderHook(() => useAudioEngine(() => {}, onPlaybackFailure));

        for (let attempt = 0; attempt < 5; attempt += 1) {
            failTrack(result.current, `http://example.com/${attempt}.mp3`);
        }

        expect(onPlaybackFailure).toHaveBeenCalledTimes(1);
    });
});

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
});

import { act, renderHook } from '@testing-library/react';
import useAudioEngine from './useAudioEngine';

class MockAudio {
    static lastInstance: MockAudio | null = null;
    constructor() {
        MockAudio.lastInstance = this;
    }
    src = '';
    currentTime = 0;
    duration = 0;
    paused = true;
    volume = 1;
    listeners: Record<string, Function[]> = {};

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

describe('useAudioEngine', () => {
    let originalAudio: typeof Audio;

    beforeAll(() => {
        originalAudio = globalThis.Audio;
        (globalThis as any).Audio = MockAudio;
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
        expect(MockAudio.lastInstance?.src).toBe('http://example.com/test.mp3');

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
        expect(MockAudio.lastInstance?.paused).toBe(true);

        act(() => {
            result.current.togglePlayPause();
        });
        await act(async () => {
            await Promise.resolve();
        });
        expect(MockAudio.lastInstance?.paused).toBe(false);

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
        expect(MockAudio.lastInstance?.src).toBe('http://example.com/song.mp3');

        act(() => {
            result.current.stop();
        });
        expect(MockAudio.lastInstance?.src).toBe('');
        expect(MockAudio.lastInstance?.paused).toBe(true);
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

        expect(MockAudio.lastInstance?.src).toBe('http://example.com/track.mp3');
        expect(MockAudio.lastInstance?.paused).toBe(false);
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
        expect(MockAudio.lastInstance?.volume).toBe(0.5);

        act(() => {
            result.current.setVolume(2);
        });
        expect(result.current.volume).toBe(1);
        expect(MockAudio.lastInstance?.volume).toBe(1);

        act(() => {
            result.current.setVolume(-0.5);
        });
        expect(result.current.volume).toBe(0);
        expect(MockAudio.lastInstance?.volume).toBe(0);
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
        expect(MockAudio.lastInstance?.currentTime).toBe(99);
    });

    it('timeupdate event updates currentTime state', async () => {
        const { result } = renderHook(() => useAudioEngine(() => {}));
        await act(async () => {
            await Promise.resolve();
            await Promise.resolve();
        });

        const audio = MockAudio.lastInstance!;
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

        const audio = MockAudio.lastInstance!;
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

        const audio = MockAudio.lastInstance!;
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

        // First play, then pause
        act(() => {
            result.current.loadAndPlayUrl('http://example.com/test.mp3');
        });
        await act(async () => {
            await Promise.resolve();
        });
        expect(result.current.isPlaying).toBe(true);

        const audio = MockAudio.lastInstance!;
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
        const audio = MockAudio.lastInstance!;
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

        const audio = MockAudio.lastInstance!;
        act(() => {
            audio.trigger('ended');
        });

        expect(first).not.toHaveBeenCalled();
        expect(second).toHaveBeenCalledTimes(1);
    });
});

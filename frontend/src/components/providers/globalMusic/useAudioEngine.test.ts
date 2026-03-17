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
		expect(MockAudio.lastInstance?.paused).toBe(false);

		act(() => {
			result.current.togglePlayPause();
		});
		await act(async () => {
			await Promise.resolve();
		});
		expect(MockAudio.lastInstance?.paused).toBe(true);

		act(() => {
			result.current.stop();
		});
		expect(result.current.currentTime).toBe(0);
		expect(result.current.isPlaying).toBe(false);
	});
});

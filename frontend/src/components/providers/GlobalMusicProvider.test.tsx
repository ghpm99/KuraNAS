import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { type IMusicData } from '../musicProvider/musicProvider';
import { GlobalMusicProvider, useGlobalMusic } from './GlobalMusicProvider';

const fakeSettings = {
	players: {
		remember_music_queue: true,
	},
};

const createEngineMock = () => ({
	audioRef: { current: { currentTime: 0, play: jest.fn().mockResolvedValue(undefined) } as { currentTime: number; play: jest.Mock } | null },
	loadAndPlayUrl: jest.fn(),
	togglePlayPause: jest.fn(),
	seek: jest.fn(),
	setVolume: jest.fn(),
	stop: jest.fn(),
	isPlaying: false,
	currentTime: 0,
	duration: 0,
	volume: 1,
});

let engineMock = createEngineMock();
const mockSyncState = jest.fn();
const mockQueueHydration = jest.fn();
let capturedOnTrackEnded: (() => void) | undefined;

jest.mock('./globalMusic/useAudioEngine', () => ({
	__esModule: true,
	default: (onTrackEnded: () => void) => {
		capturedOnTrackEnded = onTrackEnded;
		return engineMock;
	},
}));

type SyncDeps = {
	getCurrentTrackId: () => number | undefined;
	getCurrentTime: () => number;
};
let capturedSyncDeps: SyncDeps | undefined;

jest.mock('./globalMusic/useMusicStateSync', () => ({
	__esModule: true,
	default: (deps: SyncDeps) => {
		capturedSyncDeps = deps;
		return { syncState: mockSyncState };
	},
}));

jest.mock('./globalMusic/useMusicQueueHydration', () => ({
	__esModule: true,
	default: (enabled: boolean, callbacks: unknown) => mockQueueHydration(enabled, callbacks),
}));

jest.mock('./settingsProvider/settingsContext', () => ({
	__esModule: true,
	useSettings: () => ({
		settings: fakeSettings,
		isLoading: false,
		isSaving: false,
		hasError: false,
		refresh: jest.fn(),
		saveSettings: jest.fn(),
	}),
}));

const wrapper = ({ children }: { children?: ReactNode }) => (
	<GlobalMusicProvider>{children}</GlobalMusicProvider>
);

const createTrack = (id: number): IMusicData => ({
	id,
	name: `track-${id}`,
	path: `/tracks/${id}.mp3`,
	type: id,
	format: 'mp3',
	size: 1024,
	updated_at: '',
	created_at: '',
	deleted_at: '',
	last_interaction: '',
	last_backup: '',
	check_sum: '',
	directory_content_count: 0,
	starred: false,
});

describe('GlobalMusicProvider', () => {
	beforeEach(() => {
		jest.useFakeTimers();
		engineMock = createEngineMock();
		mockSyncState.mockReset();
		mockQueueHydration.mockReset();
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('manages queue operations and shuffle/previous flows', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const trackA = createTrack(1);
		const trackB = createTrack(2);
		const trackC = createTrack(3);

		act(() => {
			result.current.addToQueue(trackA, { playlistId: 10 });
		});
		act(() => {
			jest.runOnlyPendingTimers();
		});
		expect(result.current.queue).toEqual([trackA]);
		expect(result.current.currentIndex).toBe(0);
		expect(result.current.playbackContext).toEqual({ playlistId: 10 });

		act(() => {
			result.current.addToQueue(trackA);
			result.current.addToQueue(trackB);
		});
		act(() => {
			jest.runOnlyPendingTimers();
		});
		expect(result.current.queue).toHaveLength(2);

		act(() => {
			result.current.replaceQueue([trackB, trackC], 1, { playlistId: 22 });
		});
		expect(result.current.queue[1]).toEqual(trackC);
		expect(result.current.currentIndex).toBe(1);
		expect(engineMock.loadAndPlayUrl).toHaveBeenLastCalledWith(expect.stringContaining('/files/stream/3'));
		expect(mockSyncState).toHaveBeenCalledWith(
			expect.objectContaining({ fileId: 3, position: 0, playlistId: 22 }),
		);

		act(() => {
			result.current.playTrackFromQueue(0);
		});
		expect(result.current.currentIndex).toBe(0);

		act(() => {
			result.current.toggleShuffle();
		});
		const mathSpy = jest.spyOn(Math, 'random').mockReturnValue(0);
		act(() => {
			result.current.next();
		});
		expect(engineMock.loadAndPlayUrl).toHaveBeenCalled();
		mathSpy.mockRestore();

		engineMock.audioRef.current!.currentTime = 5;
		const previousCalls = engineMock.loadAndPlayUrl.mock.calls.length;
		act(() => {
			result.current.previous();
		});
		expect(engineMock.audioRef.current!.currentTime).toBe(0);
		expect(engineMock.loadAndPlayUrl.mock.calls.length).toBe(previousCalls);

		engineMock.audioRef.current!.currentTime = 1;
		act(() => {
			result.current.previous();
		});
		expect(engineMock.loadAndPlayUrl.mock.calls.length).toBe(previousCalls + 1);
	});

	it('exposes helpers and resets queue correctly', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const track = createTrack(5);

		act(() => {
			result.current.replaceQueue([], 0);
		});
		expect(result.current.queue).toEqual([]);

		act(() => {
			result.current.addToQueue(track);
		});
		act(() => {
			jest.runOnlyPendingTimers();
		});
		expect(result.current.hasQueue).toBe(true);

		act(() => {
			result.current.togglePlayPause();
			result.current.seek(12);
			result.current.setVolume(0.25);
			result.current.toggleQueue();
		});
		expect(engineMock.togglePlayPause).toHaveBeenCalled();
		expect(engineMock.seek).toHaveBeenCalledWith(12);
		expect(engineMock.setVolume).toHaveBeenCalledWith(0.25);
		expect(result.current.queueOpen).toBe(true);
		expect(mockSyncState).toHaveBeenCalledWith(expect.objectContaining({ position: 12 }));

		act(() => {
			result.current.toggleQueue();
			result.current.clearQueue();
		});
		expect(result.current.queue).toEqual([]);
		expect(result.current.currentIndex).toBeUndefined();
		expect(result.current.hasQueue).toBe(false);
		expect(engineMock.stop).toHaveBeenCalled();
	});

	it('removeFromQueue adjusts currentIndex when removing before current', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const tracks = [createTrack(1), createTrack(2), createTrack(3)];

		act(() => {
			result.current.replaceQueue(tracks, 2);
		});
		expect(result.current.currentIndex).toBe(2);

		// Remove track before current index: index should shift down by 1
		act(() => {
			result.current.removeFromQueue(0);
		});
		expect(result.current.currentIndex).toBe(1);
		expect(result.current.queue).toHaveLength(2);
	});

	it('removeFromQueue plays next track when removing the current track', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const tracks = [createTrack(1), createTrack(2), createTrack(3)];

		act(() => {
			result.current.replaceQueue(tracks, 1);
		});
		engineMock.loadAndPlayUrl.mockClear();

		// Remove the currently playing track (index 1)
		act(() => {
			result.current.removeFromQueue(1);
		});
		// Should load the next track (track 3 is now at index 1)
		expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(expect.stringContaining('/files/stream/3'));
		expect(result.current.queue).toHaveLength(2);
	});

	it('removeFromQueue wraps currentIndex when removing current track at end', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const tracks = [createTrack(1), createTrack(2)];

		act(() => {
			result.current.replaceQueue(tracks, 1);
		});
		engineMock.loadAndPlayUrl.mockClear();

		// Remove the last track which is also the current track
		act(() => {
			result.current.removeFromQueue(1);
		});
		// currentIndex should clamp to newQueue.length - 1 = 0
		expect(result.current.currentIndex).toBe(0);
		expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(expect.stringContaining('/files/stream/1'));
	});

	it('removeFromQueue does not change currentIndex when removing after current', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const tracks = [createTrack(1), createTrack(2), createTrack(3)];

		act(() => {
			result.current.replaceQueue(tracks, 0);
		});

		act(() => {
			result.current.removeFromQueue(2);
		});
		expect(result.current.currentIndex).toBe(0);
		expect(result.current.queue).toHaveLength(2);
	});

	it('removeFromQueue clears everything when removing the last remaining track', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.replaceQueue([createTrack(1)], 0);
		});

		act(() => {
			result.current.removeFromQueue(0);
		});
		expect(result.current.queue).toHaveLength(0);
		expect(result.current.currentIndex).toBeUndefined();
		expect(engineMock.stop).toHaveBeenCalled();
	});

	it('next wraps around to index 0 when at end of queue without shuffle', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const tracks = [createTrack(1), createTrack(2)];

		act(() => {
			result.current.replaceQueue(tracks, 1);
		});
		engineMock.loadAndPlayUrl.mockClear();

		act(() => {
			result.current.next();
		});
		// (1 + 1) % 2 = 0
		expect(result.current.currentIndex).toBe(0);
		expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(expect.stringContaining('/files/stream/1'));
	});

	it('next does nothing when queue is empty', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.next();
		});
		expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
	});

	it('previous does nothing when queue is empty', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.previous();
		});
		expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
	});

	it('previous wraps to last track when at index 0 and currentTime <= 3s', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const tracks = [createTrack(1), createTrack(2), createTrack(3)];

		act(() => {
			result.current.replaceQueue(tracks, 0);
		});
		engineMock.loadAndPlayUrl.mockClear();
		engineMock.audioRef.current!.currentTime = 1;

		act(() => {
			result.current.previous();
		});
		// Should wrap to last track
		expect(result.current.currentIndex).toBe(2);
		expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(expect.stringContaining('/files/stream/3'));
	});

	it('setRepeatMode changes the repeat mode', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		expect(result.current.repeatMode).toBe('none');
		act(() => {
			result.current.setRepeatMode('all');
		});
		expect(result.current.repeatMode).toBe('all');
		act(() => {
			result.current.setRepeatMode('one');
		});
		expect(result.current.repeatMode).toBe('one');
	});

	it('toggleShuffle toggles shuffle state', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		expect(result.current.shuffle).toBe(false);
		act(() => {
			result.current.toggleShuffle();
		});
		expect(result.current.shuffle).toBe(true);
		act(() => {
			result.current.toggleShuffle();
		});
		expect(result.current.shuffle).toBe(false);
	});

	it('setQueueOpen controls queue visibility', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		expect(result.current.queueOpen).toBe(false);
		act(() => {
			result.current.setQueueOpen(true);
		});
		expect(result.current.queueOpen).toBe(true);
		act(() => {
			result.current.setQueueOpen(false);
		});
		expect(result.current.queueOpen).toBe(false);
	});

	it('addToQueue does not add duplicate tracks', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const track = createTrack(1);

		act(() => {
			result.current.addToQueue(track);
		});
		act(() => {
			jest.runOnlyPendingTimers();
		});
		act(() => {
			result.current.addToQueue(track);
		});
		act(() => {
			jest.runOnlyPendingTimers();
		});
		expect(result.current.queue).toHaveLength(1);
	});

	it('addToQueue does not overwrite playbackContext when queue is already playing', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.addToQueue(createTrack(1), { playlistId: 10 });
		});
		act(() => {
			jest.runOnlyPendingTimers();
		});
		expect(result.current.playbackContext).toEqual({ playlistId: 10 });

		// Adding another track with a different context should NOT change playbackContext
		act(() => {
			result.current.addToQueue(createTrack(2), { playlistId: 20 });
		});
		act(() => {
			jest.runOnlyPendingTimers();
		});
		expect(result.current.playbackContext).toEqual({ playlistId: 10 });
	});

	it('replaceQueue with empty array is a no-op', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.replaceQueue([createTrack(1)], 0);
		});
		engineMock.loadAndPlayUrl.mockClear();

		act(() => {
			result.current.replaceQueue([]);
		});
		// Queue should remain unchanged
		expect(result.current.queue).toHaveLength(1);
		expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
	});

	it('seek syncs position to backend', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.seek(30);
		});
		expect(engineMock.seek).toHaveBeenCalledWith(30);
		expect(mockSyncState).toHaveBeenCalledWith({ position: 30 });
	});

	it('setVolume syncs volume to backend', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.setVolume(0.5);
		});
		expect(engineMock.setVolume).toHaveBeenCalledWith(0.5);
		expect(mockSyncState).toHaveBeenCalledWith({ vol: 0.5 });
	});

	it('playTrackFromQueue ignores out-of-bounds index', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.replaceQueue([createTrack(1)], 0);
		});
		engineMock.loadAndPlayUrl.mockClear();

		act(() => {
			result.current.playTrackFromQueue(5);
		});
		expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();

		act(() => {
			result.current.playTrackFromQueue(-1);
		});
		expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
	});

	it('currentTrack returns the track at currentIndex', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });
		const tracks = [createTrack(1), createTrack(2)];

		expect(result.current.currentTrack).toBeUndefined();

		act(() => {
			result.current.replaceQueue(tracks, 1);
		});
		expect(result.current.currentTrack).toEqual(tracks[1]);
	});

	it('throws when useGlobalMusic is used outside provider', () => {
		// Suppress console.error for this test
		const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
		expect(() => {
			renderHook(() => useGlobalMusic());
		}).toThrow('useGlobalMusic must be used within a GlobalMusicProvider');
		consoleSpy.mockRestore();
	});

	it('sync deps callbacks return correct values', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		// With no current track, getCurrentTrackId returns undefined
		expect(capturedSyncDeps!.getCurrentTrackId()).toBeUndefined();
		// getCurrentTime returns 0 when audio has currentTime 0
		expect(capturedSyncDeps!.getCurrentTime()).toBe(0);

		// Set up a track
		const tracks = [createTrack(1), createTrack(2)];
		act(() => {
			result.current.replaceQueue(tracks, 0);
		});

		expect(capturedSyncDeps!.getCurrentTrackId()).toBe(1);

		// Simulate audioRef with currentTime
		engineMock.audioRef.current!.currentTime = 42;
		expect(capturedSyncDeps!.getCurrentTime()).toBe(42);

		// When audioRef.current is null, getCurrentTime returns 0
		engineMock.audioRef.current = null;
		expect(capturedSyncDeps!.getCurrentTime()).toBe(0);
	});

	describe('handleTrackEnded', () => {
		it('repeat one: restarts the current track', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });
			const tracks = [createTrack(1), createTrack(2)];

			act(() => {
				result.current.replaceQueue(tracks, 0);
			});
			act(() => {
				result.current.setRepeatMode('one');
			});

			engineMock.audioRef.current!.currentTime = 50;
			engineMock.loadAndPlayUrl.mockClear();

			act(() => {
				capturedOnTrackEnded!();
			});

			// Should restart current track, not load a new one
			expect(engineMock.audioRef.current!.currentTime).toBe(0);
		});

		it('repeat one with null audioRef does not throw', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });

			act(() => {
				result.current.replaceQueue([createTrack(1)], 0);
			});
			act(() => {
				result.current.setRepeatMode('one');
			});

			engineMock.audioRef.current = null;

			// Should not throw
			act(() => {
				capturedOnTrackEnded!();
			});
		});

		it('advances to next track when not at end of queue (repeat none)', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });
			const tracks = [createTrack(1), createTrack(2), createTrack(3)];

			act(() => {
				result.current.replaceQueue(tracks, 0);
			});
			engineMock.loadAndPlayUrl.mockClear();
			mockSyncState.mockClear();

			act(() => {
				capturedOnTrackEnded!();
			});

			expect(result.current.currentIndex).toBe(1);
			expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(expect.stringContaining('/files/stream/2'));
			expect(mockSyncState).toHaveBeenCalledWith(expect.objectContaining({ fileId: 2, position: 0 }));
		});

		it('stops at end of queue when repeat is none', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });
			const tracks = [createTrack(1), createTrack(2)];

			act(() => {
				result.current.replaceQueue(tracks, 1);
			});
			engineMock.loadAndPlayUrl.mockClear();
			engineMock.stop.mockClear();

			act(() => {
				capturedOnTrackEnded!();
			});

			expect(engineMock.stop).toHaveBeenCalled();
		});

		it('wraps to first track when repeat is all and at end of queue', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });
			const tracks = [createTrack(1), createTrack(2)];

			act(() => {
				result.current.replaceQueue(tracks, 1);
			});
			act(() => {
				result.current.setRepeatMode('all');
			});
			engineMock.loadAndPlayUrl.mockClear();
			mockSyncState.mockClear();

			act(() => {
				capturedOnTrackEnded!();
			});

			expect(result.current.currentIndex).toBe(0);
			expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(expect.stringContaining('/files/stream/1'));
			expect(mockSyncState).toHaveBeenCalledWith(expect.objectContaining({ fileId: 1, position: 0 }));
		});

		it('picks a shuffled track when shuffle is on', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });
			const tracks = [createTrack(1), createTrack(2), createTrack(3)];

			act(() => {
				result.current.replaceQueue(tracks, 0);
			});
			act(() => {
				result.current.toggleShuffle();
			});
			engineMock.loadAndPlayUrl.mockClear();
			mockSyncState.mockClear();

			const mathSpy = jest.spyOn(Math, 'random').mockReturnValue(0.99);
			act(() => {
				capturedOnTrackEnded!();
			});
			mathSpy.mockRestore();

			// Should have picked a random index != 0
			expect(engineMock.loadAndPlayUrl).toHaveBeenCalled();
			expect(mockSyncState).toHaveBeenCalledWith(expect.objectContaining({ position: 0 }));
		});

		it('does nothing when currentIndex is undefined', () => {
			renderHook(() => useGlobalMusic(), { wrapper });
			engineMock.loadAndPlayUrl.mockClear();
			engineMock.stop.mockClear();

			act(() => {
				capturedOnTrackEnded!();
			});

			expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
			expect(engineMock.stop).not.toHaveBeenCalled();
		});

		it('shuffle with single-track queue returns index 0 (getShuffledIndex <= 1)', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });

			act(() => {
				result.current.replaceQueue([createTrack(1)], 0);
			});
			act(() => {
				result.current.toggleShuffle();
			});
			engineMock.loadAndPlayUrl.mockClear();

			act(() => {
				capturedOnTrackEnded!();
			});

			// getShuffledIndex returns 0 for single-track queue
			expect(result.current.currentIndex).toBe(0);
			expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(expect.stringContaining('/files/stream/1'));
		});

		it('next with shuffle on single-track queue picks index 0', () => {
			const { result } = renderHook(() => useGlobalMusic(), { wrapper });

			act(() => {
				result.current.replaceQueue([createTrack(1)], 0);
			});
			act(() => {
				result.current.toggleShuffle();
			});
			engineMock.loadAndPlayUrl.mockClear();

			act(() => {
				result.current.next();
			});

			expect(result.current.currentIndex).toBe(0);
		});
	});
});

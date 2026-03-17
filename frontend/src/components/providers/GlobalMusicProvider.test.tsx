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
	audioRef: { current: { currentTime: 0 } as { currentTime: number } | null },
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

jest.mock('./globalMusic/useAudioEngine', () => ({
	__esModule: true,
	default: () => engineMock,
}));

jest.mock('./globalMusic/useMusicStateSync', () => ({
	__esModule: true,
	default: () => ({
		syncState: mockSyncState,
	}),
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
});

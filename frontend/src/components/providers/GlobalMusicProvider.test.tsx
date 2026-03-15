import { act, renderHook } from '@testing-library/react';
import { GlobalMusicProvider, useGlobalMusic } from './GlobalMusicProvider';
import { getMusicTitle, getMusicArtist, formatMusicDuration, musicMetadata } from '@/utils/music';

const mockUpdatePlayerState = jest.fn();
const mockGetPlayerState = jest.fn();
const mockGetApiV1BaseUrl = jest.fn();
const mockGetNowPlayingPlaylist = jest.fn();
const mockGetPlaylistTracks = jest.fn();
const mockUseSettings = jest.fn();

jest.mock('@/service/playerState', () => ({
	getPlayerState: (...args: any[]) => mockGetPlayerState(...args),
	updatePlayerState: (...args: any[]) => mockUpdatePlayerState(...args),
}));

jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => mockGetApiV1BaseUrl(),
}));

jest.mock('@/service/playlist', () => ({
	getNowPlayingPlaylist: (...args: any[]) => mockGetNowPlayingPlaylist(...args),
	getPlaylistTracks: (...args: any[]) => mockGetPlaylistTracks(...args),
}));

jest.mock('./settingsProvider/settingsContext', () => ({
	useSettings: () => mockUseSettings(),
}));

class FakeAudio {
	src = '';
	volume = 1;
	currentTime = 0;
	duration = 200;
	paused = true;
	play = jest.fn(() => {
		this.paused = false;
		this.emit('play');
		return Promise.resolve();
	});
	pause = jest.fn(() => {
		this.paused = true;
		this.emit('pause');
	});
	private listeners = new Map<string, Set<() => void>>();

	addEventListener(event: string, cb: () => void) {
		if (!this.listeners.has(event)) this.listeners.set(event, new Set());
		this.listeners.get(event)?.add(cb);
	}

	removeEventListener(event: string, cb: () => void) {
		this.listeners.get(event)?.delete(cb);
	}

	emit(event: string) {
		this.listeners.get(event)?.forEach((cb) => cb());
	}
}

describe('components/providers/GlobalMusicProvider', () => {
	const wrapper = ({ children }: { children: React.ReactNode }) => (
		<GlobalMusicProvider>{children}</GlobalMusicProvider>
	);

	const track1: any = {
		id: 1,
		name: 'song-1.mp3',
		format: 'mp3',
		size: 1024,
		metadata: { title: 'Song 1', artist: 'Artist 1', duration: 121 },
	};
	const track2: any = {
		id: 2,
		name: 'song-2.flac',
		format: 'flac',
		size: 2048,
		metadata: { duration: 240 },
	};

	let fakeAudio: FakeAudio;

	beforeEach(() => {
		jest.clearAllMocks();
		jest.useFakeTimers();
		fakeAudio = new FakeAudio();
		mockGetApiV1BaseUrl.mockReturnValue('http://localhost:8000/v1');
		mockGetPlayerState.mockResolvedValue({
			playlist_id: null,
			current_file_id: null,
			current_position: 0,
			volume: 1,
			shuffle: false,
			repeat_mode: 'none',
		});
		mockGetNowPlayingPlaylist.mockResolvedValue({
			id: null,
			track_count: 0,
		});
		mockGetPlaylistTracks.mockResolvedValue({ items: [] });
		mockUseSettings.mockReturnValue({
			settings: {
				players: {
					remember_music_queue: false,
				},
			},
			isLoading: false,
		});
		mockUpdatePlayerState.mockResolvedValue({});
		(global as any).Audio = jest.fn(() => fakeAudio);
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('manages queue, playback controls, helpers and sync state', async () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		expect(result.current.hasQueue).toBe(false);
		expect(getMusicTitle(track1)).toBe('Song 1');
		expect(getMusicArtist(track2)).toBe('Unknown Artist');
		expect(formatMusicDuration(65)).toBe('1:05');
		expect(musicMetadata(track1)).toContain('2:01');

		act(() => {
			result.current.addToQueue(track1, {
				kind: 'album',
				labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
				labelParams: { name: 'Album 1' },
				href: '/music/albums',
			});
			result.current.addToQueue(track2);
			result.current.addToQueue(track1);
		});
		act(() => {
			jest.advanceTimersByTime(10);
		});
		expect(result.current.queue.length).toBe(2);
		expect(result.current.currentIndex).toBe(0);
		expect(result.current.playbackContext).toMatchObject({
			labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
			labelParams: { name: 'Album 1' },
		});

		act(() => {
			result.current.playTrackFromQueue(1);
		});
		expect(fakeAudio.src).toContain('/files/stream/2');

		act(() => {
			result.current.togglePlayPause();
		});
		expect(fakeAudio.pause).toHaveBeenCalled();
		act(() => {
			fakeAudio.emit('play');
		});
		expect(result.current.isPlaying).toBe(true);

		act(() => {
			result.current.seek(42);
			result.current.setVolume(1.5);
		});
		expect(fakeAudio.currentTime).toBe(42);
		expect(result.current.volume).toBe(1);

		act(() => {
			result.current.toggleShuffle();
			result.current.setRepeatMode('all');
		});
		expect(result.current.shuffle).toBe(true);
		expect(result.current.repeatMode).toBe('all');

		await act(async () => {
			jest.advanceTimersByTime(2200);
		});
		expect(mockUpdatePlayerState).toHaveBeenCalledWith(
			expect.objectContaining({
				playlist_id: null,
			}),
		);

		Math.random = jest.fn(() => 0);
		act(() => {
			result.current.next();
		});
		expect(fakeAudio.src).toContain('/files/stream/1');

		fakeAudio.currentTime = 5;
		act(() => {
			result.current.previous();
		});
		expect(fakeAudio.currentTime).toBe(0);

		act(() => {
			result.current.setRepeatMode('one');
			fakeAudio.emit('ended');
		});
		expect(fakeAudio.play).toHaveBeenCalled();

		act(() => {
			result.current.clearQueue();
		});
		expect(result.current.queue).toEqual([]);
		expect(result.current.hasQueue).toBe(false);
		expect(result.current.playbackContext).toBeUndefined();
	});

	it('covers next/previous and track-end branches for repeat modes', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.addToQueue(track1);
			result.current.addToQueue(track2);
		});
		act(() => {
			jest.advanceTimersByTime(10);
		});

		act(() => {
			result.current.setRepeatMode('none');
			result.current.toggleShuffle();
			result.current.toggleShuffle(); // back to non-shuffle
			result.current.next();
		});

		fakeAudio.currentTime = 1;
		act(() => {
			result.current.previous();
		});

		act(() => {
			result.current.setRepeatMode('all');
			fakeAudio.emit('ended');
		});
		expect(fakeAudio.play).toHaveBeenCalled();

		act(() => {
			result.current.setRepeatMode('none');
			fakeAudio.emit('ended');
		});
		expect(result.current.repeatMode).toBe('none');
	});

	it('covers guard paths, default sync payload and fallback helpers', async () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.togglePlayPause();
			result.current.next();
			result.current.previous();
			result.current.playTrackFromQueue(-1);
			result.current.playTrackFromQueue(99);
			result.current.seek(15);
			result.current.setVolume(-0.5);
		});
		expect(fakeAudio.play).toHaveBeenCalled();
		expect(result.current.volume).toBe(0);

		act(() => {
			fakeAudio.emit('ended');
		});

		await act(async () => {
			jest.advanceTimersByTime(2200);
		});
		expect(mockUpdatePlayerState).toHaveBeenCalledWith(
			expect.objectContaining({
				current_file_id: null,
				current_position: expect.any(Number),
				volume: expect.any(Number),
			}),
		);

		expect(
			musicMetadata({
				format: '',
				size: 512,
			} as any),
		).toContain('512 B');
		expect(getMusicTitle({ id: 7, name: 'fallback.mp3', format: 'mp3', size: 1 } as any)).toBe('fallback.mp3');
		expect(getMusicArtist({ id: 7, name: 'x', format: 'mp3', size: 1 } as any)).toBe('Unknown Artist');
	});

	it('throws when hook is used outside provider', () => {
		expect(() => renderHook(() => useGlobalMusic())).toThrow('useGlobalMusic must be used within a GlobalMusicProvider');
	});

	it('syncs playlist identifiers when a playlist context replaces the queue', async () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		act(() => {
			result.current.replaceQueue([track1], 0, {
				kind: 'playlist',
				labelKey: 'MUSIC_PLAYBACK_CONTEXT_PLAYLIST',
				labelParams: { name: 'Mix A' },
				href: '/music/playlists',
				playlistId: 15,
			});
		});

		await act(async () => {
			jest.advanceTimersByTime(2200);
		});

		expect(result.current.playbackContext).toMatchObject({ playlistId: 15 });
		expect(mockUpdatePlayerState).toHaveBeenCalledWith(
			expect.objectContaining({
				playlist_id: 15,
				current_file_id: 1,
			}),
		);
	});

	it('covers queue toggling, end-of-queue stop, and remove-from-queue branches', () => {
		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		expect(result.current.queueOpen).toBe(false);
		act(() => {
			result.current.toggleQueue();
		});
		expect(result.current.queueOpen).toBe(true);

		act(() => {
			result.current.replaceQueue([track1, track2], 1, {
				kind: 'album',
				labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
				labelParams: { name: 'Album 1' },
				href: '/music/albums',
			});
		});
		act(() => {
			fakeAudio.emit('play');
			result.current.setRepeatMode('none');
			fakeAudio.emit('ended');
		});
		expect(result.current.isPlaying).toBe(false);

		act(() => {
			result.current.replaceQueue([track1, track2], 1);
			result.current.removeFromQueue(0);
		});
		expect(result.current.currentIndex).toBe(0);

		act(() => {
			result.current.replaceQueue([track1, track2], 0);
			result.current.removeFromQueue(0);
		});
		expect(fakeAudio.src).toContain('/files/stream/2');

		act(() => {
			result.current.replaceQueue([], 0);
		});
		expect(result.current.queue.length).toBeGreaterThan(0);
	});

	it('hydrates the queue from the persisted playlist when restoration is enabled', async () => {
		mockUseSettings.mockReturnValue({
			settings: {
				players: {
					remember_music_queue: true,
				},
			},
			isLoading: false,
		});
		mockGetPlayerState.mockResolvedValue({
			playlist_id: 99,
			current_file_id: 2,
			current_position: 12,
			volume: 0.7,
			shuffle: false,
			repeat_mode: 'none',
		});
		mockGetNowPlayingPlaylist.mockResolvedValue({
			id: 99,
			name: 'Recovered Mix',
			track_count: 2,
		});
		mockGetPlaylistTracks.mockResolvedValue({
			items: [{ file: track1 }, { file: track2 }],
		});

		const { result } = renderHook(() => useGlobalMusic(), { wrapper });

		await act(async () => {
			await Promise.resolve();
		});

		expect(mockGetPlayerState).toHaveBeenCalledTimes(1);
		expect(mockGetNowPlayingPlaylist).toHaveBeenCalledTimes(1);
		expect(mockGetPlaylistTracks).toHaveBeenCalledWith(99, 1, 2);
		expect(result.current.queue).toEqual([track1, track2]);
		expect(result.current.currentIndex).toBe(1);
		expect(result.current.playbackContext).toMatchObject({
			kind: 'playlist',
			playlistId: 99,
		});
	});
});

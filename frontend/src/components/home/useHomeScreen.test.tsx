import { renderHook } from '@testing-library/react';
import { useQuery } from '@tanstack/react-query';
import useHomeScreen, { homeScreenUtils } from './useHomeScreen';

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
}));

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
	useGlobalMusic: jest.fn(),
}));

jest.mock('@/service/analytics', () => ({
	fetchAnalyticsOverview: jest.fn(() => Promise.resolve({})),
}));

jest.mock('@/service/playerState', () => ({
	getPlayerState: jest.fn(() => Promise.resolve({})),
}));

jest.mock('@/service/playlist', () => ({
	getNowPlayingPlaylist: jest.fn(() => Promise.resolve({ id: 7 })),
	getPlaylistTracks: jest.fn(() => Promise.resolve({ items: [] })),
}));

jest.mock('@/service/videoPlayback', () => ({
	getVideoHomeCatalog: jest.fn(() => Promise.resolve({ sections: [] })),
	getVideoPlaybackState: jest.fn(() => Promise.resolve(null)),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseGlobalMusic = jest.requireMock('@/components/providers/GlobalMusicProvider').useGlobalMusic as jest.Mock;

const buildQueryState = (data: unknown, overrides?: Record<string, unknown>) => ({
	data,
	isLoading: false,
	...overrides,
});

describe('components/home/useHomeScreen', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseGlobalMusic.mockReturnValue({
			queue: [],
			currentTrack: undefined,
			currentTime: 0,
			duration: 0,
			isPlaying: false,
		});
	});

	it('builds the expected home queries', async () => {
		mockedUseQuery
			.mockReturnValueOnce(buildQueryState({ recent_files: [], health: {}, storage: {}, counts: {} }))
			.mockReturnValueOnce(buildQueryState({ items: [] }))
			.mockReturnValueOnce(buildQueryState({ items: [] }))
			.mockReturnValueOnce(buildQueryState({ sections: [] }))
			.mockReturnValueOnce(buildQueryState(null))
			.mockReturnValueOnce(buildQueryState({ current_file_id: null }))
			.mockReturnValueOnce(buildQueryState({ id: 9 }))
			.mockReturnValueOnce(buildQueryState({ items: [] }));

		renderHook(() => useHomeScreen());

		const analyticsOptions = mockedUseQuery.mock.calls[0][0];
		const favoritesOptions = mockedUseQuery.mock.calls[1][0];
		const imagesOptions = mockedUseQuery.mock.calls[2][0];
		const videoCatalogOptions = mockedUseQuery.mock.calls[3][0];
		const videoPlaybackOptions = mockedUseQuery.mock.calls[4][0];
		const playerStateOptions = mockedUseQuery.mock.calls[5][0];
		const nowPlayingOptions = mockedUseQuery.mock.calls[6][0];
		const nowPlayingTracksOptions = mockedUseQuery.mock.calls[7][0];

		expect(analyticsOptions.queryKey).toEqual(['home', 'analytics-overview', '30d']);
		await expect(analyticsOptions.queryFn()).resolves.toEqual({});

		expect(favoritesOptions.queryKey).toEqual(['home', 'favorites']);
		expect(imagesOptions.queryKey).toEqual(['home', 'images']);

		expect(videoCatalogOptions.queryKey).toEqual(['home', 'video-home-catalog']);
		await expect(videoCatalogOptions.queryFn()).resolves.toEqual({ sections: [] });

		expect(videoPlaybackOptions.queryKey).toEqual(['home', 'video-playback-state']);
		expect(videoPlaybackOptions.retry).toBe(false);

		expect(playerStateOptions.queryKey).toEqual(['home', 'music-player-state']);
		expect(playerStateOptions.retry).toBe(false);

		expect(nowPlayingOptions.queryKey).toEqual(['home', 'music-now-playing']);
		expect(nowPlayingOptions.retry).toBe(false);

		expect(nowPlayingTracksOptions.queryKey).toEqual(['home', 'music-now-playing-tracks', 9]);
		expect(nowPlayingTracksOptions.enabled).toBe(true);
		await expect(nowPlayingTracksOptions.queryFn()).resolves.toEqual({ items: [] });
	});

	it('derives music and video resume data from persisted state', () => {
		mockedUseQuery
			.mockReturnValueOnce(buildQueryState({
				recent_files: [{ id: 1, name: 'movie.mp4' }],
				health: { status: 'ok', indexed_files: 12, errors_last_24h: 0, last_scan_at: '2026-03-14T12:00:00Z' },
				storage: { used_bytes: 1024, total_bytes: 2048, free_bytes: 1024 },
				counts: {},
			}))
			.mockReturnValueOnce(buildQueryState({ items: [{ id: 55, name: 'favorite.mp4' }] }))
			.mockReturnValueOnce(buildQueryState({ items: [{ id: 77, name: 'cover.jpg' }] }))
			.mockReturnValueOnce(buildQueryState({
				sections: [
					{
						key: 'continue',
						title: 'continue',
						items: [
							{ video: { id: 3, name: 'Episode 3', parent_path: '/shows' }, progress_pct: 45, status: 'in_progress' },
						],
					},
				],
			}))
			.mockReturnValueOnce(buildQueryState({
				playlist: {
					id: 12,
					items: [{ video: { id: 3, name: 'Episode 3', parent_path: '/shows' } }],
				},
				playback_state: {
					video_id: 3,
					current_time: 90,
					duration: 180,
					playlist_id: 12,
				},
			}))
			.mockReturnValueOnce(buildQueryState({ current_file_id: 99, current_position: 60 }))
			.mockReturnValueOnce(buildQueryState({ id: 7, track_count: 4 }))
			.mockReturnValueOnce(buildQueryState({
				items: [
					{
						file: {
							id: 99,
							name: 'Track.mp3',
							size: 2048,
							updated_at: '2026-03-14T12:00:00Z',
							metadata: { title: 'Track', artist: 'Artist', duration: 240 },
						},
					},
				],
			}));

		const { result } = renderHook(() => useHomeScreen());

		expect(result.current.recentFiles).toHaveLength(1);
		expect(result.current.favoriteItems).toHaveLength(1);
		expect(result.current.recentImages).toHaveLength(1);
		expect(result.current.videoContinueItems).toHaveLength(1);
		expect(result.current.videoResume).toEqual({
			video: { id: 3, name: 'Episode 3', parent_path: '/shows' },
			progressSeconds: 90,
			durationSeconds: 180,
			progressPercent: 50,
			playlistId: 12,
		});
		expect(result.current.musicResume).toEqual({
			track: {
				id: 99,
				name: 'Track.mp3',
				size: 2048,
				updated_at: '2026-03-14T12:00:00Z',
				metadata: { title: 'Track', artist: 'Artist', duration: 240 },
			},
			progressSeconds: 60,
			durationSeconds: 240,
			progressPercent: 25,
			queueCount: 4,
			isPlaying: false,
		});
	});

	it('clamps invalid and overflow progress values', () => {
		expect(homeScreenUtils.getProgressPercent(Number.NaN, 120)).toBe(0);
		expect(homeScreenUtils.getProgressPercent(300, 100)).toBe(100);
		expect(homeScreenUtils.getProgressPercent(0, 0, 135)).toBe(100);
		expect(homeScreenUtils.getProgressPercent(0, 0, -20)).toBe(0);
	});
});

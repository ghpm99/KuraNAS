import { act, renderHook, waitFor } from '@testing-library/react';
import { useMusicHomeScreen } from './useMusicHomeScreen';

const mockUseMusic = jest.fn();
const mockUseGlobalMusic = jest.fn();
const mockUseQuery = jest.fn();
const mockGetPlaylistTracks = jest.fn();
const mockGetMusicByArtist = jest.fn();
const mockGetMusicByAlbum = jest.fn();

jest.mock('@/components/providers/musicProvider/musicProvider', () => ({
	useMusic: () => mockUseMusic(),
}));

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
	useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@tanstack/react-query', () => ({
	useQuery: (...args: any[]) => mockUseQuery(...args),
}));

jest.mock('@/service/playlist', () => ({
	getPlaylists: jest.fn(() => Promise.resolve({ items: [] })),
	getPlaylistTracks: (...args: any[]) => mockGetPlaylistTracks(...args),
}));

jest.mock('@/service/music', () => ({
	getMusicByArtist: (...args: any[]) => mockGetMusicByArtist(...args),
	getMusicByAlbum: (...args: any[]) => mockGetMusicByAlbum(...args),
}));

describe('useMusicHomeScreen', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseMusic.mockReturnValue({
			status: 'success',
			music: [
				{ id: 1, created_at: '2026-03-10T10:00:00Z', metadata: { artist: 'Artist A', album: 'Album A', year: 2024 } },
				{ id: 2, created_at: '2026-03-11T10:00:00Z', metadata: { artist: 'Artist B', album: 'Album B', year: 2025 } },
			],
		});
		mockUseGlobalMusic.mockReturnValue({
			currentIndex: 0,
			currentTrack: { id: 1 },
			getMusicArtist: (track: any) => `artist-${track.id}`,
			getMusicTitle: (track: any) => `title-${track.id}`,
			hasQueue: true,
			playbackContext: { href: '/music/albums' },
			queue: [{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }],
			replaceQueue: jest.fn(),
			toggleQueue: jest.fn(),
		});
		mockUseQuery.mockReturnValue({
			data: {
				items: [{ id: 5, name: 'Mix', description: '', track_count: 3, is_system: false }],
			},
			isLoading: false,
		});
		mockGetPlaylistTracks.mockResolvedValue({
			items: [{ file: { id: 10 } }],
		});
		mockGetMusicByArtist.mockResolvedValue({
			items: [{ id: 20 }],
		});
		mockGetMusicByAlbum.mockResolvedValue({
			items: [{ id: 30 }],
		});
	});

	it('derives home metrics, next tracks, and featured cards', () => {
		const { result } = renderHook(() => useMusicHomeScreen());

		expect(result.current.totalTracks).toBe(2);
		expect(result.current.totalArtists).toBe(2);
		expect(result.current.totalAlbums).toBe(2);
		expect(result.current.totalPlaylists).toBe(1);
		expect(result.current.currentTrackTitle).toBe('title-1');
		expect(result.current.currentTrackArtist).toBe('artist-1');
		expect(result.current.nextTracks).toEqual([
			{ id: 2, title: 'title-2', artist: 'artist-2' },
			{ id: 3, title: 'title-3', artist: 'artist-3' },
			{ id: 4, title: 'title-4', artist: 'artist-4' },
		]);
		expect(result.current.returnToContextHref).toBe('/music/albums');
		expect(result.current.featuredPlaylists[0]).toMatchObject({ href: '/music/playlists', actionKey: 'playlist-5' });
		expect(result.current.artistHighlights[0]).toMatchObject({ href: '/music/artists' });
		expect(result.current.albumHighlights[0]).toMatchObject({ href: '/music/albums' });
	});

	it('handles empty playback fetches and pending action state', async () => {
		const replaceQueue = jest.fn();
		let resolvePlaylistTracks: ((value: { items: never[] }) => void) | undefined;
		mockUseGlobalMusic.mockReturnValue({
			currentIndex: undefined,
			currentTrack: undefined,
			getMusicArtist: () => '',
			getMusicTitle: () => '',
			hasQueue: false,
			playbackContext: undefined,
			queue: [],
			replaceQueue,
			toggleQueue: jest.fn(),
		});
		mockUseQuery.mockReturnValue({
			data: undefined,
			isLoading: true,
		});
		mockGetPlaylistTracks.mockImplementationOnce(
			() =>
				new Promise((resolve) => {
					resolvePlaylistTracks = resolve;
				}),
		);
		mockGetMusicByArtist.mockResolvedValueOnce({ items: [] });
		mockGetMusicByAlbum.mockResolvedValueOnce({ items: [] });

		const { result } = renderHook(() => useMusicHomeScreen());

		expect(result.current.isLoadingPlaylists).toBe(true);
		expect(result.current.featuredPlaylists).toEqual([]);
		expect(result.current.nextTracks).toEqual([]);

		let playlistPromise: Promise<void> | undefined;
		await act(async () => {
			playlistPromise = result.current.playPlaylist(5, 'Mix');
		});
		await waitFor(() => {
			expect(result.current.isActionPending('playlist-5')).toBe(true);
		});
		await act(async () => {
			resolvePlaylistTracks?.({ items: [] });
			await playlistPromise;
		});
		expect(replaceQueue).not.toHaveBeenCalled();
		expect(result.current.isActionPending('playlist-5')).toBe(false);

		await act(async () => {
			await result.current.playArtist('Artist A');
			await result.current.playAlbum('Album A');
		});
		expect(replaceQueue).not.toHaveBeenCalled();
	});
});

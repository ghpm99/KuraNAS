import { act, renderHook } from '@testing-library/react';
import useGlobalSearchProvider from './useGlobalSearchProvider';

const mockNavigate = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

jest.mock('react-router-dom', () => ({
	useNavigate: () => mockNavigate,
	useLocation: () => ({ pathname: '/files', search: '' }),
}));

jest.mock('@tanstack/react-query', () => ({
	useQuery: (opts: any) => ({
		data: opts.enabled
			? {
					files: [
						{ id: 1, name: 'FILE', path: '/path/file', format: 'mp4', type: 1, parent_path: '/', size: 1 },
					],
					folders: [{ id: 2, name: 'Folder', path: '/folder' }],
					artists: [{ key: 'artist-1', artist: 'Artist', track_count: 2, album_count: 1 }],
					albums: [{ key: 'album-1', album: 'Album', artist: 'Artist', track_count: 5 }],
					playlists: [
						{ id: 3, name: 'Playlist', scope: 'music', count: 5, source_path: '', classification: 'personal', description: '' },
					],
					videos: [{ id: 4, name: 'Video', path: '/video.mp4', format: 'video/mp4' }],
					images: [{ id: 5, name: 'Image', path: '/image.jpg', context: 'Library', category: 'folder' }],
			  }
			: undefined,
		isFetching: false,
	}),
}));

jest.mock('@/components/videos/navigation', () => ({
	getVideoDetailRoute: () => '/video/playlist',
	getVideoSectionForPlaylist: () => 'home',
}));

jest.mock('@/app/routes', () => ({
	appRoutes: {
		home: '/home',
		files: '/files',
		favorites: '/favorites',
		legacyFavorites: '/starred',
		settings: '/settings',
		about: '/about',
		images: '/images',
		music: '/music',
		videos: '/videos',
		analytics: '/analytics',
		videoPlayerBase: '/video',
	},
	getMusicRoute: (section: string) => `/music/${section}`,
	getVideoRoute: (section: string) => `/videos/${section}`,
	getAnalyticsRoute: (section: string) => `/analytics/${section}`,
}));

describe('useGlobalSearchProvider', () => {
	let platformSpy: jest.SpyInstance;

	beforeEach(() => {
		platformSpy = jest.spyOn(window.navigator, 'platform', 'get').mockReturnValue('Win32');
		mockNavigate.mockReset();
	});

	afterEach(() => {
		platformSpy.mockRestore();
	});

	it('builds search sections with data and handles keyboard navigation', async () => {
		const { result } = renderHook(() => useGlobalSearchProvider());

		await act(async () => {
			result.current.openSearch();
			result.current.setQuery('search');
			await Promise.resolve();
		});

		expect(result.current.sections.some((section) => section.id === 'files')).toBe(true);
		expect(result.current.sections.some((section) => section.id === 'artists')).toBe(true);
		expect(result.current.showEmptyState).toBe(false);
		const event = { key: 'ArrowDown', preventDefault: jest.fn() } as any;
		act(() => {
			result.current.handleInputKeyDown(event);
		});
		expect(event.preventDefault).toHaveBeenCalled();
	});

	it('activates items and closes search', async () => {
		const { result } = renderHook(() => useGlobalSearchProvider());

		await act(async () => {
			result.current.openSearch();
			result.current.setQuery('search');
			await Promise.resolve();
		});

		const firstItem = result.current.sections[0].items[0];
		act(() => {
			result.current.activateItem(firstItem);
		});

		expect(mockNavigate).toHaveBeenCalled();
		expect(result.current.open).toBe(false);
		expect(result.current.query).toBe('');
	});
});

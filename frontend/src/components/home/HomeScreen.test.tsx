import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import HomeScreen from './HomeScreen';

const mockUseHomeScreen = jest.fn();
const mockOpenMediaItem = jest.fn();
const mockNavigate = jest.fn();
const mockOpenSearch = jest.fn();

jest.mock('./useHomeScreen', () => ({
	__esModule: true,
	default: () => mockUseHomeScreen(),
}));
jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
	__esModule: true,
	default: () => ({
		openMediaItem: (...args: any[]) => mockOpenMediaItem(...args),
	}),
}));
jest.mock('@/components/search/useGlobalSearch', () => ({
	__esModule: true,
	default: () => ({
		openSearch: mockOpenSearch,
		shortcut: 'Ctrl+K',
	}),
}));
jest.mock('react-router-dom', () => {
	const actual = jest.requireActual('react-router-dom');
	return {
		...actual,
		useNavigate: () => mockNavigate,
	};
});

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string | number>) => {
			if (key === 'HOME_QUEUE_COUNT') {
				return `${params?.count} in queue`;
			}

			const translations: Record<string, string> = {
				HOME_PAGE_TITLE: 'Welcome to KuraNAS',
				HOME_PAGE_DESCRIPTION: 'Home description',
				HOME_HERO_EYEBROW: 'Home hub',
				SEARCH_PLACEHOLDER: 'Search...',
					HOME_LIBRARY_TITLE: 'Explore library',
				HOME_LIBRARY_DESCRIPTION: 'Library routes',
				HOME_MEDIA_DESCRIPTION: 'Media routes',
				HOME_SYSTEM_DESCRIPTION: 'System routes',
				FILES: 'Files',
				STARRED_FILES: 'Favorites',
				NAV_IMAGES: 'Images',
				NAV_MUSIC: 'Music',
				NAV_VIDEOS: 'Videos',
				ANALYTICS: 'Analytics',
				SETTINGS: 'Settings',
				HOME_OPEN_SECTION: 'Open area',
				RECENT_FILES: 'Recent files',
				HOME_RECENT_DESCRIPTION: 'Recent files description',
				MUSIC_NOW_PLAYING: 'Now playing',
				HOME_MUSIC_DESCRIPTION: 'Music description',
				HOME_RESUME_ACTION: 'Resume',
				HOME_UNKNOWN_ARTIST: 'Unknown artist',
				VIDEO_CONTINUE_WATCHING: 'Continue watching',
				HOME_VIDEO_DESCRIPTION: 'Video description',
				STATUS_SYSTEM_TITLE: 'System status',
				HOME_STATUS_DESCRIPTION: 'System status description',
				HOME_STORAGE_LABEL: 'Storage used',
				HOME_INDEX_LABEL: 'Indexed files',
				HOME_LAST_SCAN_LABEL: 'Last scan',
				HOME_ERRORS_LABEL: 'Errors 24h',
				ANALYTICS_STATUS_OK: 'OK',
				HOME_RECENT_EMPTY: 'No recent files',
				HOME_MUSIC_EMPTY: 'No music',
				HOME_VIDEO_EMPTY: 'No videos',
					GLOBAL_SEARCH_SHORTCUT: `Open global search with ${params?.shortcut}`,
					LOADING: 'Loading...',
				IMAGE_JPEG: 'JPEG Image',
				IN_PROGRESS: 'In progress',
				VIDEO_CONTINUE_BADGE_RESUME: 'Resume',
				HOME_LAST_SCAN_EMPTY: 'No recent scan',
			};

			return translations[key] || key;
		},
	}),
}));

jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => 'http://localhost/api/v1',
}));

describe('components/home/HomeScreen', () => {
		beforeEach(() => {
			jest.clearAllMocks();
			mockOpenMediaItem.mockReturnValue(false);
		mockUseHomeScreen.mockReturnValue({
			searchQuery: '',
			setSearchQuery: jest.fn(),
			recentFiles: [
					{
						id: 21,
						name: 'photo.jpg',
						path: '/photos/photo.jpg',
						parent_path: '/photos',
						format: '.jpg',
					size_bytes: 1024,
					created_at: '2026-03-14T12:00:00Z',
				},
			],
			videoContinueItems: [
				{
					video: { id: 5, name: 'Episode 5', parent_path: '/shows' },
					progress_pct: 38,
					status: 'in_progress',
				},
			],
			videoResume: {
				video: { id: 5, name: 'Episode 5', parent_path: '/shows' },
				progressSeconds: 100,
				durationSeconds: 200,
				progressPercent: 50,
				playlistId: 14,
			},
			musicResume: {
				track: {
					id: 11,
					name: 'track.mp3',
					size: 2048,
					updated_at: '2026-03-14T12:00:00Z',
					metadata: { title: 'Track Title', artist: 'Artist Name', duration: 220 },
				},
				progressSeconds: 40,
				durationSeconds: 220,
				progressPercent: 18,
				queueCount: 3,
				isPlaying: true,
			},
			analytics: {
				storage: { used_bytes: 1024, total_bytes: 2048, free_bytes: 1024 },
				health: { status: 'ok', indexed_files: 10, errors_last_24h: 0, last_scan_at: '2026-03-14T12:00:00Z' },
			},
			isAnalyticsLoading: false,
			isVideoLoading: false,
			isMusicLoading: false,
		});
	});

	it('renders the home hub with resume and status sections', () => {
		render(
			<MemoryRouter>
				<HomeScreen />
			</MemoryRouter>,
		);

		expect(screen.getByText('Welcome to KuraNAS')).toBeInTheDocument();
		expect(screen.getByText('Track Title')).toBeInTheDocument();
		expect(screen.getByText('Artist Name')).toBeInTheDocument();
		expect(screen.getByText('Episode 5')).toBeInTheDocument();
		expect(screen.getByText('photo.jpg')).toBeInTheDocument();
		expect(screen.getByText('System status')).toBeInTheDocument();
		expect(screen.getAllByText('Open area').length).toBeGreaterThan(0);
		expect(screen.getAllByText('Resume').length).toBeGreaterThan(0);
	});

	it('uses the shared media opener for recent files and falls back to files when needed', () => {
		render(
			<MemoryRouter>
				<HomeScreen />
			</MemoryRouter>,
		);

			fireEvent.click(screen.getByRole('button', { name: /photo\.jpg/i }));

			expect(mockOpenMediaItem).toHaveBeenCalledWith(expect.objectContaining({ id: 21, name: 'photo.jpg' }));
			expect(mockNavigate).toHaveBeenCalledWith({
				pathname: '/files',
				search: '?path=%2Fphotos%2Fphoto.jpg',
			});
		});

		it('opens the global search from the home hero and renders empty states when no content is available', () => {
			mockUseHomeScreen.mockReturnValue({
				recentFiles: [],
				videoContinueItems: [],
				videoResume: null,
			musicResume: null,
			analytics: null,
			isAnalyticsLoading: false,
			isVideoLoading: false,
			isMusicLoading: false,
		});

			render(
				<MemoryRouter>
					<HomeScreen />
				</MemoryRouter>,
			);

			fireEvent.click(screen.getByLabelText('GLOBAL_SEARCH_OPEN'));

			expect(mockOpenSearch).toHaveBeenCalled();
			expect(screen.getByText('No recent files')).toBeInTheDocument();
			expect(screen.getByText('No music')).toBeInTheDocument();
			expect(screen.getByText('No videos')).toBeInTheDocument();
	});
});

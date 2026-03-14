import { render, screen } from '@testing-library/react';
import App from './App';

const mockUseLocation = jest.fn();

jest.mock('react-router-dom', () => ({
	Routes: ({ children }: any) => <div data-testid='routes'>{children}</div>,
	Route: ({ element }: any) => <div>{element}</div>,
	Navigate: () => <div>Navigate</div>,
	useLocation: () => mockUseLocation(),
}));

jest.mock('@/components/providers/appProviders', () => ({ children }: any) => <div data-testid='app-providers'>{children}</div>);
jest.mock('@/components/providers/GlobalMusicProvider', () => ({ GlobalMusicProvider: ({ children }: any) => <div data-testid='music-providers'>{children}</div> }));
jest.mock('@/components/player/GlobalPlayerControl', () => () => <div>GlobalPlayerControl</div>);
jest.mock('@/components/ErrorBoundary', () => ({ children }: any) => <div data-testid='error-boundary'>{children}</div>);

jest.mock('@/pages/activityDiary', () => () => <div>ActivityDiaryPage</div>);
jest.mock('@/pages/analytics', () => () => <div>AnalyticsPage</div>);
jest.mock('@/pages/favorites', () => () => <div>FavoritesPage</div>);
jest.mock('@/pages/files', () => () => <div>FilePage</div>);
jest.mock('@/pages/home', () => () => <div>HomePage</div>);
jest.mock('@/pages/about', () => () => <div>AboutPage</div>);
jest.mock('@/pages/images', () => () => <div>ImagesPage</div>);
jest.mock('@/pages/music', () => () => <div>MusicPage</div>);
jest.mock('@/components/music/MusicHomeScreen', () => () => <div>MusicHomeScreen</div>);
jest.mock('@/pages/music/views/AlbumsView', () => () => <div>AlbumsView</div>);
jest.mock('@/pages/music/views/ArtistsView', () => () => <div>ArtistsView</div>);
jest.mock('@/pages/music/views/FoldersView', () => () => <div>FoldersView</div>);
jest.mock('@/pages/music/views/GenresView', () => () => <div>GenresView</div>);
jest.mock('@/pages/music/views/PlaylistsView', () => () => <div>PlaylistsView</div>);
jest.mock('@/pages/settings', () => () => <div>SettingsPage</div>);
jest.mock('@/pages/videos/videos', () => () => <div>VideosPage</div>);
jest.mock('@/pages/videoPlayer/videoPlayer', () => () => <div>VideoPlayerPage</div>);

describe('App', () => {
	it('shows global player when route is not video', () => {
		mockUseLocation.mockReturnValue({ pathname: '/music' });
		render(<App />);

		expect(screen.getByTestId('app-providers')).toBeInTheDocument();
		expect(screen.getByTestId('music-providers')).toBeInTheDocument();
		expect(screen.getByText('GlobalPlayerControl')).toBeInTheDocument();
	});

	it('hides global player on video route', () => {
		mockUseLocation.mockReturnValue({ pathname: '/video/22' });
		render(<App />);

		expect(screen.queryByText('GlobalPlayerControl')).not.toBeInTheDocument();
		expect(screen.getByTestId('routes')).toBeInTheDocument();
	});
});

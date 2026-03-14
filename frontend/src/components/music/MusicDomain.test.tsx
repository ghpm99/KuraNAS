import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import MusicDomainHeader from './MusicDomainHeader';
import MusicHomeScreen from './MusicHomeScreen';
import MusicSidebar from './MusicSidebar';

const mockUseMusic = jest.fn();
const mockUseGlobalMusic = jest.fn();

jest.mock('@/components/providers/musicProvider/musicProvider', () => ({
	useMusic: () => mockUseMusic(),
}));

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
	useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string, params?: Record<string, string>) => params?.count ? `${key}:${params.count}` : key }),
}));

describe('components/music domain shell', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseMusic.mockReturnValue({
			status: 'success',
			music: [{ id: 1 }, { id: 2 }, { id: 3 }],
		});
		mockUseGlobalMusic.mockReturnValue({
			currentTrack: { id: 10, name: 'Song A' },
			getMusicArtist: () => 'Artist A',
			getMusicTitle: () => 'Song A',
			hasQueue: true,
			queue: [{ id: 10 }, { id: 11 }],
		});
	});

	it('renders contextual header and active sidebar item from route', () => {
		render(
			<MemoryRouter initialEntries={['/music/albums']}>
				<MusicDomainHeader />
				<MusicSidebar />
			</MemoryRouter>,
		);

		expect(screen.getByRole('heading', { name: 'MUSIC_ALBUMS' })).toBeInTheDocument();
		expect(screen.getAllByText('MUSIC_ALBUMS_DESCRIPTION')[0]).toBeInTheDocument();
		expect(screen.getByRole('link', { name: /MUSIC_ALBUMS/i })).toHaveAttribute('href', '/music/albums');
	});

	it('renders music home with queue and section shortcuts', () => {
		render(
			<MemoryRouter initialEntries={['/music']}>
				<MusicHomeScreen />
			</MemoryRouter>,
		);

		expect(screen.getByText('MUSIC_HOME_QUEUE_READY')).toBeInTheDocument();
		expect(screen.getByText('Song A')).toBeInTheDocument();
		expect(screen.getByText('MUSIC_HOME_QUEUE_COUNT:2')).toBeInTheDocument();
		expect(screen.getByText('3')).toBeInTheDocument();
		expect(screen.getAllByRole('link', { name: 'MUSIC_HOME_OPEN_SECTION' })[0]).toHaveAttribute('href', '/music/playlists');
	});

	it('renders music home empty state while catalog is loading', () => {
		mockUseMusic.mockReturnValue({
			status: 'pending',
			music: [],
		});
		mockUseGlobalMusic.mockReturnValue({
			currentTrack: undefined,
			getMusicArtist: () => '',
			getMusicTitle: () => '',
			hasQueue: false,
			queue: [],
		});

		render(
			<MemoryRouter initialEntries={['/music']}>
				<MusicHomeScreen />
			</MemoryRouter>,
		);

		expect(screen.getByText('MUSIC_HOME_QUEUE_EMPTY')).toBeInTheDocument();
		expect(screen.getByText('MUSIC_HOME_QUEUE_EMPTY_STATE')).toBeInTheDocument();
		expect(screen.getByText('MUSIC_HOME_LIBRARY_LOADING')).toBeInTheDocument();
	});
});

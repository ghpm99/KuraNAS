import { render, screen } from '@testing-library/react';
import React from 'react';
import MusicContent from './musicContent';

const mockUseMusic = jest.fn();
const mockUseGlobalMusic = jest.fn();

jest.mock('../hooks/musicProvider/musicProvider', () => ({ useMusic: () => mockUseMusic() }));
jest.mock('../providers/GlobalMusicProvider', () => ({ useGlobalMusic: () => mockUseGlobalMusic() }));

jest.mock('@/pages/music/views/AllTracksView', () => () => <div>AllTracksView</div>);
jest.mock('@/pages/music/views/ArtistsView', () => () => <div>ArtistsView</div>);
jest.mock('@/pages/music/views/AlbumsView', () => () => <div>AlbumsView</div>);
jest.mock('@/pages/music/views/GenresView', () => () => <div>GenresView</div>);
jest.mock('@/pages/music/views/FoldersView', () => () => <div>FoldersView</div>);
jest.mock('@/pages/music/views/PlaylistsView', () => () => <div>PlaylistsView</div>);
jest.mock('../playlist/paylist', () => () => <div>Playlist</div>);

describe('musicContent', () => {
	it('renders all views and queue panel branch', () => {
		mockUseGlobalMusic.mockReturnValue({ hasQueue: true });
		const views = ['all', 'artists', 'albums', 'genres', 'folders', 'playlists'] as const;
		const labels = ['AllTracksView', 'ArtistsView', 'AlbumsView', 'GenresView', 'FoldersView', 'PlaylistsView'];

		views.forEach((view, i) => {
			mockUseMusic.mockReturnValue({ currentView: view });
			render(<MusicContent />);
			expect(screen.getByText(labels[i])).toBeInTheDocument();
		});
		expect(screen.getAllByText('Playlist').length).toBeGreaterThan(0);

		mockUseGlobalMusic.mockReturnValue({ hasQueue: false });
		mockUseMusic.mockReturnValue({ currentView: 'all' });
		render(<MusicContent />);
		expect(screen.getAllByText('AllTracksView').length).toBeGreaterThan(0);
	});
});

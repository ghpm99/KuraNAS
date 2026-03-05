import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import AllTracksView from './AllTracksView';

const mockUseMusic = jest.fn();
const mockUseGlobalMusic = jest.fn();
const mockAddToQueue = jest.fn();

jest.mock('@/components/providers/musicProvider/musicProvider', () => ({ useMusic: () => mockUseMusic() }));
jest.mock('@/components/providers/GlobalMusicProvider', () => ({ useGlobalMusic: () => mockUseGlobalMusic() }));

jest.mock('@/components/music/AddToPlaylistMenu', () => (props: any) => (
	<div>
		<span>AddToPlaylistMenu-{String(props.fileId)}</span>
		<span>MenuAnchor-{props.anchorEl ? 'open' : 'closed'}</span>
		<button type='button' onClick={props.onClose}>
			close-menu
		</button>
	</div>
));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => k }),
}));

const baseTrack = {
	id: 1,
	name: 'track-1',
	path: '/root/folder/track-1.mp3',
	artist: 'artist-1',
	format: 'mp3',
	parent_path: '/root/folder',
	size: 1024,
};

describe('AllTracksView', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseGlobalMusic.mockReturnValue({
			getMusicTitle: (m: any) => m.name,
			getMusicArtist: (m: any) => m.artist,
			musicMetadata: () => 'meta',
			addToQueue: mockAddToQueue,
		});
	});

	it('renders tracks, allows queue action and shows all-loaded message', () => {
		mockUseMusic.mockReturnValue({
			music: [baseTrack],
			hasNextPage: false,
			isFetchingNextPage: false,
			lastItemRef: jest.fn(),
		});

		render(<AllTracksView />);
		expect(screen.getByText('track-1')).toBeInTheDocument();
		expect(screen.getByText('MUSIC_ALL_LOADED')).toBeInTheDocument();
		expect(screen.getByText('AddToPlaylistMenu-0')).toBeInTheDocument();
		expect(screen.getByText('MenuAnchor-closed')).toBeInTheDocument();
		fireEvent.click(screen.getByText('track-1'));
		expect(mockAddToQueue).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }));

		fireEvent.click(screen.getByRole('button', { name: 'add track-1 to playlist' }));
		expect(screen.getByText('AddToPlaylistMenu-1')).toBeInTheDocument();
		expect(screen.getByText('MenuAnchor-open')).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: 'close-menu' }));
		expect(screen.getByText('AddToPlaylistMenu-0')).toBeInTheDocument();
	});

	it('renders fetching spinner and omits all-loaded when there are next pages', () => {
		mockUseMusic.mockReturnValue({
			music: [baseTrack, { ...baseTrack, id: 2, name: 'track-2' }],
			hasNextPage: true,
			isFetchingNextPage: true,
			lastItemRef: jest.fn(),
		});

		render(<AllTracksView />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('MUSIC_ALL_LOADED')).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'play track-2' })).toBeInTheDocument();
		expect(screen.getByText('MenuAnchor-closed')).toBeInTheDocument();
	});

	it('does not show all-loaded text when there are no tracks', () => {
		mockUseMusic.mockReturnValue({
			music: [],
			hasNextPage: false,
			isFetchingNextPage: false,
			lastItemRef: jest.fn(),
		});

		render(<AllTracksView />);
		expect(screen.queryByText('MUSIC_ALL_LOADED')).not.toBeInTheDocument();
		expect(screen.getByText('AddToPlaylistMenu-0')).toBeInTheDocument();
	});
});

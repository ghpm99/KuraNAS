import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import AllTracksView from './AllTracksView';
import ArtistsView from './ArtistsView';
import AlbumsView from './AlbumsView';
import GenresView from './GenresView';
import FoldersView from './FoldersView';

const mockUseMusic = jest.fn();
const mockUseGlobalMusic = jest.fn();
const mockUseInfiniteQuery = jest.fn();
const mockUseQuery = jest.fn();
const mockGetMusicArtists = jest.fn();
const mockGetMusicByArtist = jest.fn();
const mockGetMusicAlbums = jest.fn();
const mockGetMusicByAlbum = jest.fn();
const mockGetMusicGenres = jest.fn();
const mockGetMusicByGenre = jest.fn();
const mockGetMusicFolders = jest.fn();
const mockApiGet = jest.fn();
const mockAddToQueue = jest.fn();

jest.mock('@/components/providers/musicProvider/musicProvider', () => ({ useMusic: () => mockUseMusic() }));
jest.mock('@/components/providers/GlobalMusicProvider', () => ({ useGlobalMusic: () => mockUseGlobalMusic() }));
jest.mock('@tanstack/react-query', () => ({
	useInfiniteQuery: (...args: any[]) => mockUseInfiniteQuery(...args),
	useQuery: (...args: any[]) => mockUseQuery(...args),
}));

jest.mock('@/service/music', () => ({
	getMusicArtists: (...args: any[]) => mockGetMusicArtists(...args),
	getMusicByArtist: (...args: any[]) => mockGetMusicByArtist(...args),
	getMusicAlbums: (...args: any[]) => mockGetMusicAlbums(...args),
	getMusicByAlbum: (...args: any[]) => mockGetMusicByAlbum(...args),
	getMusicGenres: (...args: any[]) => mockGetMusicGenres(...args),
	getMusicByGenre: (...args: any[]) => mockGetMusicByGenre(...args),
	getMusicFolders: (...args: any[]) => mockGetMusicFolders(...args),
}));

jest.mock('@/service', () => ({
	apiBase: {
		get: (...args: any[]) => mockApiGet(...args),
	},
}));

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

const track = {
	id: 1,
	name: 'track-1',
	path: '/root/folder/track-1.mp3',
	artist: 'artist-1',
	album: 'album-1',
	genre: 'genre-1',
	format: 'mp3',
	parent_path: '/root/folder',
	size: 1024,
};

const makePagination = (items: any[], hasNext = false, pageNo = 1) => ({
	items,
	pagination: {
		page: pageNo,
		has_next: hasNext,
		has_prev: pageNo > 1,
		total_pages: hasNext ? pageNo + 1 : pageNo,
		total_items: items.length,
		page_size: 50,
	},
});

const makeInfiniteResult = (items: any[], hasNext = false, isFetchingNextPage = false) => ({
	data: { pages: [makePagination(items, hasNext)] },
	isLoading: false,
	fetchNextPage: jest.fn(),
	hasNextPage: hasNext,
	isFetchingNextPage,
});

beforeEach(() => {
	jest.clearAllMocks();
	mockUseGlobalMusic.mockReturnValue({
		getMusicTitle: (m: any) => m.name,
		getMusicArtist: (m: any) => m.artist ?? 'artist',
		musicMetadata: () => 'meta',
		addToQueue: mockAddToQueue,
	});
	mockUseMusic.mockReturnValue({
		music: [track],
		hasNextPage: false,
		isFetchingNextPage: false,
		lastItemRef: jest.fn(),
	});
	mockGetMusicArtists.mockResolvedValue(makePagination([{ artist: 'artist-1', album_count: 1, track_count: 1 }]));
	mockGetMusicByArtist.mockResolvedValue(makePagination([track]));
	mockGetMusicAlbums.mockResolvedValue(makePagination([{ album: 'album-1', artist: 'artist-1', track_count: 1, year: 2024 }]));
	mockGetMusicByAlbum.mockResolvedValue(makePagination([track]));
	mockGetMusicGenres.mockResolvedValue(makePagination([{ genre: 'genre-1', track_count: 1 }]));
	mockGetMusicByGenre.mockResolvedValue(makePagination([track]));
	mockGetMusicFolders.mockResolvedValue(makePagination([{ folder: '/root/folder', track_count: 1 }]));
	mockApiGet.mockResolvedValue({
		data: {
			items: [track, { ...track, id: 2, path: '/other/track-2.mp3', parent_path: '/other' }],
			pagination: { page: 1, has_next: false, has_prev: false, total_pages: 1, total_items: 2 },
		},
	});

	mockUseInfiniteQuery.mockImplementation((options: any) => {
		const [key] = options.queryKey as [string, ...any[]];
		options.getNextPageParam?.({ pagination: { has_next: true, page: 1 } });
		options.getNextPageParam?.({ pagination: { has_next: false, page: 1 } });
		options.queryFn?.({});
		options.queryFn?.({ pageParam: 1 });

		switch (key) {
			case 'music-artists':
				return makeInfiniteResult([{ artist: 'artist-1', album_count: 1, track_count: 1 }], true);
			case 'music-by-artist':
				return makeInfiniteResult([track], true);
			case 'music-albums':
				return makeInfiniteResult([
					{ album: 'album-1', artist: 'artist-1', track_count: 1, year: 2024 },
					{ album: 'album-2', artist: 'artist-2', track_count: 2, year: undefined },
				]);
			case 'music-by-album':
				return makeInfiniteResult([track], true);
			case 'music-genres':
				return makeInfiniteResult([{ genre: 'genre-1', track_count: 1 }]);
			case 'music-by-genre':
				return makeInfiniteResult([track], true);
			case 'music-folders':
				return makeInfiniteResult([{ folder: '/root/folder', track_count: 1 }, { folder: '/', track_count: 2 }], true);
			default:
				return makeInfiniteResult([]);
		}
	});

	mockUseQuery.mockImplementation((options: any) => {
		options.queryFn?.();
		return {
			data: {
				items: [track],
				pagination: { page: 1, has_next: false, has_prev: false, total_pages: 1, total_items: 1 },
			},
			isLoading: false,
		};
	});
});

describe('music views', () => {
	it('renders all tracks view and handles add/menu states', () => {
		render(<AllTracksView />);
		expect(screen.getByText('track-1')).toBeInTheDocument();
		expect(screen.getByText('AddToPlaylistMenu-0')).toBeInTheDocument();
		expect(screen.getByText('MUSIC_ALL_LOADED')).toBeInTheDocument();

		fireEvent.click(screen.getByText('track-1'));
		expect(mockAddToQueue).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }));

		fireEvent.click(screen.getByRole('button', { name: 'add track-1 to playlist' }));
		fireEvent.click(screen.getAllByRole('button', { name: 'close-menu' })[0]);
	});

	it('renders artists list/detail flow, load-more and back', () => {
		const fetchArtistTracks = jest.fn();
		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			options.getNextPageParam?.({ pagination: { has_next: true, page: 1 } });
			options.getNextPageParam?.({ pagination: { has_next: false, page: 1 } });
			if (key === 'music-artists') {
				return makeInfiniteResult([{ artist: 'artist-1', album_count: 1, track_count: 1 }], true, true);
			}
			if (key === 'music-by-artist') {
				return { ...makeInfiniteResult([track], true), fetchNextPage: fetchArtistTracks };
			}
			return makeInfiniteResult([]);
		});

		render(<ArtistsView />);
		fireEvent.click(screen.getByText('artist-1'));
		expect(screen.getByText('track-1')).toBeInTheDocument();
		fireEvent.click(screen.getByText('Load more'));
		expect(fetchArtistTracks).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[0]);
		expect(screen.getByText('artist-1')).toBeInTheDocument();
	});

	it('renders albums flow and exercises track actions', () => {
		render(<AlbumsView />);
		fireEvent.click(screen.getByText('album-1'));
		expect(screen.getByText('track-1')).toBeInTheDocument();
		expect(screen.getByText('AddToPlaylistMenu-0')).toBeInTheDocument();

		fireEvent.click(screen.getByText('track-1'));
		expect(mockAddToQueue).toHaveBeenCalled();

		fireEvent.click(screen.getAllByRole('button')[0]);
		expect(screen.getByText('album-2')).toBeInTheDocument();
	});

	it('renders genres flow, detail load-more and menu click', () => {
		const fetchGenreTracks = jest.fn();
		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			options.getNextPageParam?.({ pagination: { has_next: true, page: 1 } });
			options.getNextPageParam?.({ pagination: { has_next: false, page: 1 } });
			if (key === 'music-genres') return makeInfiniteResult([{ genre: 'genre-1', track_count: 1 }]);
			if (key === 'music-by-genre') return { ...makeInfiniteResult([track], true), fetchNextPage: fetchGenreTracks };
			return makeInfiniteResult([]);
		});

		render(<GenresView />);
		fireEvent.click(screen.getByText('genre-1'));
		expect(screen.getByText('track-1')).toBeInTheDocument();
		expect(screen.getByText('AddToPlaylistMenu-0')).toBeInTheDocument();
		expect(screen.getByText('MenuAnchor-closed')).toBeInTheDocument();
		fireEvent.click(screen.getByText('Load more'));
		expect(fetchGenreTracks).toHaveBeenCalled();

		fireEvent.click(screen.getByRole('button', { name: 'add track-1 to playlist' }));
		expect(screen.getByText('AddToPlaylistMenu-1')).toBeInTheDocument();
		expect(screen.getByText('MenuAnchor-open')).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: 'close-menu' }));
		expect(screen.getByText('MenuAnchor-closed')).toBeInTheDocument();
	});

	it('renders folders flow and folder filter/query path', () => {
		render(<FoldersView />);
		fireEvent.click(screen.getByText('folder'));
		expect(screen.getByText('track-1')).toBeInTheDocument();
		expect(screen.getByText('AddToPlaylistMenu-0')).toBeInTheDocument();

		fireEvent.click(screen.getByText('track-1'));
		expect(mockAddToQueue).toHaveBeenCalled();

		fireEvent.click(screen.getAllByRole('button')[0]);
		expect(screen.getByText('/ - 2 tracks')).toBeInTheDocument();
	});

	it('renders loading states of artists and albums', () => {
		mockUseInfiniteQuery.mockReturnValueOnce({
			data: undefined,
			isLoading: true,
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});
		const { unmount } = render(<ArtistsView />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		unmount();

		mockUseInfiniteQuery.mockReturnValueOnce({
			data: undefined,
			isLoading: true,
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});
		render(<AlbumsView />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
	});

	it('renders genres/folders loading and fetching branches', () => {
		mockUseInfiniteQuery.mockReturnValueOnce({
			data: undefined,
			isLoading: true,
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});
		const { unmount } = render(<GenresView />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		unmount();

		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			if (key === 'music-genres') {
				return {
					data: { pages: [makePagination([{ genre: 'genre-1', track_count: 1 }], true)] },
					isLoading: false,
					fetchNextPage: jest.fn(),
					hasNextPage: true,
					isFetchingNextPage: true,
				};
			}
			if (key === 'music-by-genre') {
				return makeInfiniteResult([track], true, true);
			}
			return makeInfiniteResult([]);
		});
		render(<GenresView />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		fireEvent.click(screen.getByText('genre-1'));
		expect(screen.getAllByRole('progressbar').length).toBeGreaterThan(0);
		fireEvent.click(screen.getAllByRole('button')[0]);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
	});

	it('renders folder root name and detail loading state', () => {
		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			if (key === 'music-folders') {
				return makeInfiniteResult([{ folder: '/', track_count: 2 }], true, true);
			}
			return makeInfiniteResult([]);
		});
		mockUseQuery.mockReturnValue({
			data: undefined,
			isLoading: true,
		});

		render(<FoldersView />);
		expect(screen.getByText('/')).toBeInTheDocument();
		fireEvent.click(screen.getByText('/'));
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
	});

	it('renders artist/album fetching spinners in both list and detail views', () => {
		const fetchAlbums = jest.fn();
		const fetchArtists = jest.fn();
		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			if (key === 'music-albums') {
				return {
					data: { pages: [makePagination([{ album: 'album-1', artist: 'artist-1', track_count: 1 }], true)] },
					isLoading: false,
					fetchNextPage: fetchAlbums,
					hasNextPage: true,
					isFetchingNextPage: true,
				};
			}
			if (key === 'music-by-album') {
				return {
					data: { pages: [makePagination([track], true)] },
					isLoading: false,
					fetchNextPage: fetchAlbums,
					hasNextPage: true,
					isFetchingNextPage: true,
				};
			}
			if (key === 'music-artists') {
				return {
					data: { pages: [makePagination([{ artist: 'artist-1', album_count: 1, track_count: 1 }], true)] },
					isLoading: false,
					fetchNextPage: fetchArtists,
					hasNextPage: true,
					isFetchingNextPage: true,
				};
			}
			if (key === 'music-by-artist') {
				return {
					data: { pages: [makePagination([track], true)] },
					isLoading: false,
					fetchNextPage: fetchArtists,
					hasNextPage: true,
					isFetchingNextPage: true,
				};
			}
			return makeInfiniteResult([]);
		});

		const { unmount } = render(<AlbumsView />);
		expect(screen.getByText('album-1')).toBeInTheDocument();
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		fireEvent.click(screen.getByText('album-1'));
		expect(screen.getByText('track-1')).toBeInTheDocument();
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		unmount();

		render(<ArtistsView />);
		expect(screen.getByText('artist-1')).toBeInTheDocument();
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		fireEvent.click(screen.getByText('artist-1'));
		expect(screen.getByText('track-1')).toBeInTheDocument();
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(fetchAlbums).toHaveBeenCalledTimes(0);
		expect(fetchArtists).toHaveBeenCalledTimes(0);
	});

	it('renders load-more text/actions for albums and artists and opens folder add-menu', () => {
		const fetchAlbumsList = jest.fn();
		const fetchAlbumsTracks = jest.fn();
		const fetchArtistsList = jest.fn();
		const fetchArtistsTracks = jest.fn();

		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({});
			options.queryFn?.({ pageParam: 1 });
			if (key === 'music-albums') {
				return {
					data: { pages: [makePagination([{ album: 'album-1', artist: 'artist-1', track_count: 1 }], true)] },
					isLoading: false,
					fetchNextPage: fetchAlbumsList,
					hasNextPage: true,
					isFetchingNextPage: false,
				};
			}
			if (key === 'music-by-album') {
				return {
					data: { pages: [makePagination([track], true)] },
					isLoading: false,
					fetchNextPage: fetchAlbumsTracks,
					hasNextPage: true,
					isFetchingNextPage: false,
				};
			}
			if (key === 'music-artists') {
				return {
					data: { pages: [makePagination([{ artist: 'artist-1', album_count: 1, track_count: 1 }], true)] },
					isLoading: false,
					fetchNextPage: fetchArtistsList,
					hasNextPage: true,
					isFetchingNextPage: false,
				};
			}
			if (key === 'music-by-artist') {
				return {
					data: { pages: [makePagination([track], true)] },
					isLoading: false,
					fetchNextPage: fetchArtistsTracks,
					hasNextPage: true,
					isFetchingNextPage: false,
				};
			}
			if (key === 'music-folders') return makeInfiniteResult([{ folder: '/', track_count: 1 }]);
			return makeInfiniteResult([]);
		});
		mockUseQuery.mockReturnValue({
			data: { items: [track], pagination: { page: 1, has_next: false, has_prev: false, total_pages: 1, total_items: 1 } },
			isLoading: false,
		});

		const { unmount } = render(<AlbumsView />);
		fireEvent.click(screen.getByText('Load more'));
		expect(fetchAlbumsList).toHaveBeenCalled();
		fireEvent.click(screen.getByText('album-1'));
		fireEvent.click(screen.getByText('Load more'));
		expect(fetchAlbumsTracks).toHaveBeenCalled();
		fireEvent.click(screen.getByRole('button', { name: 'add track-1 to playlist' }));
		expect(screen.getByText('MenuAnchor-open')).toBeInTheDocument();
		unmount();

		const artistsRender = render(<ArtistsView />);
		fireEvent.click(screen.getByText('Load more'));
		expect(fetchArtistsList).toHaveBeenCalled();
		fireEvent.click(screen.getByText('artist-1'));
		fireEvent.click(screen.getByText('Load more'));
		expect(fetchArtistsTracks).toHaveBeenCalled();
		fireEvent.click(screen.getByRole('button', { name: 'add track-1 to playlist' }));
		expect(screen.getByText('MenuAnchor-open')).toBeInTheDocument();
		artistsRender.unmount();

		render(<FoldersView />);
		fireEvent.click(screen.getByText('/'));
		fireEvent.click(screen.getByRole('button', { name: 'add track-1 to playlist' }));
		expect(screen.getByText('MenuAnchor-open')).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: 'close-menu' }));
		expect(screen.getByText('MenuAnchor-closed')).toBeInTheDocument();
	});
});

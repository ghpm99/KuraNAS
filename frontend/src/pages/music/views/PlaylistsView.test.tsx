import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import PlaylistsView from './PlaylistsView';

const mockUseInfiniteQuery = jest.fn();
const mockUseMutation = jest.fn();
const mockUseQueryClient = jest.fn();
const mockEnqueueSnackbar = jest.fn();
const mockUseGlobalMusic = jest.fn();
const mockGetPlaylists = jest.fn();
const mockCreatePlaylist = jest.fn();
const mockDeletePlaylist = jest.fn();
const mockGetPlaylistTracks = jest.fn();
const mockRemoveTrackFromPlaylist = jest.fn();

jest.mock('@tanstack/react-query', () => ({
	useInfiniteQuery: (...args: any[]) => mockUseInfiniteQuery(...args),
	useMutation: (...args: any[]) => mockUseMutation(...args),
	useQueryClient: () => mockUseQueryClient(),
}));

jest.mock('@/service/playlist', () => ({
	getPlaylists: (...args: any[]) => mockGetPlaylists(...args),
	createPlaylist: (...args: any[]) => mockCreatePlaylist(...args),
	deletePlaylist: (...args: any[]) => mockDeletePlaylist(...args),
	getPlaylistTracks: (...args: any[]) => mockGetPlaylistTracks(...args),
	removeTrackFromPlaylist: (...args: any[]) => mockRemoveTrackFromPlaylist(...args),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
	useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => k }),
}));

const page = (items: any[], hasNext = false) => ({
	data: {
		pages: [
			{
				items,
				pagination: { page: 1, has_next: hasNext, has_prev: false, total_pages: 1, total_items: items.length },
			},
		],
	},
	isLoading: false,
	fetchNextPage: jest.fn(),
	hasNextPage: hasNext,
	isFetchingNextPage: false,
});

describe('pages/music/views/PlaylistsView', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseQueryClient.mockReturnValue({ invalidateQueries: jest.fn() });
		mockUseGlobalMusic.mockReturnValue({
			getMusicTitle: (m: any) => m.name,
			getMusicArtist: () => 'artist',
			musicMetadata: () => 'meta',
			addToQueue: jest.fn(),
		});
		mockGetPlaylists.mockResolvedValue({
			items: [{ id: 1, name: 'P1', track_count: 1, description: 'desc', is_system: false }],
			pagination: { page: 1, has_next: false, has_prev: false, total_pages: 1, total_items: 1 },
		});
		mockGetPlaylistTracks.mockResolvedValue({
			items: [{ id: 7, file: { id: 90, name: 'track-1', format: 'mp3', size: 1000 } }],
			pagination: { page: 1, has_next: false, has_prev: false, total_pages: 1, total_items: 1 },
		});
		mockCreatePlaylist.mockResolvedValue({ id: 10, name: 'created' });
		mockDeletePlaylist.mockResolvedValue({});
		mockRemoveTrackFromPlaylist.mockResolvedValue({});

		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			options.getNextPageParam?.({ pagination: { has_next: true, page: 1 } });
			options.getNextPageParam?.({ pagination: { has_next: false, page: 1 } });

			if (key === 'playlists') {
				return page([{ id: 1, name: 'P1', track_count: 1, description: 'desc', is_system: false }]);
			}
			return page([{ id: 7, file: { id: 90, name: 'track-1', format: 'mp3', size: 1000 } }]);
		});
		mockUseMutation.mockImplementation((options: any) => ({
			mutate: (...args: any[]) => {
				options.mutationFn?.(...args);
				options.onSuccess?.(...args);
			},
			isPending: false,
		}));
	});

	it('handles list view actions and creation flow', () => {
		const { container } = render(<PlaylistsView />);
		expect(screen.getByText('MUSIC_PLAYLISTS')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'MUSIC_NEW' }));
		fireEvent.change(screen.getByLabelText('NAME'), { target: { value: 'Roadtrip' } });
		fireEvent.change(screen.getByLabelText('MUSIC_DESCRIPTION_OPTIONAL'), { target: { value: 'desc' } });
		fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE' }));
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATED', { variant: 'success' });
		expect(mockCreatePlaylist).toHaveBeenCalledWith({ name: 'Roadtrip', description: 'desc' });

		const deleteButton = container.querySelector('svg.lucide-trash2')?.closest('button') as HTMLElement;
		fireEvent.click(deleteButton);
		expect(mockDeletePlaylist).toHaveBeenCalledWith(1);
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_DELETED', { variant: 'success' });
	});

	it('renders detail view and handles remove flow', () => {
		const { container } = render(<PlaylistsView />);
		fireEvent.click(screen.getByText('P1'));
		expect(screen.getByText('track-1')).toBeInTheDocument();
		fireEvent.click(screen.getByText('track-1'));

		const removeButton = container.querySelector('svg.lucide-trash2')?.closest('button') as HTMLElement;
		fireEvent.click(removeButton);
		expect(mockRemoveTrackFromPlaylist).toHaveBeenCalledWith(1, 90);
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_TRACK_REMOVED', { variant: 'success' });
	});

	it('handles empty detail and mutation errors', () => {
		let mutationCall = 0;
		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			if (key === 'playlists') return page([{ id: 1, name: 'P1', track_count: 0, is_system: false }]);
			return page([]);
		});
		mockUseMutation.mockImplementation((options: any) => {
			mutationCall += 1;
			return {
				mutate: (...args: any[]) => {
					options.mutationFn?.(...args);
					options.onError?.();
				},
				isPending: mutationCall === 1,
			};
		});

		render(<PlaylistsView />);
		fireEvent.click(screen.getByRole('button', { name: 'MUSIC_NEW' }));
		fireEvent.change(screen.getByLabelText('NAME'), { target: { value: 'x' } });
		fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE' }));
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATE_FAILED', { variant: 'error' });

		fireEvent.click(screen.getByText('P1'));
		expect(screen.getByText('MUSIC_PLAYLIST_EMPTY')).toBeInTheDocument();
	});

	it('renders loading and load-more states', () => {
		mockUseInfiniteQuery.mockReturnValueOnce({
			data: undefined,
			isLoading: true,
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});
		const { unmount } = render(<PlaylistsView />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		unmount();

		const fetchNextPage = jest.fn();
		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			if (key === 'playlists') {
				return {
					...page([{ id: 1, name: 'P1', track_count: 1, description: '', is_system: false }], true),
					fetchNextPage,
				};
			}
			return page([{ id: 7, file: { id: 90, name: 'track-1', format: 'mp3', size: 1000 } }]);
		});

		render(<PlaylistsView />);
		fireEvent.click(screen.getByText('ACTION_LOAD_MORE'));
		expect(fetchNextPage).toHaveBeenCalled();
	});

	it('supports load more in detail view', () => {
		const fetchTracksNext = jest.fn();
		mockUseInfiniteQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			options.queryFn?.({ pageParam: 1 });
			if (key === 'playlists') return page([{ id: 1, name: 'P1', track_count: 1, is_system: false }]);
			return {
				...page([{ id: 7, file: { id: 90, name: 'track-1', format: 'mp3', size: 1000 } }], true),
				fetchNextPage: fetchTracksNext,
			};
		});
		render(<PlaylistsView />);
		fireEvent.click(screen.getByText('P1'));
		fireEvent.click(screen.getByText('ACTION_LOAD_MORE'));
		expect(fetchTracksNext).toHaveBeenCalled();
	});
});

import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import PlaylistDetailSection from './PlaylistDetailSection';
import PlaylistListSection from './PlaylistListSection';

const mockUseGlobalMusic = jest.fn();
const mockReplaceQueue = jest.fn();
const mockGetPlaylistTracks = jest.fn();

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
	useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
	getMusicTitle: (m: any) => m.name ?? m.metadata?.title ?? '',
	getMusicArtist: (m: any) => m.metadata?.artist ?? 'Unknown Artist',
	musicMetadata: () => 'meta',
	formatMusicDuration: (s: number) => `${Math.floor(s/60)}:${String(Math.floor(s%60)).padStart(2,'0')}`,
}));

jest.mock('@/components/hooks/usePlaylistTrackHandlers/usePlaylistTrackHandlers', () => ({
	usePlaylistTrackHandlers: () => ({
		getMusicTitle: (track: any) => track.name,
		getMusicArtist: (track: any) => track.metadata?.artist ?? 'artist',
	}),
}));

jest.mock('@/service/playlist', () => ({
	getPlaylistTracks: (...args: any[]) => mockGetPlaylistTracks(...args),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const playlist = {
	id: 1,
	name: 'Roadtrip',
	description: 'favorites',
	is_system: false,
	is_auto: false,
	kind: 'manual',
	source_key: '',
	created_at: '',
	updated_at: '',
	track_count: 2,
};

const systemPlaylist = {
	...playlist,
	id: 2,
	name: 'System Mix',
	is_system: true,
	is_auto: true,
	kind: 'automatic',
	source_key: 'favorites',
};

const tracks = [
	{
		id: 10,
		position: 1,
		added_at: '',
		file: {
			id: 100,
			name: 'track-1',
			format: 'mp3',
			size: 1000,
			metadata: { artist: 'Artist 1', duration: 180 },
		},
	},
	{
		id: 11,
		position: 2,
		added_at: '',
		file: {
			id: 101,
			name: 'track-2',
			format: 'mp3',
			size: 1000,
			metadata: { artist: 'Artist 2', duration: 120 },
		},
	},
];

describe('playlist sections', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseGlobalMusic.mockReturnValue({
			replaceQueue: mockReplaceQueue,
			currentTrack: { id: 200 },
			isPlaying: false,
		});
		mockGetPlaylistTracks.mockResolvedValue({
			items: tracks,
			pagination: { page: 1, has_next: false, has_prev: false },
		});
	});

	it('handles list section actions and play context', async () => {
		const onSelect = jest.fn();
		const onDelete = jest.fn();
		const onLoadMore = jest.fn();
		const onCreateOpen = jest.fn();

		const { container } = render(
			<PlaylistListSection
				playlists={[playlist, systemPlaylist]}
				isLoading={false}
				hasNextPage={true}
				isFetchingNextPage={false}
				onSelect={onSelect}
				onDelete={onDelete}
				onLoadMore={onLoadMore}
				onCreateOpen={onCreateOpen}
			/>,
		);

		fireEvent.click(screen.getByText('Roadtrip'));
		expect(onSelect).toHaveBeenCalledWith(playlist);

		const playButtons = container.querySelectorAll('svg.lucide-play');
		fireEvent.click(playButtons[0]!.closest('button') as HTMLElement);
		expect(mockGetPlaylistTracks).toHaveBeenCalledWith(1, 1, 200);
		await waitFor(() =>
			expect(mockReplaceQueue).toHaveBeenCalledWith(
				[expect.objectContaining({ id: 100 }), expect.objectContaining({ id: 101 })],
				0,
				expect.objectContaining({ kind: 'playlist' }),
			),
		);

		const deleteButtons = container.querySelectorAll('svg.lucide-trash2');
		fireEvent.click(deleteButtons[0]!.closest('button') as HTMLElement);
		expect(onDelete).toHaveBeenCalledWith(1);

		fireEvent.click(screen.getByRole('button', { name: 'MUSIC_NEW' }));
		expect(onCreateOpen).toHaveBeenCalled();

		fireEvent.click(screen.getByText('ACTION_LOAD_MORE'));
		expect(onLoadMore).toHaveBeenCalled();
	});

	it('handles detail section actions, empty state, and loading state', () => {
		const onBack = jest.fn();
		const onRemoveTrack = jest.fn();
		const onLoadMore = jest.fn();
		const { container, rerender } = render(
			<PlaylistDetailSection
				playlist={playlist}
				tracks={tracks}
				isLoading={false}
				hasNextPage={true}
				isFetchingNextPage={false}
				onBack={onBack}
				onRemoveTrack={onRemoveTrack}
				onLoadMore={onLoadMore}
			/>,
		);

		fireEvent.click(screen.getByText('track-2'));
		expect(mockReplaceQueue).toHaveBeenCalledWith(
			[expect.objectContaining({ id: 100 }), expect.objectContaining({ id: 101 })],
			1,
			expect.objectContaining({ kind: 'playlist' }),
		);

		const actionButtons = screen.getAllByRole('button');
		fireEvent.click(actionButtons[0]!);
		fireEvent.click(actionButtons[1]!);
		fireEvent.click(actionButtons[2]!);
		expect(onBack).toHaveBeenCalled();
		expect(mockReplaceQueue).toHaveBeenCalledWith(
			[expect.objectContaining({ id: 100 }), expect.objectContaining({ id: 101 })],
			0,
			expect.objectContaining({ kind: 'playlist' }),
		);
		expect(mockReplaceQueue).toHaveBeenCalledWith(
			expect.arrayContaining([expect.objectContaining({ id: 100 }), expect.objectContaining({ id: 101 })]),
			0,
			expect.objectContaining({ kind: 'playlist' }),
		);

		const removeButton = container.querySelector('svg.lucide-trash2')?.closest('button') as HTMLElement;
		fireEvent.click(removeButton);
		expect(onRemoveTrack).toHaveBeenCalledWith(100);

		fireEvent.click(screen.getByText('ACTION_LOAD_MORE'));
		expect(onLoadMore).toHaveBeenCalled();

		rerender(
			<PlaylistDetailSection
				playlist={playlist}
				tracks={[]}
				isLoading={false}
				hasNextPage={false}
				isFetchingNextPage={false}
				onBack={onBack}
				onRemoveTrack={onRemoveTrack}
				onLoadMore={onLoadMore}
			/>,
		);
		expect(screen.getByText('MUSIC_PLAYLIST_EMPTY')).toBeInTheDocument();

		rerender(
			<PlaylistDetailSection
				playlist={playlist}
				tracks={tracks}
				isLoading={true}
				hasNextPage={false}
				isFetchingNextPage={false}
				onBack={onBack}
				onRemoveTrack={onRemoveTrack}
				onLoadMore={onLoadMore}
			/>,
		);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
	});

	it('covers list section keyboard, loading, spinner, and empty-playlist branches', async () => {
		const onSelect = jest.fn();
		const onDelete = jest.fn();
		const onLoadMore = jest.fn();
		const onCreateOpen = jest.fn();

		const { container, rerender } = render(
			<PlaylistListSection
				playlists={[playlist]}
				isLoading={true}
				hasNextPage={false}
				isFetchingNextPage={false}
				onSelect={onSelect}
				onDelete={onDelete}
				onLoadMore={onLoadMore}
				onCreateOpen={onCreateOpen}
			/>,
		);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();

		rerender(
			<PlaylistListSection
				playlists={[playlist]}
				isLoading={false}
				hasNextPage={true}
				isFetchingNextPage={true}
				onSelect={onSelect}
				onDelete={onDelete}
				onLoadMore={onLoadMore}
				onCreateOpen={onCreateOpen}
			/>,
		);

		const listItemButton = screen.getByText('Roadtrip').closest('[role="button"]') as HTMLElement;
		fireEvent.keyDown(listItemButton, { key: 'Enter' });
		fireEvent.keyDown(listItemButton, { key: ' ' });
		expect(onSelect).toHaveBeenCalledTimes(3);
		fireEvent.keyDown(listItemButton, { key: 'Escape' });
		expect(onSelect).toHaveBeenCalledTimes(3);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();

		mockGetPlaylistTracks.mockResolvedValueOnce({
			items: [],
			pagination: { page: 1, has_next: false, has_prev: false },
		});
		fireEvent.click(container.querySelector('svg.lucide-play')?.closest('button') as HTMLElement);
		await waitFor(() => {
			expect(mockGetPlaylistTracks).toHaveBeenCalledWith(1, 1, 200);
		});
		expect(mockReplaceQueue).not.toHaveBeenCalled();

		rerender(
			<PlaylistListSection
				playlists={[]}
				isLoading={false}
				hasNextPage={false}
				isFetchingNextPage={false}
				onSelect={onSelect}
				onDelete={onDelete}
				onLoadMore={onLoadMore}
				onCreateOpen={onCreateOpen}
			/>,
		);

		expect(screen.getByText('MUSIC_NO_PLAYLISTS_MSG')).toBeInTheDocument();
	});

	it('covers detail section keyboard, active-track states, and no-track actions', () => {
		const onBack = jest.fn();
		const onRemoveTrack = jest.fn();
		const onLoadMore = jest.fn();
		const playlistWithoutDescription = { ...playlist, description: '' };
		const trackWithoutDuration = {
			...tracks[0],
			file: {
				...tracks[0].file,
				metadata: {
					...tracks[0].file.metadata,
					duration: undefined,
				},
			},
		};

		const { container, rerender } = render(
			<PlaylistDetailSection
				playlist={playlistWithoutDescription}
				tracks={tracks}
				isLoading={false}
				hasNextPage={true}
				isFetchingNextPage={true}
				onBack={onBack}
				onRemoveTrack={onRemoveTrack}
				onLoadMore={onLoadMore}
			/>,
		);

		const firstTrackRow = screen.getByText('track-1').closest('[role="button"]') as HTMLElement;
		fireEvent.keyDown(firstTrackRow, { key: 'Enter' });
		fireEvent.keyDown(firstTrackRow, { key: ' ' });
		fireEvent.keyDown(firstTrackRow, { key: 'Escape' });
		expect(mockReplaceQueue).toHaveBeenCalledWith(
			[expect.objectContaining({ id: 100 }), expect.objectContaining({ id: 101 })],
			0,
			expect.objectContaining({ kind: 'playlist' }),
		);
		expect(screen.getAllByRole('progressbar').length).toBeGreaterThan(0);
		expect(screen.queryByText('favorites')).not.toBeInTheDocument();

		mockUseGlobalMusic.mockReturnValue({
			replaceQueue: mockReplaceQueue,
			currentTrack: { id: 100 },
			isPlaying: false,
		});
		rerender(
			<PlaylistDetailSection
				playlist={playlistWithoutDescription}
				tracks={tracks}
				isLoading={false}
				hasNextPage={false}
				isFetchingNextPage={false}
				onBack={onBack}
				onRemoveTrack={onRemoveTrack}
				onLoadMore={onLoadMore}
			/>,
		);
		expect(container.querySelector('svg.lucide-pause')).toBeInTheDocument();

		mockUseGlobalMusic.mockReturnValue({
			replaceQueue: mockReplaceQueue,
			currentTrack: { id: 100 },
			isPlaying: true,
		});
		rerender(
			<PlaylistDetailSection
				playlist={playlistWithoutDescription}
				tracks={[trackWithoutDuration, tracks[1]!]}
				isLoading={false}
				hasNextPage={false}
				isFetchingNextPage={false}
				onBack={onBack}
				onRemoveTrack={onRemoveTrack}
				onLoadMore={onLoadMore}
			/>,
		);

		const playingTrackRow = screen.getByText('track-1').closest('[role="button"]') as HTMLElement;
		expect(within(playingTrackRow).queryByText('1')).not.toBeInTheDocument();
		expect(within(playingTrackRow).queryByText('3:00')).not.toBeInTheDocument();
		expect(container.querySelector('svg.lucide-pause')).not.toBeInTheDocument();

		const replaceQueueCallCount = mockReplaceQueue.mock.calls.length;
		rerender(
			<PlaylistDetailSection
				playlist={playlistWithoutDescription}
				tracks={[]}
				isLoading={false}
				hasNextPage={false}
				isFetchingNextPage={false}
				onBack={onBack}
				onRemoveTrack={onRemoveTrack}
				onLoadMore={onLoadMore}
			/>,
		);

		const buttons = screen.getAllByRole('button');
		fireEvent.click(buttons[1]!);
		fireEvent.click(buttons[2]!);
		expect(mockReplaceQueue).toHaveBeenCalledTimes(replaceQueueCallCount);
	});
});

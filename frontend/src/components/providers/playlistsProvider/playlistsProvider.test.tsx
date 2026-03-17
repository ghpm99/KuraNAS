import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { PlaylistsProvider, usePlaylistsProvider } from './playlistsProvider';
import type { ReactNode } from 'react';

const mockGetPlaylists = jest.fn();
const mockGetAutomaticPlaylists = jest.fn();
const mockGetPlaylistTracks = jest.fn();
const mockCreatePlaylist = jest.fn();
const mockDeletePlaylist = jest.fn();
const mockRemoveTrackFromPlaylist = jest.fn();
const mockEnqueueSnackbar = jest.fn();

jest.mock('@/service/playlist', () => ({
	getPlaylists: (...args: unknown[]) => mockGetPlaylists(...args),
	getAutomaticPlaylists: (...args: unknown[]) => mockGetAutomaticPlaylists(...args),
	getPlaylistTracks: (...args: unknown[]) => mockGetPlaylistTracks(...args),
	createPlaylist: (...args: unknown[]) => mockCreatePlaylist(...args),
	deletePlaylist: (...args: unknown[]) => mockDeletePlaylist(...args),
	removeTrackFromPlaylist: (...args: unknown[]) => mockRemoveTrackFromPlaylist(...args),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

const createPlaylistItem = (overrides: Record<string, unknown> = {}) => ({
	id: 1,
	name: 'Test Playlist',
	description: 'A test playlist',
	is_system: false,
	is_auto: false,
	kind: 'manual',
	source_key: '',
	created_at: '2026-01-01T00:00:00Z',
	updated_at: '2026-01-01T00:00:00Z',
	track_count: 5,
	...overrides,
});

const createQueryClient = () =>
	new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});

const createWrapper = (initialEntries: string[] = ['/music/playlists']) => {
	const queryClient = createQueryClient();
	return ({ children }: { children: ReactNode }) => (
		<QueryClientProvider client={queryClient}>
			<MemoryRouter initialEntries={initialEntries}>
				<PlaylistsProvider>{children}</PlaylistsProvider>
			</MemoryRouter>
		</QueryClientProvider>
	);
};

const setupDefaultMocks = () => {
	mockGetAutomaticPlaylists.mockResolvedValue([]);
	mockGetPlaylists.mockResolvedValue({
		items: [],
		pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
	});
	mockGetPlaylistTracks.mockResolvedValue({
		items: [],
		pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
	});
	mockCreatePlaylist.mockResolvedValue(createPlaylistItem());
	mockDeletePlaylist.mockResolvedValue(undefined);
	mockRemoveTrackFromPlaylist.mockResolvedValue(undefined);
};

describe('PlaylistsProvider', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		setupDefaultMocks();
	});

	it('throws when usePlaylistsProvider is used outside the provider', () => {
		expect(() => {
			renderHook(() => usePlaylistsProvider());
		}).toThrow('usePlaylistsProvider must be used within PlaylistsProvider');
	});

	it('provides default state', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		expect(result.current.playlists).toEqual([]);
		expect(result.current.tracks).toEqual([]);
		expect(result.current.selectedPlaylist).toBeNull();
		expect(result.current.createOpen).toBe(false);
	});

	it('creates a playlist and shows success snackbar', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		act(() => result.current.setNewName('My Playlist'));
		act(() => result.current.setNewDescription('Description'));
		act(() => result.current.openCreateDialog());

		expect(result.current.createOpen).toBe(true);

		act(() => result.current.submitCreatePlaylist());

		await waitFor(() => expect(mockCreatePlaylist).toHaveBeenCalled());
		await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATED', { variant: 'success' }));
		expect(result.current.createOpen).toBe(false);
		expect(result.current.newName).toBe('');
		expect(result.current.newDescription).toBe('');
	});

	it('shows error snackbar on create failure', async () => {
		mockCreatePlaylist.mockRejectedValue(new Error('fail'));

		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		act(() => result.current.setNewName('Fail Playlist'));
		act(() => result.current.submitCreatePlaylist());

		await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATE_FAILED', { variant: 'error' }));
	});

	it('deletes a playlist and shows success snackbar', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		act(() => result.current.deletePlaylistById(5));

		await waitFor(() => expect(mockDeletePlaylist).toHaveBeenCalledWith(5));
		await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_DELETED', { variant: 'success' }));
	});

	it('shows error snackbar on delete failure', async () => {
		mockDeletePlaylist.mockRejectedValue(new Error('fail'));

		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		act(() => result.current.deletePlaylistById(5));

		await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_DELETE_FAILED', { variant: 'error' }));
	});

	it('removes a track when a playlist is selected and shows success snackbar', async () => {
		mockGetAutomaticPlaylists.mockResolvedValue([]);
		mockGetPlaylists.mockResolvedValue({
			items: [createPlaylistItem({ id: 10 })],
			pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
		});

		const { result } = renderHook(() => usePlaylistsProvider(), {
			wrapper: createWrapper(['/music/playlists?playlist=10']),
		});

		await waitFor(() => expect(result.current.selectedPlaylist).not.toBeNull());

		act(() => result.current.removeTrackByFileId(42));

		await waitFor(() => expect(mockRemoveTrackFromPlaylist).toHaveBeenCalledWith(10, 42));
		await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_TRACK_REMOVED', { variant: 'success' }));
	});

	it('remove mutation resolves immediately when no playlist is selected', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		act(() => result.current.removeTrackByFileId(42));

		await waitFor(() => expect(result.current.isRemovingTrack).toBe(false));
		expect(mockRemoveTrackFromPlaylist).not.toHaveBeenCalled();
	});

	it('shows error snackbar on remove track failure', async () => {
		mockGetPlaylists.mockResolvedValue({
			items: [createPlaylistItem({ id: 10 })],
			pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
		});
		mockRemoveTrackFromPlaylist.mockRejectedValue(new Error('fail'));

		const { result } = renderHook(() => usePlaylistsProvider(), {
			wrapper: createWrapper(['/music/playlists?playlist=10']),
		});

		await waitFor(() => expect(result.current.selectedPlaylist).not.toBeNull());

		act(() => result.current.removeTrackByFileId(42));

		await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_TRACK_REMOVE_FAILED', { variant: 'error' }));
	});

	it('backToList removes playlist param from search params', async () => {
		mockGetPlaylists.mockResolvedValue({
			items: [createPlaylistItem({ id: 10 })],
			pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
		});

		const { result } = renderHook(() => usePlaylistsProvider(), {
			wrapper: createWrapper(['/music/playlists?playlist=10']),
		});

		await waitFor(() => expect(result.current.selectedPlaylist).not.toBeNull());

		act(() => result.current.backToList());

		await waitFor(() => expect(result.current.selectedPlaylist).toBeNull());
	});

	it('closeCreateDialog sets createOpen to false', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		act(() => result.current.openCreateDialog());
		expect(result.current.createOpen).toBe(true);

		act(() => result.current.closeCreateDialog());
		expect(result.current.createOpen).toBe(false);
	});

	it('selectPlaylist sets playlist search param', async () => {
		mockGetPlaylists.mockResolvedValue({
			items: [createPlaylistItem({ id: 10 })],
			pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
		});

		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.playlists).toHaveLength(1));

		act(() => result.current.selectPlaylist(result.current.playlists[0]!));

		await waitFor(() => expect(result.current.selectedPlaylist).not.toBeNull());
		expect(result.current.selectedPlaylist?.id).toBe(10);
	});

	it('combines automatic and manual playlists', async () => {
		mockGetAutomaticPlaylists.mockResolvedValue([
			createPlaylistItem({ id: 100, name: 'Auto Playlist', is_system: true }),
		]);
		mockGetPlaylists.mockResolvedValue({
			items: [createPlaylistItem({ id: 200, name: 'Manual Playlist' })],
			pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
		});

		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.playlists).toHaveLength(2));

		expect(result.current.playlists[0]?.name).toBe('Auto Playlist');
		expect(result.current.playlists[1]?.name).toBe('Manual Playlist');
	});

	it('remove onSuccess does not invalidate tracks query when no playlist is selected', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoadingPlaylists).toBe(false));

		act(() => result.current.removeTrackByFileId(42));

		await waitFor(() => expect(result.current.isRemovingTrack).toBe(false));
		// No snackbar for success when no playlist selected (early return in onSuccess)
		expect(mockEnqueueSnackbar).not.toHaveBeenCalledWith('MUSIC_TRACK_REMOVED', expect.anything());
	});
});

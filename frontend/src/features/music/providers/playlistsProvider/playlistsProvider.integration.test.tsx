import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { PlaylistsProvider, usePlaylistsProvider } from './playlistsProvider';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real PlaylistsProvider +
// service/playlist.ts run, so each command asserts the exact endpoint/payload
// the backend music playlist handlers decode.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: jest.fn() }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; delete: jest.Mock };

const wrapper = ({ children }: { children: ReactNode }) => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return (
		<QueryClientProvider client={client}>
			<MemoryRouter>
				<PlaylistsProvider>{children}</PlaylistsProvider>
			</MemoryRouter>
		</QueryClientProvider>
	);
};

describe('features/music/playlistsProvider (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/music/playlists/system') return Promise.resolve({ data: [] });
			if (url.includes('/tracks')) {
				return Promise.resolve({ data: { items: [], pagination: { has_next: false } } });
			}
			return Promise.resolve({
				data: { items: [{ id: 3, name: 'X', is_system: false }], pagination: { has_next: false } },
			});
		});
		mockedApi.post.mockResolvedValue({ data: { id: 9, name: 'Nova' } });
		mockedApi.delete.mockResolvedValue({ data: undefined });
	});

	it('deletePlaylistById DELETEs /music/playlists/:id', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper });

		act(() => result.current.deletePlaylistById(3));

		await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/music/playlists/3'));
	});

	it('removeTrackByFileId DELETEs the track from the selected playlist', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper });
		await waitFor(() => expect(result.current.playlists.some((p) => p.id === 3)).toBe(true));

		act(() => result.current.selectPlaylist({ id: 3, name: 'X' } as never));
		await waitFor(() => expect(result.current.selectedPlaylist?.id).toBe(3));

		act(() => result.current.removeTrackByFileId(7));

		await waitFor(() =>
			expect(mockedApi.delete).toHaveBeenCalledWith('/music/playlists/3/tracks/7')
		);
	});

	it('submitCreatePlaylist POSTs /music/playlists/ with name and description', async () => {
		const { result } = renderHook(() => usePlaylistsProvider(), { wrapper });

		act(() => result.current.setNewName('Nova'));
		act(() => result.current.submitCreatePlaylist());

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/', { name: 'Nova', description: '' })
		);
	});
});

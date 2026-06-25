import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import AddToPlaylistMenu from './AddToPlaylistMenu';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real AddToPlaylistMenu +
// service/playlist.ts run, so each command asserts the exact endpoint/payload
// the backend music playlist handlers decode.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const playlistsPage = {
	items: [{ id: 5, name: 'Favoritas', is_system: false }],
	pagination: { page: 1, page_size: 100, has_next: false, has_prev: false },
};

const renderMenu = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<AddToPlaylistMenu fileId={42} anchorEl={document.body} onClose={jest.fn()} />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('features/music/AddToPlaylistMenu (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: playlistsPage });
		mockedApi.post.mockImplementation((url: string) => {
			if (url === '/music/playlists/') return Promise.resolve({ data: { id: 9, name: 'Nova' } });
			return Promise.resolve({ data: { id: 1 } });
		});
	});

	it('adding to an existing playlist POSTs the file id to its tracks endpoint', async () => {
		renderMenu();
		await screen.findByText('Favoritas');

		fireEvent.click(screen.getByText('Favoritas'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/5/tracks', { file_id: 42 })
		);
	});

	it('creating a playlist POSTs the name then adds the track', async () => {
		renderMenu();
		await screen.findByText('MUSIC_NEW_PLAYLIST');

		fireEvent.click(screen.getByText('MUSIC_NEW_PLAYLIST'));
		fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_NAME'), { target: { value: 'Nova' } });
		fireEvent.click(screen.getByText('ACTION_CREATE_ADD'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/', { name: 'Nova' })
		);
		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/9/tracks', { file_id: 42 })
		);
	});
});

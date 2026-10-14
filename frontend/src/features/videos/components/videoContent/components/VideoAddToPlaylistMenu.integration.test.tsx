import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import VideoAddToPlaylistMenu from './VideoAddToPlaylistMenu';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const pageOf = (names: Array<[number, string]>) => ({
    items: names.map(([id, name]) => ({ id, name, type: 'custom' })),
    pagination: { page: 1, page_size: 30, has_next: false, has_prev: false },
});

const renderMenu = (onSelectPlaylist = jest.fn()) => {
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <VideoAddToPlaylistMenu
                videoId={5}
                anchorEl={document.body}
                onClose={jest.fn()}
                onSelectPlaylist={onSelectPlaylist}
            />
        </QueryClientProvider>
    );
    return onSelectPlaylist;
};

describe('features/videos/VideoAddToPlaylistMenu (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string, config?: { params?: { q?: string } }) => {
            if (url === '/video/playlists/by-video/5') {
                return Promise.resolve({ data: [{ id: 2, name: 'Favorites', type: 'custom' }] });
            }
            const query = config?.params?.q;
            const items: Array<[number, string]> =
                query === 'fav'
                    ? [[2, 'Favorites']]
                    : [
                          [1, 'Season 1'],
                          [2, 'Favorites'],
                      ];
            return Promise.resolve({ data: pageOf(items) });
        });
    });

    it('lists the searched playlists through the paginated search endpoint', async () => {
        renderMenu();

        await screen.findByText('Season 1');

        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/search', {
            params: { q: '', page: 1, page_size: 30 },
        });
        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/by-video/5');
    });

    it('sends the typed text as the search query', async () => {
        renderMenu();
        await screen.findByText('Season 1');

        fireEvent.change(screen.getByLabelText('VIDEO_PLAYLIST_SEARCH'), {
            target: { value: 'fav' },
        });

        await waitFor(() =>
            expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/search', {
                params: { q: 'fav', page: 1, page_size: 30 },
            })
        );
        await waitFor(() => expect(screen.queryByText('Season 1')).not.toBeInTheDocument());
    });

    it('selects the chosen playlist and nothing before the choice', async () => {
        const onSelectPlaylist = renderMenu();
        await screen.findByText('Season 1');
        expect(onSelectPlaylist).not.toHaveBeenCalled();

        fireEvent.click(screen.getByText('Season 1'));

        expect(onSelectPlaylist).toHaveBeenCalledTimes(1);
        expect(onSelectPlaylist).toHaveBeenCalledWith(1);
    });

    it('disables the playlists that already contain the video', async () => {
        const onSelectPlaylist = renderMenu();
        await screen.findByText('VIDEO_ALREADY_ADDED');

        const containingItem = screen.getByText('Favorites').closest('[role="menuitem"]');
        expect(containingItem).toHaveAttribute('aria-disabled', 'true');
        fireEvent.click(screen.getByText('Favorites'));

        expect(onSelectPlaylist).not.toHaveBeenCalled();
    });

    it('shows the empty state when no playlist matches', async () => {
        mockedApi.get.mockImplementation((url: string) =>
            Promise.resolve({ data: url.includes('by-video') ? [] : pageOf([]) })
        );
        renderMenu();

        expect(await screen.findByText('VIDEO_NO_PLAYLISTS')).toBeInTheDocument();
    });
});

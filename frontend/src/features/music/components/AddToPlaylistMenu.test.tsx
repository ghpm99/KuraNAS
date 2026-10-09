import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AddToPlaylistMenu, { AddToPlaylistButton } from './AddToPlaylistMenu';

const mockEnqueueSnackbar = jest.fn();
const mockGetPlaylists = jest.fn();
const mockAddTrackToPlaylist = jest.fn();
const mockCreatePlaylist = jest.fn();

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/playlist', () => ({
    getPlaylists: (...args: unknown[]) => mockGetPlaylists(...args),
    addTrackToPlaylist: (...args: unknown[]) => mockAddTrackToPlaylist(...args),
    createPlaylist: (...args: unknown[]) => mockCreatePlaylist(...args),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const paginationOf = (page: number, hasNext: boolean) => ({
    page,
    page_size: 30,
    has_next: hasNext,
    has_prev: page > 1,
});

const userPlaylist = (id: number, name: string) => ({
    id,
    name,
    is_system: false,
    is_auto: false,
    is_ai_generated: false,
});

describe('components/music/AddToPlaylistMenu', () => {
    const onClose = jest.fn();
    const anchor = document.createElement('button');

    const renderMenu = (props?: { fileId?: number }) => {
        const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        return render(
            <QueryClientProvider client={client}>
                <AddToPlaylistMenu fileId={props?.fileId ?? 5} anchorEl={anchor} onClose={onClose} />
            </QueryClientProvider>
        );
    };

    beforeEach(() => {
        jest.clearAllMocks();
        mockGetPlaylists.mockResolvedValue({
            items: [
                userPlaylist(10, 'My Playlist'),
                { ...userPlaylist(11, 'System Playlist'), is_system: true, is_auto: true },
                { ...userPlaylist(12, 'Suggested Mix'), is_ai_generated: true },
            ],
            pagination: paginationOf(1, false),
        });
        mockAddTrackToPlaylist.mockResolvedValue({});
        mockCreatePlaylist.mockResolvedValue({ id: 99, name: 'Roadtrip' });
    });

    it('shows loading state while the first page is pending', () => {
        mockGetPlaylists.mockReturnValue(new Promise(() => undefined));
        renderMenu();
        expect(screen.getByText('LOADING')).toBeInTheDocument();
    });

    it('lists only own playlists and adds the track to the selected one', async () => {
        renderMenu();

        await screen.findByText('My Playlist');
        expect(screen.queryByText('System Playlist')).not.toBeInTheDocument();
        expect(screen.queryByText('Suggested Mix')).not.toBeInTheDocument();

        fireEvent.click(screen.getByText('My Playlist'));

        await waitFor(() => expect(mockAddTrackToPlaylist).toHaveBeenCalledWith(10, 5));
        await waitFor(() => expect(onClose).toHaveBeenCalled());
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_TRACK_ADDED', {
            variant: 'success',
        });
    });

    it('shows the backend message verbatim when the track is already in the playlist', async () => {
        mockAddTrackToPlaylist.mockRejectedValue({
            response: { status: 409, data: { error: 'Esta faixa já está na playlist' } },
        });
        renderMenu();

        fireEvent.click(await screen.findByText('My Playlist'));

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Esta faixa já está na playlist', {
                variant: 'warning',
            })
        );
        expect(onClose).not.toHaveBeenCalled();
    });

    it('falls back to the translated failure when the backend sends no message', async () => {
        mockAddTrackToPlaylist.mockRejectedValue(new Error('network'));
        renderMenu();

        fireEvent.click(await screen.findByText('My Playlist'));

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_TRACK_ADD_FAILED', {
                variant: 'warning',
            })
        );
    });

    it('creates a playlist and adds the track', async () => {
        renderMenu();
        fireEvent.click(await screen.findByText('MUSIC_NEW_PLAYLIST'));
        expect(screen.getByText('MUSIC_CREATE_PLAYLIST_ADD')).toBeInTheDocument();

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_NAME'), {
            target: { value: 'Roadtrip' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE_ADD' }));

        await waitFor(() => {
            expect(mockCreatePlaylist).toHaveBeenCalledWith({ name: 'Roadtrip' });
            expect(mockAddTrackToPlaylist).toHaveBeenCalledWith(99, 5);
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATED_ADDED', {
                variant: 'success',
            });
        });
    });

    it('reports a failed playlist creation', async () => {
        mockCreatePlaylist.mockRejectedValue(new Error('fails'));
        renderMenu();
        fireEvent.click(await screen.findByText('MUSIC_NEW_PLAYLIST'));
        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_NAME'), {
            target: { value: 'Fails' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE_ADD' }));

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATE_FAILED', {
                variant: 'error',
            })
        );
    });

    it('shows the empty entry when there are no playlists', async () => {
        mockGetPlaylists.mockResolvedValue({ items: [], pagination: paginationOf(1, false) });
        renderMenu();
        expect(await screen.findByText('MUSIC_NO_PLAYLISTS')).toBeInTheDocument();
    });

    it('searches playlists by name through the server after the debounce', async () => {
        renderMenu();
        await screen.findByText('My Playlist');
        jest.useFakeTimers();

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_SEARCH'), {
            target: { value: 'road' },
        });
        await act(async () => {
            jest.advanceTimersByTime(400);
        });
        jest.useRealTimers();

        await waitFor(() => expect(mockGetPlaylists).toHaveBeenCalledWith(1, 30, 'road'));
    });

    it('shows a search specific empty state when nothing matches', async () => {
        renderMenu();
        await screen.findByText('My Playlist');
        mockGetPlaylists.mockResolvedValue({ items: [], pagination: paginationOf(1, false) });

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_SEARCH'), {
            target: { value: 'zzz' },
        });

        expect(await screen.findByText('MUSIC_PLAYLIST_SEARCH_EMPTY')).toBeInTheDocument();
    });

    it('loads the next page of playlists through the load more button', async () => {
        mockGetPlaylists.mockImplementation((page: number) =>
            Promise.resolve(
                page === 1
                    ? { items: [userPlaylist(10, 'First')], pagination: paginationOf(1, true) }
                    : { items: [userPlaylist(20, 'Second')], pagination: paginationOf(2, false) }
            )
        );
        renderMenu();
        await screen.findByText('First');

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByText('Second')).toBeInTheDocument();
        expect(mockGetPlaylists).toHaveBeenCalledWith(2, 30, '');
    });

    it('opens the menu from the button wrapper', async () => {
        const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        render(
            <QueryClientProvider client={client}>
                <AddToPlaylistButton fileId={6} />
            </QueryClientProvider>
        );

        fireEvent.click(screen.getAllByRole('button')[0]!);

        expect(await screen.findByText('My Playlist')).toBeInTheDocument();
    });
});

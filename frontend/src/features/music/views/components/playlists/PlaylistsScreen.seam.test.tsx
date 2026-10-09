import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { SnackbarProvider } from 'notistack';
import { PlaylistsProvider } from '@/features/music/providers/playlistsProvider';
import { apiBase } from '@/service';
import PlaylistsScreen from './PlaylistsScreen';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    put: jest.Mock;
    delete: jest.Mock;
};

const pagination = { page: 1, page_size: 50, has_next: false, has_prev: false };

const trackNamed = (id: number, name: string) => ({
    id,
    position: id,
    added_at: '',
    file: { id: id * 10, name, metadata: { title: name, artist: 'Band' } },
});

const renderScreen = (initialEntry: string) => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <MemoryRouter initialEntries={[initialEntry]}>
                    <PlaylistsProvider>
                        <PlaylistsScreen />
                    </PlaylistsProvider>
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('features/music/PlaylistsScreen editing (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/music/playlists/system') return Promise.resolve({ data: [] });
            if (url.endsWith('/tracks')) {
                return Promise.resolve({
                    data: {
                        items: [trackNamed(1, 'Alpha'), trackNamed(2, 'Beta')],
                        pagination,
                    },
                });
            }
            return Promise.resolve({
                data: {
                    items: [
                        {
                            id: 3,
                            name: 'Mix',
                            description: 'desc',
                            is_system: false,
                            is_auto: false,
                            is_ai_generated: false,
                            track_count: 2,
                        },
                        {
                            id: 4,
                            name: 'Suggested',
                            description: '',
                            is_system: false,
                            is_auto: false,
                            is_ai_generated: true,
                            track_count: 5,
                        },
                    ],
                    pagination,
                },
            });
        });
        mockedApi.put.mockResolvedValue({ data: {} });
        mockedApi.delete.mockResolvedValue({ data: {} });
    });

    it('deletes only after the confirmation dialog is accepted', async () => {
        renderScreen('/');
        await screen.findByText('Mix');

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        expect(mockedApi.delete).not.toHaveBeenCalled();
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'DELETE' }));

        await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/music/playlists/3'));
    });

    it('renames the open playlist with a PUT of name and description', async () => {
        renderScreen('/?playlist=3');
        fireEvent.click(await screen.findByRole('button', { name: 'MUSIC_PLAYLIST_EDIT' }));

        fireEvent.change(screen.getByLabelText('NAME'), { target: { value: 'Road trip' } });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_SAVE' }));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/3', {
                name: 'Road trip',
                description: 'desc',
            })
        );
    });

    it('moves a track down with a PUT of the target position', async () => {
        renderScreen('/?playlist=3');
        await screen.findByText('Alpha');

        fireEvent.click(screen.getAllByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_DOWN' })[0]!);

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/3/tracks/reorder', {
                tracks: [{ file_id: 10, position: 2 }],
            })
        );
    });

    it('moves a track by drag and drop with a PUT of the target position', async () => {
        renderScreen('/?playlist=3');
        const alpha = (await screen.findByText('Alpha')).closest('li') as HTMLElement;
        const beta = screen.getByText('Beta').closest('li') as HTMLElement;

        fireEvent.dragStart(alpha);
        fireEvent.dragOver(beta);
        fireEvent.drop(beta);

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/3/tracks/reorder', {
                tracks: [{ file_id: 10, position: 2 }],
            })
        );
    });

    it('lists the suggestion without delete and opens it read-only', async () => {
        renderScreen('/?playlist=4');

        await screen.findByText('Alpha');

        expect(screen.queryByRole('button', { name: 'MUSIC_PLAYLIST_EDIT' })).not.toBeInTheDocument();
        expect(screen.queryAllByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_DOWN' })).toHaveLength(0);
    });
});

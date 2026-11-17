import { useQuery } from '@tanstack/react-query';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { apiBase } from '@/service';
import GenresView from '@/features/music/views/GenresView';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const genreKey = 'rock n roll';
const renameUrl = '/music/library/genres/rock%20n%20roll';

const emptyPage = { items: [], pagination: { page: 1, page_size: 50, has_next: false } };

const LocationProbe = () => {
    const location = useLocation();
    return <span data-testid="location">{location.search}</span>;
};

const ActiveSearchQuery = ({
    fetchGlobalSearch,
}: {
    fetchGlobalSearch: () => Promise<unknown>;
}) => {
    useQuery({ queryKey: ['global-search', 'rock'], queryFn: fetchGlobalSearch });
    return null;
};

describe('RenameGenreDialog (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === renameUrl) {
                return Promise.resolve({
                    data: {
                        key: genreKey,
                        name: 'Rock n Roll',
                        track_count: 8,
                        total_length_seconds: 60,
                    },
                });
            }
            return Promise.resolve({ data: emptyPage });
        });
        mockedApi.put.mockResolvedValue({
            data: { key: 'rock', name: 'Rock', track_count: 20, merged: true },
        });
    });

    it('renames from the genre page, refetches dependent views and opens the resulting genre', async () => {
        const fetchGlobalSearch = jest.fn().mockResolvedValue([]);
        render(
            <QueryClientProvider
                client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
            >
                <SnackbarProvider>
                    <ActiveSearchQuery fetchGlobalSearch={fetchGlobalSearch} />
                    <MemoryRouter initialEntries={['/?genre=rock%20n%20roll']}>
                        <GenresView />
                        <LocationProbe />
                    </MemoryRouter>
                </SnackbarProvider>
            </QueryClientProvider>
        );
        await waitFor(() => expect(fetchGlobalSearch).toHaveBeenCalledTimes(1));

        fireEvent.click(await screen.findByRole('button', { name: 'MUSIC_GENRE_RENAME' }));
        fireEvent.change(screen.getByLabelText('MUSIC_GENRE_RENAME_FIELD_NAME'), {
            target: { value: 'Rock' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_REVIEW' }));
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_CONFIRM' }));

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
        expect(mockedApi.put).toHaveBeenCalledWith(renameUrl, { name: 'Rock' });
        await waitFor(() => expect(screen.getByTestId('location').textContent).toBe('?genre=rock'));
        await waitFor(() => expect(fetchGlobalSearch).toHaveBeenCalledTimes(2));
        expect(await screen.findByText('MUSIC_GENRE_RENAME_MERGED_DONE')).toBeInTheDocument();
    });
});

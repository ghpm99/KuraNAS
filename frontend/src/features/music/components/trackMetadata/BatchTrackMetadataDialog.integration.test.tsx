import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { apiBase } from '@/service';
import AlbumPage from '@/features/music/views/components/AlbumPage';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useOptionalGlobalMusic: () => undefined,
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const albumKey = 'the band::hits';
const batchUrl = '/music/tracks/metadata';

const ActiveSearchQuery = ({
    fetchGlobalSearch,
}: {
    fetchGlobalSearch: () => Promise<unknown>;
}) => {
    useQuery({ queryKey: ['global-search', 'rock'], queryFn: fetchGlobalSearch });
    return null;
};

describe('BatchTrackMetadataDialog (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({
            data: { items: [], pagination: { page: 1, page_size: 50, has_next: false } },
        });
        mockedApi.put.mockResolvedValue({ data: { matched: 4, updated: 0, failed: [] } });
    });

    it('simulates then applies the album batch and refetches the search', async () => {
        const fetchGlobalSearch = jest.fn().mockResolvedValue([]);
        render(
            <QueryClientProvider
                client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
            >
                <ActiveSearchQuery fetchGlobalSearch={fetchGlobalSearch} />
                <MemoryRouter>
                    <AlbumPage albumKey={albumKey} onBack={jest.fn()} />
                </MemoryRouter>
            </QueryClientProvider>
        );
        await waitFor(() => expect(fetchGlobalSearch).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_EDIT_ALBUM' }));
        const genresInput = screen.getByLabelText('MUSIC_METADATA_FIELD_GENRES');
        fireEvent.change(genresInput, { target: { value: 'Rock' } });
        fireEvent.keyDown(genresInput, { key: 'Enter' });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_REVIEW' }));

        await screen.findByText('MUSIC_BATCH_METADATA_WILL_CHANGE');
        expect(mockedApi.put).toHaveBeenCalledTimes(1);
        expect(mockedApi.put).toHaveBeenLastCalledWith(
            batchUrl,
            { fields: { genres: ['Rock'] }, album_key: albumKey },
            { params: { dry_run: true } }
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_CONFIRM' }));

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(2));
        expect(mockedApi.put).toHaveBeenLastCalledWith(batchUrl, {
            fields: { genres: ['Rock'] },
            album_key: albumKey,
        });
        await waitFor(() => expect(fetchGlobalSearch).toHaveBeenCalledTimes(2));
    });
});

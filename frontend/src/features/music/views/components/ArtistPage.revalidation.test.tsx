import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { buildServerPage } from '@/shared/test/buildServerPage';
import { getMusicAlbumsByArtist, getMusicArtistSummary, getMusicByArtist } from '@/service/music';
import ArtistPage from './ArtistPage';

jest.mock('@/service/music', () => ({
    getMusicAlbumsByArtist: jest.fn(),
    getMusicArtistSummary: jest.fn(),
    getMusicByArtist: jest.fn(),
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const albumsService = getMusicAlbumsByArtist as jest.Mock;

const firstAlbum = {
    key: 'first',
    album: 'First Album',
    artist: 'The Band',
    year: '2000',
    track_count: 2,
};

describe('ArtistPage album pagination on background revalidation failure', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (getMusicArtistSummary as jest.Mock).mockResolvedValue({
            key: 'band',
            name: 'The Band',
            track_count: 2,
            album_count: 2,
            total_length_seconds: 100,
        });
        (getMusicByArtist as jest.Mock).mockResolvedValue(
            buildServerPage([], { page: 1, pageSize: 50, hasNext: false })
        );
    });

    it('keeps the albums and the load more button when a refetch fails', async () => {
        albumsService.mockResolvedValueOnce(
            buildServerPage([firstAlbum], { page: 1, pageSize: 50, hasNext: true })
        );
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        render(
            <QueryClientProvider client={queryClient}>
                <MemoryRouter>
                    <ArtistPage artistKey="band" onBack={jest.fn()} />
                </MemoryRouter>
            </QueryClientProvider>
        );
        await screen.findByText('First Album');

        albumsService.mockRejectedValue(new Error('network down'));
        await queryClient.refetchQueries({ queryKey: ['music-albums-by-artist', 'band'] });

        await waitFor(() =>
            expect(queryClient.getQueryState(['music-albums-by-artist', 'band'])?.status).toBe(
                'error'
            )
        );
        expect(screen.getByText('First Album')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'LOAD_MORE' })).toBeInTheDocument();
        expect(screen.queryByText('MUSIC_LIST_ERROR_TITLE')).not.toBeInTheDocument();
    });
});

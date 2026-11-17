import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import { apiBase } from '@/service';
import AllTracksView from './AllTracksView';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), put: jest.fn() },
}));

const mockedReplaceQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useOptionalGlobalMusic: () => undefined,
    useGlobalMusic: () => ({
        replaceQueue: mockedReplaceQueue,
        currentTrack: null,
        isPlaying: false,
    }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const batchUrl = '/music/tracks/metadata';

const paginationOf = (items: unknown[]) => ({
    items,
    pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
});

const trackNamed = (id: number) => ({
    id,
    name: `track-${id}.mp3`,
    path: `/music/track-${id}.mp3`,
    metadata: { title: `Track ${id}`, artist: 'Artist' },
});

const renderAllTracks = () =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <SnackbarProvider>
                <MemoryRouter>
                    <AllTracksView />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );

describe('AllTracksView track selection (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) =>
            Promise.resolve({
                data: paginationOf(
                    url === '/music/library' ? [trackNamed(11), trackNamed(12), trackNamed(13)] : []
                ),
            })
        );
        mockedApi.put.mockResolvedValue({ data: { matched: 2, updated: 2, failed: [] } });
    });

    it('selects two tracks, simulates and applies the batch by file ids, then leaves the mode', async () => {
        renderAllTracks();
        await screen.findByRole('button', { name: 'play Track 11' });

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECT_TRACKS' }));
        fireEvent.click(screen.getByRole('button', { name: 'play Track 11' }));
        fireEvent.click(screen.getByRole('button', { name: 'play Track 13' }));
        expect(mockedReplaceQueue).not.toHaveBeenCalled();
        expect(screen.getByText('MUSIC_SELECTION_COUNT')).toBeVisible();

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECTION_EDIT' }));
        const genresInput = screen.getByLabelText('MUSIC_METADATA_FIELD_GENRES');
        fireEvent.change(genresInput, { target: { value: 'Rock' } });
        fireEvent.keyDown(genresInput, { key: 'Enter' });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_REVIEW' }));

        await screen.findByText('MUSIC_BATCH_METADATA_WILL_CHANGE');
        expect(mockedApi.put).toHaveBeenCalledTimes(1);
        expect(mockedApi.put).toHaveBeenLastCalledWith(
            batchUrl,
            { fields: { genres: ['Rock'] }, file_ids: [11, 13] },
            { params: { dry_run: true } }
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_CONFIRM' }));

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(2));
        expect(mockedApi.put).toHaveBeenLastCalledWith(batchUrl, {
            fields: { genres: ['Rock'] },
            file_ids: [11, 13],
        });
        await waitFor(() => expect(screen.queryByRole('toolbar')).not.toBeInTheDocument());
        expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'MUSIC_SELECT_TRACKS' })).toBeVisible();
    });
});

import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import I18nProvider from '@/components/i18n/provider';
import * as musicService from '@/service/music';
import AllTracksView from './AllTracksView';

jest.mock('@/service/music');
jest.mock('@/service/playlist');

const mockReplaceQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({
        replaceQueue: mockReplaceQueue,
        currentTrack: null,
        isPlaying: false,
    }),
}));

const getMusicMock = musicService.getMusic as jest.Mock;

const pageOf = <ItemType,>(items: ItemType[], page: number, hasNext: boolean) => ({
    items,
    pagination: { page, page_size: 50, has_next: hasNext, has_prev: page > 1 },
});

const trackNamed = (id: number) => ({
    id,
    name: `track-${id}.mp3`,
    path: `/music/track-${id}.mp3`,
    metadata: { title: `Track ${id}`, artist: 'Artist' },
});

const renderTracksRoute = () =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <I18nProvider>
                <MemoryRouter initialEntries={['/music/tracks']}>
                    <Routes>
                        <Route path="/music/tracks" element={<AllTracksView />} />
                    </Routes>
                </MemoryRouter>
            </I18nProvider>
        </QueryClientProvider>
    );

describe('AllTracksView', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows the empty state when the library has no tracks', async () => {
        getMusicMock.mockResolvedValue(pageOf([], 1, false));

        renderTracksRoute();

        expect(
            await screen.findByText(/MUSIC_COLLECTION_TRACKS_EMPTY|Nenhuma|No /)
        ).toBeInTheDocument();
    });

    it('shows the error state and retries', async () => {
        getMusicMock.mockRejectedValueOnce(new Error('offline'));
        getMusicMock.mockResolvedValueOnce(pageOf([trackNamed(1)], 1, false));

        renderTracksRoute();

        expect(await screen.findByText('MUSIC_TRACKS_ERROR_TITLE')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /TRY_AGAIN|Tentar novamente/i }));
        expect(await screen.findByText('Track 1')).toBeInTheDocument();
    });

    it('loads the next page through the sentinel', async () => {
        getMusicMock.mockImplementation(async (page: number) =>
            page === 1 ? pageOf([trackNamed(1)], 1, true) : pageOf([trackNamed(2)], 2, false)
        );

        renderTracksRoute();

        expect(await screen.findByText('Track 1')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /LOAD_MORE|Carregar mais/i }));
        expect(await screen.findByText('Track 2')).toBeInTheDocument();
        expect(getMusicMock).toHaveBeenCalledWith(2, 50);
    });

    it('queues the loaded tracks from the clicked track, in order or shuffled', async () => {
        getMusicMock.mockResolvedValue(pageOf([trackNamed(1), trackNamed(2)], 1, false));

        renderTracksRoute();

        fireEvent.click(await screen.findByText('Track 2'));
        expect(mockReplaceQueue).toHaveBeenLastCalledWith(
            [expect.objectContaining({ id: 1 }), expect.objectContaining({ id: 2 })],
            1,
            expect.objectContaining({ kind: 'all-tracks', href: '/music' })
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAY_LOADED' }));
        expect(mockReplaceQueue).toHaveBeenLastCalledWith(
            expect.arrayContaining([expect.objectContaining({ id: 1 })]),
            0,
            expect.any(Object)
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SHUFFLE_LOADED' }));
        await waitFor(() => expect(mockReplaceQueue).toHaveBeenCalledTimes(3));
        expect(mockReplaceQueue.mock.calls[2][0]).toHaveLength(2);
    });
});

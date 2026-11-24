import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { SnackbarProvider } from 'notistack';
import VideoSeriesDetailView from './VideoSeriesDetailView';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            key === 'VIDEO_SERIES_SEASON_TOTAL'
                ? `${params?.count} episodios`
                : key === 'VIDEO_DETAIL_SEASON_LABEL'
                  ? `Temporada ${params?.season}`
                  : key,
    }),
}));

import { apiBase } from '@/service';

const mockedGet = apiBase.get as jest.Mock;
const mockedPut = apiBase.put as jest.Mock;

const buildEpisode = (episode: number) => ({
    id: episode,
    order_index: episode,
    source_kind: 'auto',
    status: 'not_started',
    progress_pct: 0,
    season: 1,
    episode,
    video: {
        id: 100 + episode,
        name: `Show S01E0${episode}.mkv`,
        format: 'mkv',
        path: '',
        parent_path: '',
        size: 1,
    },
});

let watchedCount: number;

const serve = () =>
    mockedGet.mockImplementation(async (url: string) => {
        if (url === '/video/playlists/5') {
            return {
                data: {
                    id: 5,
                    name: 'Show',
                    classification: 'series',
                    item_count: 2,
                    seasons: [{ season: 1, episode_count: 2, watched_count: watchedCount }],
                    resume: null,
                },
            };
        }
        return {
            data: {
                items: [buildEpisode(1), buildEpisode(2)],
                pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
            },
        };
    });

const renderView = (onHide = jest.fn()) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <SnackbarProvider>
                <VideoSeriesDetailView
                    playlist={{ id: 5, name: 'Show', item_count: 2 }}
                    onBack={jest.fn()}
                    onOpenVideo={jest.fn()}
                    onHide={onHide}
                />
            </SnackbarProvider>
        </QueryClientProvider>
    );

const openEpisodeMenu = async (itemName: string) => {
    const menuButtons = await screen.findAllByRole('button', {
        name: 'VIDEO_CORRECTION_MENU',
    });
    fireEvent.click(menuButtons[0] as HTMLElement);
    fireEvent.click(await screen.findByRole('menuitem', { name: itemName }));
    return screen.findByRole('dialog');
};

describe('VideoSeriesDetailView actions seam', () => {
    beforeEach(() => {
        mockedGet.mockReset();
        mockedPut.mockReset();
        mockedPut.mockResolvedValue({ data: { updated: 2 } });
        watchedCount = 0;
        serve();
    });

    it('marks the selected season as watched with the exact body and refetches the counts', async () => {
        renderView();
        const seasonButton = await screen.findByRole('button', {
            name: 'VIDEO_SERIES_MARK_SEASON_WATCHED',
        });
        const headerCallsBefore = mockedGet.mock.calls.filter(
            ([url]) => url === '/video/playlists/5'
        ).length;
        watchedCount = 2;

        fireEvent.click(seasonButton);

        await waitFor(() =>
            expect(mockedPut).toHaveBeenCalledWith('/video/playlists/5/watched', {
                watched: true,
                season: 1,
            })
        );
        expect(
            await screen.findByRole('button', { name: 'VIDEO_SERIES_MARK_SEASON_UNWATCHED' })
        ).toBeInTheDocument();
        expect(
            mockedGet.mock.calls.filter(([url]) => url === '/video/playlists/5').length
        ).toBeGreaterThan(headerCallsBefore);
    });

    it('marks the whole series as watched only after confirming, without a season in the body', async () => {
        renderView();
        fireEvent.click(
            await screen.findByRole('button', { name: 'VIDEO_SERIES_MARK_SERIES_WATCHED' })
        );
        expect(mockedPut).not.toHaveBeenCalled();

        fireEvent.click(
            within(await screen.findByRole('dialog')).getByRole('button', {
                name: 'VIDEO_SERIES_MARK_SERIES_WATCHED',
            })
        );

        await waitFor(() =>
            expect(mockedPut).toHaveBeenCalledWith('/video/playlists/5/watched', {
                watched: true,
            })
        );
    });

    it('marks a fully watched season as unwatched', async () => {
        watchedCount = 2;
        renderView();

        fireEvent.click(
            await screen.findByRole('button', { name: 'VIDEO_SERIES_MARK_SEASON_UNWATCHED' })
        );

        await waitFor(() =>
            expect(mockedPut).toHaveBeenCalledWith('/video/playlists/5/watched', {
                watched: false,
                season: 1,
            })
        );
    });

    it('shows the server message verbatim when marking the series fails', async () => {
        mockedPut.mockRejectedValue({ response: { data: { error: 'playlist do dono' } } });
        renderView();

        fireEvent.click(
            await screen.findByRole('button', { name: 'VIDEO_SERIES_MARK_SEASON_WATCHED' })
        );

        expect(await screen.findByText('playlist do dono')).toBeInTheDocument();
    });

    it('asks for confirmation before hiding the series', async () => {
        const onHide = jest.fn();
        renderView(onHide);

        fireEvent.click(await screen.findByRole('button', { name: 'VIDEO_HIDE_SERIES' }));
        expect(onHide).not.toHaveBeenCalled();
        fireEvent.click(
            within(await screen.findByRole('dialog')).getByRole('button', {
                name: 'VIDEO_HIDE_SERIES',
            })
        );

        expect(onHide).toHaveBeenCalledTimes(1);
    });

    it('shows the season episode total at the end of the list', async () => {
        renderView();

        expect(await screen.findByText('2 episodios')).toBeInTheDocument();
    });

    it('PUTs the chosen classification for the episode file', async () => {
        renderView();
        const dialog = await openEpisodeMenu('VIDEO_CORRECT_CLASSIFICATION');

        fireEvent.click(within(dialog).getByRole('radio', { name: 'VIDEO_CLASS_MOVIE' }));
        fireEvent.click(within(dialog).getByRole('button', { name: 'VIDEO_CORRECTION_SAVE' }));

        await waitFor(() =>
            expect(mockedPut).toHaveBeenCalledWith('/video/library/files/101/classification', {
                classification: 'movie',
            })
        );
        await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    });

    it('PUTs a null classification for the automatic option', async () => {
        renderView();
        const dialog = await openEpisodeMenu('VIDEO_CORRECT_CLASSIFICATION');

        fireEvent.click(within(dialog).getByRole('button', { name: 'VIDEO_CORRECTION_SAVE' }));

        await waitFor(() =>
            expect(mockedPut).toHaveBeenCalledWith('/video/library/files/101/classification', {
                classification: null,
            })
        );
    });

    it('PUTs the corrected season and episode and refetches the episodes', async () => {
        renderView();
        const dialog = await openEpisodeMenu('VIDEO_CORRECT_EPISODE');
        const episodeCallsBefore = mockedGet.mock.calls.filter(
            ([url]) => url === '/video/playlists/5/items'
        ).length;

        fireEvent.change(within(dialog).getByLabelText('VIDEO_CORRECTION_SEASON_LABEL'), {
            target: { value: '2' },
        });
        fireEvent.change(within(dialog).getByLabelText('VIDEO_CORRECTION_EPISODE_LABEL'), {
            target: { value: '7' },
        });
        fireEvent.click(within(dialog).getByRole('button', { name: 'VIDEO_CORRECTION_SAVE' }));

        await waitFor(() =>
            expect(mockedPut).toHaveBeenCalledWith('/video/library/files/101/episode', {
                season: 2,
                episode: 7,
            })
        );
        await waitFor(() =>
            expect(
                mockedGet.mock.calls.filter(([url]) => url === '/video/playlists/5/items').length
            ).toBeGreaterThan(episodeCallsBefore)
        );
    });

    it('PUTs nulls for the automatic episode', async () => {
        renderView();
        const dialog = await openEpisodeMenu('VIDEO_CORRECT_EPISODE');

        fireEvent.click(within(dialog).getByRole('button', { name: 'VIDEO_CORRECTION_AUTOMATIC' }));

        await waitFor(() =>
            expect(mockedPut).toHaveBeenCalledWith('/video/library/files/101/episode', {
                season: null,
                episode: null,
            })
        );
    });

    it('shows the 400 message verbatim and marks the offending field', async () => {
        mockedPut.mockRejectedValue({
            response: { data: { error: 'episodio invalido', field: 'episode' } },
        });
        renderView();
        const dialog = await openEpisodeMenu('VIDEO_CORRECT_EPISODE');

        fireEvent.change(within(dialog).getByLabelText('VIDEO_CORRECTION_EPISODE_LABEL'), {
            target: { value: '0' },
        });
        fireEvent.click(within(dialog).getByRole('button', { name: 'VIDEO_CORRECTION_SAVE' }));

        expect(await within(dialog).findByText('episodio invalido')).toBeInTheDocument();
        expect(within(dialog).getByLabelText('VIDEO_CORRECTION_EPISODE_LABEL')).toHaveAttribute(
            'aria-invalid',
            'true'
        );
        expect(within(dialog).getByLabelText('VIDEO_CORRECTION_SEASON_LABEL')).toHaveAttribute(
            'aria-invalid',
            'false'
        );
    });
});

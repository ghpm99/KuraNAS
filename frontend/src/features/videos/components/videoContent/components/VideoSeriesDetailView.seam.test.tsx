import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { SnackbarProvider } from 'notistack';
import VideoSeriesDetailView from './VideoSeriesDetailView';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            key === 'VIDEO_SERIES_CONTINUE'
                ? `Continuar T${params?.season}E${params?.episode}`
                : key === 'VIDEO_DETAIL_SEASON_LABEL'
                  ? `Temporada ${params?.season}`
                  : key === 'VIDEO_SERIES_EPISODE_NUMBER'
                    ? `E${params?.episode}`
                    : key,
    }),
}));

import { apiBase } from '@/service';

const mockedGet = apiBase.get as jest.Mock;

const buildHeader = (mode: string, resumeSeason: number | null) => ({
    id: 5,
    name: 'Show',
    classification: 'series',
    item_count: 4,
    seasons: [
        { season: 1, episode_count: 2, watched_count: 2 },
        { season: 2, episode_count: 2, watched_count: 0 },
    ],
    resume: {
        video_id: 62,
        season: resumeSeason,
        episode: 1,
        position_seconds: 0,
        mode,
    },
});

const buildEpisode = (season: number, episode: number) => ({
    id: season * 10 + episode,
    order_index: season * 10 + episode,
    source_kind: 'auto',
    status: 'not_started',
    progress_pct: 0,
    season,
    episode,
    video: {
        id: season * 100 + episode,
        name: `Show S0${season}E0${episode}.mkv`,
        format: 'mkv',
        path: '',
        parent_path: '',
        size: 1,
    },
});

const serve = (header: object, failSecondPage = false) =>
    mockedGet.mockImplementation(
        async (url: string, config?: { params: Record<string, number> }) => {
            if (url === '/video/playlists/5') return { data: header };
            const season = config?.params.season ?? 1;
            const page = config?.params.page ?? 1;
            if (failSecondPage && page === 2) throw new Error('page 2 down');
            return {
                data: {
                    items: [buildEpisode(season, page)],
                    pagination: { page, page_size: 50, has_next: page < 2, has_prev: page > 1 },
                },
            };
        }
    );

const renderView = (onOpenVideo = jest.fn()) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <SnackbarProvider>
                <VideoSeriesDetailView
                    playlist={{ id: 5, name: 'Show', item_count: 4 }}
                    onBack={jest.fn()}
                    onOpenVideo={onOpenVideo}
                />
            </SnackbarProvider>
        </QueryClientProvider>
    );

describe('VideoSeriesDetailView seam', () => {
    beforeEach(() => {
        mockedGet.mockReset();
    });

    it.each([
        ['continue', 'Continuar T2E1'],
        ['start', 'VIDEO_SERIES_WATCH'],
        ['rewatch', 'VIDEO_SERIES_REWATCH'],
    ])('labels the main button for the %s mode and opens the resume video', async (mode, label) => {
        serve(buildHeader(mode, 2));
        const onOpenVideo = jest.fn();
        renderView(onOpenVideo);

        const button = await screen.findByRole('button', { name: new RegExp(label) });
        await waitFor(() => expect(button).toBeEnabled());
        fireEvent.click(button);

        expect(onOpenVideo).toHaveBeenCalledWith(62);
    });

    it('shows the season of the resume target, the progress and switches season', async () => {
        serve(buildHeader('continue', 2));
        renderView();

        expect(await screen.findByText('E1')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Temporada 2' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
        expect(screen.getByText('VIDEO_DETAIL_SERIES_PROGRESS')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'Temporada 1' }));

        await waitFor(() =>
            expect(mockedGet).toHaveBeenCalledWith('/video/playlists/5/items', {
                params: { page: 1, page_size: 50, season: 1 },
            })
        );
        expect(screen.getByRole('button', { name: 'Temporada 1' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
    });

    it('keeps the first page and offers a retry when the second page fails', async () => {
        serve(buildHeader('start', 1), true);
        renderView();

        await screen.findByText('E1');
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.getByText('E1')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'TRY_AGAIN' })).toBeInTheDocument();
    });
});

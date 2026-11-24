import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { useVideoSeriesDetail } from './useVideoSeriesDetail';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

import { apiBase } from '@/service';

const mockedGet = apiBase.get as jest.Mock;

const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        {children}
    </QueryClientProvider>
);

const header = {
    id: 5,
    name: 'Show',
    classification: 'series',
    item_count: 6,
    seasons: [
        { season: 1, episode_count: 2, watched_count: 2 },
        { season: 2, episode_count: 4, watched_count: 1 },
    ],
    resume: { video_id: 62, season: 2, episode: 2, position_seconds: 30, mode: 'continue' },
};

const episodePage = (page: number, hasNext: boolean) => ({
    data: {
        items: [{ id: page * 10, season: 2, episode: page, video: { id: page * 100 } }],
        pagination: { page, page_size: 50, has_next: hasNext, has_prev: page > 1 },
    },
});

describe('useVideoSeriesDetail seam', () => {
    beforeEach(() => {
        mockedGet.mockReset();
    });

    it('requests the header without items, defaults to the resume season and sums progress', async () => {
        mockedGet.mockImplementation(async (url: string) =>
            url === '/video/playlists/5' ? { data: header } : episodePage(1, false)
        );

        const { result } = renderHook(() => useVideoSeriesDetail(5), { wrapper });

        await waitFor(() => expect(result.current.episodes).toHaveLength(1));
        expect(mockedGet).toHaveBeenCalledWith('/video/playlists/5', {
            params: { include_items: false },
        });
        expect(mockedGet).toHaveBeenCalledWith('/video/playlists/5/items', {
            params: { page: 1, page_size: 50, season: 2 },
        });
        expect(result.current.selectedSeason).toBe(2);
        expect(result.current.totalEpisodes).toBe(6);
        expect(result.current.watchedEpisodes).toBe(3);
    });

    it('falls back to the first season when there is no resume target', async () => {
        mockedGet.mockImplementation(async (url: string) =>
            url === '/video/playlists/5'
                ? { data: { ...header, resume: null } }
                : episodePage(1, false)
        );

        const { result } = renderHook(() => useVideoSeriesDetail(5), { wrapper });

        await waitFor(() => expect(result.current.episodes).toHaveLength(1));
        expect(result.current.selectedSeason).toBe(1);
        expect(mockedGet).toHaveBeenCalledWith('/video/playlists/5/items', {
            params: { page: 1, page_size: 50, season: 1 },
        });
    });

    it('survives a header without seasons and loads the items without a season filter', async () => {
        mockedGet.mockImplementation(async (url: string) =>
            url === '/video/playlists/5'
                ? { data: { id: 5, name: 'Show', item_count: 1 } }
                : episodePage(1, false)
        );

        const { result } = renderHook(() => useVideoSeriesDetail(5), { wrapper });

        await waitFor(() => expect(result.current.episodes).toHaveLength(1));
        expect(result.current.selectedSeason).toBeUndefined();
        expect(mockedGet).toHaveBeenCalledWith('/video/playlists/5/items', {
            params: { page: 1, page_size: 50 },
        });
    });

    it('follows has_next across three pages of the selected season', async () => {
        mockedGet.mockImplementation(async (url: string, config?: { params: { page: number } }) => {
            if (url === '/video/playlists/5') return { data: header };
            const page = config?.params.page ?? 1;
            return episodePage(page, page < 3);
        });

        const { result } = renderHook(() => useVideoSeriesDetail(5), { wrapper });

        await waitFor(() => expect(result.current.episodes).toHaveLength(1));
        expect(result.current.episodesQuery.hasNextPage).toBe(true);

        await act(async () => {
            await result.current.episodesQuery.fetchNextPage();
        });
        await waitFor(() => expect(result.current.episodes).toHaveLength(2));

        await act(async () => {
            await result.current.episodesQuery.fetchNextPage();
        });
        await waitFor(() => expect(result.current.episodes).toHaveLength(3));

        expect(result.current.episodes.map((episode) => episode.episode)).toEqual([1, 2, 3]);
        expect(result.current.episodesQuery.hasNextPage).toBe(false);
        expect(mockedGet).toHaveBeenCalledWith('/video/playlists/5/items', {
            params: { page: 3, page_size: 50, season: 2 },
        });
    });

    it('keeps the loaded page when the second page fails', async () => {
        mockedGet.mockImplementation(async (url: string, config?: { params: { page: number } }) => {
            if (url === '/video/playlists/5') return { data: header };
            if (config?.params.page === 2) throw new Error('page 2 down');
            return episodePage(1, true);
        });

        const { result } = renderHook(() => useVideoSeriesDetail(5), { wrapper });

        await waitFor(() => expect(result.current.episodes).toHaveLength(1));
        act(() => {
            void result.current.episodesQuery.fetchNextPage();
        });

        await waitFor(() => expect(result.current.episodesQuery.isFetchNextPageError).toBe(true));
        expect(result.current.episodes).toHaveLength(1);
    });

    it('drops an episode repeated by a later page', async () => {
        mockedGet.mockImplementation(async (url: string, config?: { params: { page: number } }) => {
            if (url === '/video/playlists/5') return { data: header };
            const page = config?.params.page ?? 1;
            return page === 1
                ? { data: { ...episodePage(1, true).data } }
                : {
                      data: {
                          ...episodePage(1, false).data,
                          pagination: { page: 2, page_size: 50, has_next: false, has_prev: true },
                      },
                  };
        });

        const { result } = renderHook(() => useVideoSeriesDetail(5), { wrapper });

        await waitFor(() => expect(result.current.episodes).toHaveLength(1));
        await act(async () => {
            await result.current.episodesQuery.fetchNextPage();
        });

        await waitFor(() => expect(result.current.episodesQuery.data?.pages).toHaveLength(2));
        expect(result.current.episodes).toHaveLength(1);
    });

    it('switches season and requests that season from the server', async () => {
        mockedGet.mockImplementation(async (url: string) =>
            url === '/video/playlists/5' ? { data: header } : episodePage(1, false)
        );

        const { result } = renderHook(() => useVideoSeriesDetail(5), { wrapper });
        await waitFor(() => expect(result.current.episodes).toHaveLength(1));

        act(() => result.current.selectSeason(1));

        await waitFor(() => expect(result.current.selectedSeason).toBe(1));
        await waitFor(() =>
            expect(mockedGet).toHaveBeenCalledWith('/video/playlists/5/items', {
                params: { page: 1, page_size: 50, season: 1 },
            })
        );
    });
});

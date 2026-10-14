import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { useHiddenVideoPlaylists } from './useVideoQueries';

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

describe('useHiddenVideoPlaylists seam', () => {
    beforeEach(() => {
        mockedGet.mockReset();
    });

    it('requests the hidden route with page and page_size and follows has_next', async () => {
        mockedGet
            .mockResolvedValueOnce({
                data: {
                    items: [{ id: 1 }],
                    pagination: { page: 1, page_size: 24, has_next: true, has_prev: false },
                },
            })
            .mockResolvedValueOnce({
                data: {
                    items: [{ id: 2 }],
                    pagination: { page: 2, page_size: 24, has_next: false, has_prev: true },
                },
            });

        const { result } = renderHook(() => useHiddenVideoPlaylists(24, true), { wrapper });

        await waitFor(() => expect(result.current.data?.pages).toHaveLength(1));
        expect(mockedGet).toHaveBeenNthCalledWith(1, '/video/playlists/hidden', {
            params: { page: 1, page_size: 24 },
        });
        expect(result.current.hasNextPage).toBe(true);

        await result.current.fetchNextPage();

        await waitFor(() => expect(result.current.data?.pages).toHaveLength(2));
        expect(mockedGet).toHaveBeenNthCalledWith(2, '/video/playlists/hidden', {
            params: { page: 2, page_size: 24 },
        });
        expect(result.current.hasNextPage).toBe(false);
    });

    it('does not call the backend while disabled', () => {
        renderHook(() => useHiddenVideoPlaylists(24, false), { wrapper });

        expect(mockedGet).not.toHaveBeenCalled();
    });
});

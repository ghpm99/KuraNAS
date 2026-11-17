import { ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, useInfiniteQuery } from '@tanstack/react-query';
import useRetryFailedPage from './useRetryFailedPage';

const createWrapper = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
};

const useRetryableQuery = (fetchPage: (pageNumber: number) => Promise<{ nextPage?: number }>) => {
    const query = useInfiniteQuery({
        queryKey: ['retry-failed-page'],
        queryFn: ({ pageParam }) => fetchPage(pageParam),
        initialPageParam: 1,
        getNextPageParam: (lastPage) => lastPage.nextPage,
    });
    const retryFailedPage = useRetryFailedPage(query);
    return { query, retryFailedPage };
};

describe('useRetryFailedPage', () => {
    it('refaz a primeira página quando ela falhou', async () => {
        const requestedPages: number[] = [];
        const fetchPage = async (pageNumber: number) => {
            requestedPages.push(pageNumber);
            throw new Error('falha');
        };

        const { result } = renderHook(() => useRetryableQuery(fetchPage), {
            wrapper: createWrapper(),
        });
        await waitFor(() => expect(result.current.query.isError).toBe(true));
        expect(result.current.query.isFetchNextPageError).toBe(false);

        act(() => result.current.retryFailedPage());

        await waitFor(() => expect(requestedPages).toEqual([1, 1]));
    });

    it('pede de novo a página que falhou ao carregar a próxima', async () => {
        const requestedPages: number[] = [];
        const fetchPage = async (pageNumber: number) => {
            requestedPages.push(pageNumber);
            if (pageNumber === 1) return { nextPage: 2 };
            throw new Error('falha');
        };

        const { result } = renderHook(() => useRetryableQuery(fetchPage), {
            wrapper: createWrapper(),
        });
        await waitFor(() => expect(result.current.query.isSuccess).toBe(true));
        act(() => void result.current.query.fetchNextPage());
        await waitFor(() => expect(result.current.query.isFetchNextPageError).toBe(true));

        act(() => result.current.retryFailedPage());

        await waitFor(() => expect(requestedPages).toEqual([1, 2, 2]));
    });
});

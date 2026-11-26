import { ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, useInfiniteQuery } from '@tanstack/react-query';
import type { InfiniteData } from '@tanstack/react-query';
import useReloadInfiniteListFromStart from './useReloadInfiniteListFromStart';

type TestPage = { label: string; nextPage?: number };

const listQueryKey = ['reload-from-start'];

const setup = (fetchPage: (pageNumber: number) => Promise<TestPage>) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    queryClient.setQueryData<InfiniteData<TestPage, number>>(listQueryKey, {
        pages: [{ label: 'old-1', nextPage: 2 }, { label: 'old-2' }],
        pageParams: [1, 2],
    });
    const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const rendered = renderHook(
        () => {
            const query = useInfiniteQuery({
                queryKey: listQueryKey,
                queryFn: ({ pageParam }) => fetchPage(pageParam),
                initialPageParam: 1,
                getNextPageParam: (lastPage) => lastPage.nextPage,
                staleTime: Infinity,
            });
            const reloadFromStart = useReloadInfiniteListFromStart(listQueryKey, query);
            return { query, observedPages: query.data?.pages, reloadFromStart };
        },
        { wrapper }
    );
    return { queryClient, ...rendered };
};

const cachedLabels = (queryClient: QueryClient) =>
    queryClient
        .getQueryData<InfiniteData<TestPage>>(listQueryKey)
        ?.pages.map((cachedPage) => cachedPage.label);

describe('useReloadInfiniteListFromStart', () => {
    it('keeps only the freshly fetched first page and requests page 1 only', async () => {
        const requestedPages: number[] = [];
        const { queryClient, result } = setup(async (pageNumber) => {
            requestedPages.push(pageNumber);
            return { label: 'new-1' };
        });

        let outcome: Awaited<ReturnType<typeof result.current.reloadFromStart>> | undefined;
        await act(async () => {
            outcome = await result.current.reloadFromStart();
        });

        expect(requestedPages).toEqual([1]);
        expect(cachedLabels(queryClient)).toEqual(['new-1']);
        expect(outcome?.isReloadFailed).toBe(false);
    });

    it('keeps the current items available while the response has not arrived', async () => {
        let resolveFirstPage: (page: TestPage) => void = () => undefined;
        const { queryClient, result } = setup(
            () => new Promise<TestPage>((resolve) => (resolveFirstPage = resolve))
        );

        let reloadPromise: Promise<unknown> = Promise.resolve();
        act(() => {
            reloadPromise = result.current.reloadFromStart();
        });

        await waitFor(() => expect(queryClient.isFetching({ queryKey: listQueryKey })).toBe(1));
        expect(cachedLabels(queryClient)).toEqual(['old-1']);
        await waitFor(() => expect(result.current.observedPages).toHaveLength(1));

        await act(async () => {
            resolveFirstPage({ label: 'new-1' });
            await reloadPromise;
        });
        expect(cachedLabels(queryClient)).toEqual(['new-1']);
    });

    it('reports the failure and keeps the old first page in cache', async () => {
        const failure = new Error('falha');
        const { queryClient, result } = setup(async () => {
            throw failure;
        });

        let outcome: Awaited<ReturnType<typeof result.current.reloadFromStart>> | undefined;
        await act(async () => {
            outcome = await result.current.reloadFromStart();
        });

        expect(outcome).toEqual({ isReloadFailed: true, reloadError: failure });
        expect(cachedLabels(queryClient)).toEqual(['old-1']);
    });
});

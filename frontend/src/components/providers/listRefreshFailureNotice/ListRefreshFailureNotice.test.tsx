import { act, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, useInfiniteQuery } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import ListRefreshFailureNotice from './ListRefreshFailureNotice';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

type ListQueryHandle = ReturnType<typeof useInfiniteQuery<number, Error>>;

const ListProbe = ({
    loadPage,
    onReady,
}: {
    loadPage: (pageParam: number) => Promise<number>;
    onReady: (handle: ListQueryHandle) => void;
}) => {
    const listQuery = useInfiniteQuery<number, Error>({
        queryKey: ['probe-list'],
        queryFn: ({ pageParam }) => loadPage(pageParam as number),
        initialPageParam: 0,
        getNextPageParam: (lastPage) => lastPage + 1,
        retry: false,
    });
    onReady({ ...listQuery });
    return null;
};

const renderNotice = (loadPage: (pageParam: number) => Promise<number>) => {
    const queryClient = new QueryClient();
    let latestHandle: ListQueryHandle | undefined;
    render(
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>
                <ListRefreshFailureNotice />
                <ListProbe
                    loadPage={loadPage}
                    onReady={(handle) => {
                        latestHandle = handle;
                    }}
                />
            </SnackbarProvider>
        </QueryClientProvider>
    );
    return () => latestHandle as ListQueryHandle;
};

const noticeText = 'LIST_REFRESH_FAILED';

describe('ListRefreshFailureNotice', () => {
    it('renders without a notice when nothing fails', () => {
        render(
            <QueryClientProvider client={new QueryClient()}>
                <SnackbarProvider>
                    <ListRefreshFailureNotice />
                </SnackbarProvider>
            </QueryClientProvider>
        );

        expect(screen.queryByText(noticeText)).not.toBeInTheDocument();
    });

    it('warns once when a refetch fails after the list had data', async () => {
        let shouldFail = false;
        const getHandle = renderNotice(async (pageParam) => {
            if (shouldFail) throw new Error('boom');
            return pageParam;
        });
        await waitFor(() => expect(getHandle().data).toBeDefined());

        shouldFail = true;
        await act(async () => {
            await getHandle().refetch();
        });

        await waitFor(() => expect(screen.getAllByText(noticeText)).toHaveLength(1));
    });

    it('stays silent when fetching the next page fails', async () => {
        const getHandle = renderNotice(async (pageParam) => {
            if (pageParam > 0) throw new Error('boom');
            return pageParam;
        });
        await waitFor(() => expect(getHandle().data).toBeDefined());

        await act(async () => {
            void getHandle().fetchNextPage();
        });

        await waitFor(() => expect(getHandle().isFetchNextPageError).toBe(true));
        expect(screen.queryByText(noticeText)).not.toBeInTheDocument();
    });

    it('stays silent when the first load fails', async () => {
        const getHandle = renderNotice(async () => {
            throw new Error('boom');
        });

        await waitFor(() => expect(getHandle().isError).toBe(true));
        expect(screen.queryByText(noticeText)).not.toBeInTheDocument();
    });

    it('shows a single notice for consecutive refetch failures', async () => {
        let shouldFail = false;
        const getHandle = renderNotice(async (pageParam) => {
            if (shouldFail) throw new Error('boom');
            return pageParam;
        });
        await waitFor(() => expect(getHandle().data).toBeDefined());

        shouldFail = true;
        await act(async () => {
            await getHandle().refetch();
        });
        await act(async () => {
            await getHandle().refetch();
        });

        await waitFor(() => expect(screen.getAllByText(noticeText)).toHaveLength(1));
    });
});

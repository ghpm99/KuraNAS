import { useMemo } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { Pagination } from '@/types/pagination';

export function useMusicInfinitePages<ItemType>(
    queryKey: readonly unknown[],
    fetchPage: (pageNumber: number) => Promise<Pagination<ItemType>>,
    isEnabled = true
) {
    const query = useInfiniteQuery({
        queryKey,
        queryFn: ({ pageParam = 1 }) => fetchPage(pageParam),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
        enabled: isEnabled,
    });
    const { data, isFetchNextPageError, fetchNextPage, refetch } = query;
    const items = useMemo(() => data?.pages.flatMap((page) => page.items ?? []) ?? [], [data]);

    const retry = () => {
        if (isFetchNextPageError) {
            void fetchNextPage();
            return;
        }
        void refetch();
    };

    return {
        items,
        isLoading: query.isLoading,
        isError: query.isError,
        errorMessage: query.isError ? extractBackendErrorMessage(query.error) : undefined,
        retry,
        fetchNextPage: () => void fetchNextPage(),
        hasNextPage: Boolean(query.hasNextPage),
        isFullyLoaded: query.isSuccess && !query.hasNextPage,
        isFetchingNextPage: query.isFetchingNextPage,
    };
}

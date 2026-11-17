import { useInfiniteQuery } from '@tanstack/react-query';
import { analyticsStaleTimeMs } from '@/components/providers/queryFreshness';
import { getDuplicateGroupsPage } from '@/service/analytics';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { DuplicateCategory, DuplicateGroupDetail } from '@/types/analytics';
import { getNextPageNumber } from '@/components/analytics/getNextPageNumber';

const duplicateGroupsPageSize = 50;

export const duplicateGroupsQueryKey = ['analytics', 'duplicate-groups-page'] as const;

export const useDuplicateGroupsPage = (category: DuplicateCategory | undefined) => {
    const query = useInfiniteQuery({
        queryKey: [...duplicateGroupsQueryKey, category ?? 'all'],
        queryFn: ({ pageParam }) =>
            getDuplicateGroupsPage({
                page: pageParam,
                pageSize: duplicateGroupsPageSize,
                type: category,
            }),
        initialPageParam: 1,
        getNextPageParam: getNextPageNumber,
        retry: false,
        staleTime: analyticsStaleTimeMs,
    });

    const retry = useRetryFailedPage(query);
    const groups: DuplicateGroupDetail[] =
        query.data?.pages.flatMap((page) => page?.items ?? []) ?? [];

    return {
        groups,
        isPending: query.isPending,
        isError: query.isError,
        isFetchNextPageError: query.isFetchNextPageError,
        isFetchingNextPage: query.isFetchingNextPage,
        hasNextPage: query.hasNextPage,
        fetchNextPage: () => void query.fetchNextPage(),
        errorMessage: extractBackendErrorMessage(query.error),
        retry,
    };
};

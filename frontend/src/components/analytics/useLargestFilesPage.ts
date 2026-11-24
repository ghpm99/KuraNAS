import { useInfiniteQuery } from '@tanstack/react-query';
import { analyticsStaleTimeMs } from '@/components/providers/queryFreshness';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { getLargestFilesPage } from '@/service/files';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';

const largestFilesPageSize = 50;

export const largestFilesQueryKey = ['analytics', 'largest-files'] as const;

export const useLargestFilesPage = () => {
    const query = useInfiniteQuery({
        queryKey: largestFilesQueryKey,
        queryFn: ({ pageParam }) =>
            getLargestFilesPage({ page: pageParam, pageSize: largestFilesPageSize }),
        initialPageParam: 1,
        getNextPageParam: getNextPageNumber,
        retry: false,
        staleTime: analyticsStaleTimeMs,
    });

    const retry = useRetryFailedPage(query);
    const files: FileData[] = query.data?.pages.flatMap((page) => page?.items ?? []) ?? [];

    return {
        files,
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

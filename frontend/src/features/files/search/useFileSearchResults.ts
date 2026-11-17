import { useInfiniteQuery } from '@tanstack/react-query';
import { searchFiles } from '@/service/files';
import { extractBackendErrorMessage } from '@/features/files/fileActions/bulkOutcome';
import type { FileData, PaginationResponse } from '@/features/files/providers/fileProvider/fileContext';
import type { FileSearchRefinements } from '@/types/fileSearch';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';

const searchPageSize = 60;

type FileSearchScope = {
    query: string;
    parentId?: number;
    isRecursive: boolean;
    refinements?: FileSearchRefinements;
};

const useFileSearchResults = ({ query, parentId, isRecursive, refinements }: FileSearchScope) => {
    const isSearchActive = query !== '';

    const {
        data,
        status,
        error,
        refetch,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        isFetchNextPageError,
    } = useInfiniteQuery({
        queryKey: ['files-search', query, parentId ?? null, isRecursive, refinements ?? null],
        queryFn: ({ pageParam = 1 }): Promise<PaginationResponse> =>
            searchFiles({
                q: query,
                parentId,
                recursive: isRecursive,
                page: pageParam,
                pageSize: searchPageSize,
                refinements,
            }),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage?.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
        enabled: isSearchActive,
        staleTime: listingStaleTimeMs,
    });

    const retry = useRetryFailedPage({ isFetchNextPageError, fetchNextPage, refetch });

    const items: FileData[] = data?.pages.flatMap((page) => page?.items ?? []) ?? [];

    return {
        items,
        status,
        hasNextPage: Boolean(hasNextPage),
        isFetchingNextPage,
        isFetchNextPageError,
        fetchNextPage: () => {
            fetchNextPage();
        },
        errorMessage: extractBackendErrorMessage(error),
        retry,
    };
};

export default useFileSearchResults;

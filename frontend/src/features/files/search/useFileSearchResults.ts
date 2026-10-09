import { useInfiniteQuery } from '@tanstack/react-query';
import { searchFiles } from '@/service/files';
import { extractBackendErrorMessage } from '@/features/files/fileActions/bulkOutcome';
import type { FileData, PaginationResponse } from '@/features/files/providers/fileProvider/fileContext';
import type { FileSearchRefinements } from '@/types/fileSearch';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';

const searchPageSize = 100;

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
            lastPage?.pagination?.hasNext ? lastPage.pagination.page + 1 : undefined,
        enabled: isSearchActive,
        staleTime: listingStaleTimeMs,
    });

    const items: FileData[] = data?.pages.flatMap((page) => page?.items ?? []) ?? [];

    return {
        items,
        status,
        hasNextPage: Boolean(hasNextPage),
        isFetchingNextPage,
        fetchNextPage: () => {
            fetchNextPage();
        },
        errorMessage: extractBackendErrorMessage(error),
        retry: () => {
            refetch();
        },
    };
};

export default useFileSearchResults;

import { useInfiniteQuery } from '@tanstack/react-query';
import { searchFiles } from '@/service/files';
import type { FileData, PaginationResponse } from '@/features/files/providers/fileProvider/fileContext';

const searchPageSize = 100;

type FileSearchScope = {
    query: string;
    parentId?: number;
    isRecursive: boolean;
};

const useFileSearchResults = ({ query, parentId, isRecursive }: FileSearchScope) => {
    const isSearchActive = query !== '';

    const { data, status, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['files-search', query, parentId ?? null, isRecursive],
        queryFn: ({ pageParam = 1 }): Promise<PaginationResponse> =>
            searchFiles({
                q: query,
                parentId,
                recursive: isRecursive,
                page: pageParam,
                pageSize: searchPageSize,
            }),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage?.pagination?.hasNext ? lastPage.pagination.page + 1 : undefined,
        enabled: isSearchActive,
        staleTime: 0,
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
    };
};

export default useFileSearchResults;

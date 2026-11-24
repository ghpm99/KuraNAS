import { useInfiniteQuery } from '@tanstack/react-query';
import { searchDocuments } from '@/service/documents';
import type { DocumentFilters, DocumentSearchResult } from '@/service/documents';
import { extractBackendErrorMessage } from '@/features/files/fileActions/bulkOutcome';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { documentsSearchQueryKey } from './documentsQueryKeys';

export const documentsSearchPageSize = 20;

const dedupeByFileId = (results: DocumentSearchResult[]): DocumentSearchResult[] => {
    const seenFileIds = new Set<number>();
    return results.filter((result) => {
        if (seenFileIds.has(result.file_id)) return false;
        seenFileIds.add(result.file_id);
        return true;
    });
};

const useDocumentsSearch = (query: string, filters: DocumentFilters) => {
    const searchQuery = useInfiniteQuery({
        queryKey: documentsSearchQueryKey(query, filters),
        queryFn: ({ pageParam = 1, signal }) =>
            searchDocuments({
                q: query,
                filters,
                page: pageParam,
                pageSize: documentsSearchPageSize,
                signal,
            }),
        initialPageParam: 1,
        getNextPageParam: getNextPageNumber,
        enabled: query !== '',
        staleTime: listingStaleTimeMs,
    });
    const retry = useRetryFailedPage(searchQuery);

    return {
        results: dedupeByFileId(searchQuery.data?.pages.flatMap((page) => page?.items ?? []) ?? []),
        isLoading: searchQuery.isPending,
        isError: searchQuery.isError,
        isNextPageError: searchQuery.isFetchNextPageError,
        hasNextPage: Boolean(searchQuery.hasNextPage),
        isFetchingNextPage: searchQuery.isFetchingNextPage,
        fetchNextPage: () => {
            void searchQuery.fetchNextPage();
        },
        errorMessage: extractBackendErrorMessage(searchQuery.error),
        retry,
    };
};

export default useDocumentsSearch;

import { useInfiniteQuery } from '@tanstack/react-query';
import { searchDocuments } from '@/service/documents';
import type { DocumentSearchResult } from '@/service/documents';
import { extractBackendErrorMessage } from '@/features/files/fileActions/bulkOutcome';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';

const documentSearchPageSize = 20;

type DocumentSearchScope = {
    query: string;
    isEnabled: boolean;
};

const useDocumentSearchResults = ({ query, isEnabled }: DocumentSearchScope) => {
    const documentsQuery = useInfiniteQuery({
        queryKey: ['documents-search', query],
        queryFn: ({ pageParam = 1, signal }) =>
            searchDocuments({
                q: query,
                page: pageParam,
                pageSize: documentSearchPageSize,
                signal,
            }),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage?.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
        enabled: isEnabled && query !== '',
        staleTime: listingStaleTimeMs,
    });
    const { data, status, error, fetchNextPage, hasNextPage, isFetchingNextPage, isFetchNextPageError } =
        documentsQuery;
    const retry = useRetryFailedPage(documentsQuery);

    const items: DocumentSearchResult[] = data?.pages.flatMap((page) => page?.items ?? []) ?? [];

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

export default useDocumentSearchResults;

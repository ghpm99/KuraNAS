import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { countDocuments, listDocuments } from '@/service/documents';
import type { DocumentFilters, DocumentListItem, DocumentSort } from '@/service/documents';
import { extractBackendErrorMessage } from '@/features/files/fileActions/bulkOutcome';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { documentsCountQueryKey, documentsListQueryKey } from './documentsQueryKeys';

export const documentsPageSize = 50;

const noFilters: DocumentFilters = {};
const defaultSort: DocumentSort = {};

const dedupeById = (documents: DocumentListItem[]): DocumentListItem[] => {
    const seenIds = new Set<number>();
    return documents.filter((document) => {
        if (seenIds.has(document.id)) return false;
        seenIds.add(document.id);
        return true;
    });
};

const useDocumentsList = (
    filters: DocumentFilters = noFilters,
    sort: DocumentSort = defaultSort,
    isEnabled = true
) => {
    const listQuery = useInfiniteQuery({
        queryKey: documentsListQueryKey(filters, sort),
        queryFn: ({ pageParam = 1, signal }) =>
            listDocuments({ filters, sort, page: pageParam, pageSize: documentsPageSize, signal }),
        initialPageParam: 1,
        getNextPageParam: getNextPageNumber,
        enabled: isEnabled,
        staleTime: listingStaleTimeMs,
    });
    const countQuery = useQuery({
        queryKey: documentsCountQueryKey(filters),
        queryFn: () => countDocuments(filters),
        enabled: isEnabled,
        staleTime: listingStaleTimeMs,
    });
    const retry = useRetryFailedPage(listQuery);
    const { refetch: refetchList } = listQuery;
    const { refetch: refetchCount } = countQuery;

    const documents = dedupeById(listQuery.data?.pages.flatMap((page) => page?.items ?? []) ?? []);

    return {
        documents,
        total: countQuery.data?.total,
        isLoading: listQuery.isPending,
        isError: listQuery.isError,
        isNextPageError: listQuery.isFetchNextPageError,
        hasNextPage: Boolean(listQuery.hasNextPage),
        isFetchingNextPage: listQuery.isFetchingNextPage,
        fetchNextPage: () => {
            void listQuery.fetchNextPage();
        },
        errorMessage: extractBackendErrorMessage(listQuery.error),
        retry,
        reload: () => {
            void refetchList();
            void refetchCount();
        },
    };
};

export default useDocumentsList;

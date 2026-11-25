import { useInfiniteQuery } from '@tanstack/react-query';
import { getIngestDownloads } from '@/service/ingest';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { IngestDownload } from '@/types/ingest';
import { finishedIngestDownloadsQueryKey } from './ingestQueryKeys';

export const FINISHED_DOWNLOADS_PAGE_SIZE = 20;

const removeDuplicateIds = (downloads: IngestDownload[]): IngestDownload[] => {
    const seenIds = new Set<number>();
    return downloads.filter((download) => {
        if (seenIds.has(download.id)) return false;
        seenIds.add(download.id);
        return true;
    });
};

export const useFinishedIngestDownloads = () => {
    const historyQuery = useInfiniteQuery({
        queryKey: finishedIngestDownloadsQueryKey,
        queryFn: ({ pageParam = 1 }) =>
            getIngestDownloads({
                page: pageParam,
                pageSize: FINISHED_DOWNLOADS_PAGE_SIZE,
                state: 'finished',
            }),
        initialPageParam: 1,
        getNextPageParam: getNextPageNumber,
        refetchOnWindowFocus: false,
    });
    const retry = useRetryFailedPage(historyQuery);

    const downloads = removeDuplicateIds(
        historyQuery.data?.pages.flatMap((page) =>
            Array.isArray(page?.items) ? page.items : []
        ) ?? []
    );

    return {
        downloads,
        isLoading: historyQuery.isPending,
        isError: historyQuery.isError,
        isNextPageError: historyQuery.isFetchNextPageError,
        hasNextPage: Boolean(historyQuery.hasNextPage),
        isFetchingNextPage: historyQuery.isFetchingNextPage,
        fetchNextPage: () => {
            void historyQuery.fetchNextPage();
        },
        errorMessage: extractBackendErrorMessage(historyQuery.error),
        retry,
    };
};

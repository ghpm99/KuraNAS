import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getIngestDownloads, retryIngestDownload } from '@/service/ingest';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { IngestDownload } from '@/types/ingest';
import { activeIngestDownloadsQueryKey, finishedIngestDownloadsQueryKey } from './ingestQueryKeys';
import { useInFlightIds } from './useInFlightIds';

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
    const queryClient = useQueryClient();
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const retries = useInFlightIds();
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

    const retryDownload = (downloadId: number) =>
        retries.runExclusively(downloadId, async () => {
            try {
                await retryIngestDownload(downloadId);
                await Promise.all([
                    queryClient.invalidateQueries({ queryKey: activeIngestDownloadsQueryKey }),
                    queryClient.invalidateQueries({ queryKey: finishedIngestDownloadsQueryKey }),
                ]);
            } catch (error) {
                enqueueSnackbar(
                    extractBackendErrorMessage(error) ?? t('INGEST_HISTORY_RETRY_ERROR'),
                    { variant: 'error' }
                );
            }
        });

    return {
        downloads,
        retryDownload,
        isRetrying: retries.isInFlight,
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

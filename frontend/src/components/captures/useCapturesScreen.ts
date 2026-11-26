import useI18n from '@/components/i18n/provider/i18nContext';
import { captureThumbnailHref, deleteCapture, getCaptures } from '@/service/captures';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { formatSize } from '@/shared/utils/formatSize';
import type { Capture, CaptureStatus } from '@/types/captures';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';
import useReloadInfiniteListFromStart from '@/shared/hooks/useReloadInfiniteListFromStart';
import { focusRevalidationStaleTimeMs } from '@/components/providers/queryFreshness';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useMemo, useState } from 'react';

export const CAPTURES_PAGE_SIZE = 50;
export const CAPTURES_PROCESSING_REFETCH_INTERVAL_MS = 2000;
export const capturesQueryKey = ['captures'] as const;

const isCaptureProcessing = (capture: Capture): boolean =>
    capture.status === 'uploaded' || capture.status === 'promoting';

const hasProcessingCapture = (
    pages: ReadonlyArray<{ items?: Capture[] } | undefined> | undefined
): boolean =>
    (pages ?? []).some(
        (page) => Array.isArray(page?.items) && page.items.some(isCaptureProcessing)
    );

const removeDuplicateIds = (captures: Capture[]): Capture[] => {
    const seenIds = new Set<number>();
    return captures.filter((capture) => {
        if (seenIds.has(capture.id)) return false;
        seenIds.add(capture.id);
        return true;
    });
};

export const captureStatusLabelKey = (status: CaptureStatus): string => {
    switch (status) {
        case 'uploaded':
            return 'CAPTURES_STATUS_UPLOADED';
        case 'promoting':
            return 'CAPTURES_STATUS_PROMOTING';
        case 'promoted':
            return 'CAPTURES_STATUS_PROMOTED';
        case 'failed':
            return 'CAPTURES_STATUS_FAILED';
        default:
            return 'CAPTURES_STATUS_UPLOADED';
    }
};

export const useCapturesScreen = () => {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const { enqueueSnackbar } = useSnackbar();
    const [deletingId, setDeletingId] = useState<number | null>(null);
    const [captureAwaitingDeletion, setCaptureAwaitingDeletion] = useState<Capture | null>(null);

    const capturesQuery = useInfiniteQuery({
        queryKey: capturesQueryKey,
        queryFn: ({ pageParam = 1 }) =>
            getCaptures({ page: pageParam, pageSize: CAPTURES_PAGE_SIZE }),
        initialPageParam: 1,
        getNextPageParam: getNextPageNumber,
        refetchOnWindowFocus: true,
        staleTime: focusRevalidationStaleTimeMs,
        refetchInterval: (query) =>
            hasProcessingCapture(query.state.data?.pages)
                ? CAPTURES_PROCESSING_REFETCH_INTERVAL_MS
                : false,
    });
    const retryFailedPage = useRetryFailedPage(capturesQuery);
    const reloadFromStart = useReloadInfiniteListFromStart(capturesQueryKey, capturesQuery);
    const isLoading = capturesQuery.isPending;
    const isReloading = capturesQuery.isRefetching && !capturesQuery.isFetchingNextPage;

    const reload = async () => {
        const { isReloadFailed, reloadError } = await reloadFromStart();
        if (!isReloadFailed) return;
        enqueueSnackbar(extractBackendErrorMessage(reloadError) ?? t('CAPTURES_RELOAD_ERROR'), {
            variant: 'error',
        });
    };
    const isError = capturesQuery.isError;

    const removeMutation = useMutation({
        mutationFn: (id: number) => deleteCapture(id),
        onMutate: (id: number) => setDeletingId(id),
        onSuccess: (deletion) => {
            const messageKey = deletion?.library_file_kept
                ? 'CAPTURE_DELETED_LIBRARY_FILE_KEPT'
                : 'CAPTURE_DELETED';
            enqueueSnackbar(t(messageKey), { variant: 'success' });
        },
        onError: (deletionError) => {
            enqueueSnackbar(
                extractBackendErrorMessage(deletionError) ?? t('CAPTURE_DELETE_ERROR'),
                {
                    variant: 'error',
                }
            );
        },
        onSettled: () => {
            setDeletingId(null);
            queryClient.invalidateQueries({ queryKey: capturesQueryKey });
        },
    });

    const items = useMemo(
        () =>
            removeDuplicateIds(
                capturesQuery.data?.pages.flatMap((page) =>
                    Array.isArray(page?.items) ? page.items : []
                ) ?? []
            ),
        [capturesQuery.data]
    );

    const confirmRemoval = () => {
        if (!captureAwaitingDeletion) return;
        removeMutation.mutate(captureAwaitingDeletion.id);
        setCaptureAwaitingDeletion(null);
    };

    return {
        t,
        items,
        isLoading,
        isError,
        isReloading,
        reload,
        errorMessage:
            extractBackendErrorMessage(capturesQuery.error) ?? t('CONNECTION_BANNER_UNAVAILABLE'),
        retryLoading: retryFailedPage,
        isNextPageError: capturesQuery.isFetchNextPageError,
        hasNextPage: Boolean(capturesQuery.hasNextPage),
        isFetchingNextPage: capturesQuery.isFetchingNextPage,
        fetchNextPage: () => {
            void capturesQuery.fetchNextPage();
        },
        isEmpty: !isLoading && !isError && items.length === 0,
        deletingId,
        captureAwaitingDeletion,
        requestRemoval: setCaptureAwaitingDeletion,
        cancelRemoval: () => setCaptureAwaitingDeletion(null),
        confirmRemoval,
        formatSize,
        captureThumbnailHref,
        captureStatusLabelKey,
    };
};

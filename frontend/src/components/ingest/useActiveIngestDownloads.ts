import { useEffect, useRef } from 'react';
import { keepPreviousData, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { cancelJob } from '@/service/jobs';
import { getIngestDownloads, getIngestPresets } from '@/service/ingest';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { IngestDownload } from '@/types/ingest';
import { useInFlightIds } from '@/shared/hooks/useInFlightIds';
import { activeIngestDownloadsQueryKey, finishedIngestDownloadsQueryKey } from './ingestQueryKeys';

export const ACTIVE_DOWNLOADS_POLL_INTERVAL_MS = 2000;

export const useActiveIngestDownloads = () => {
    const queryClient = useQueryClient();
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const cancelations = useInFlightIds();
    const previousActiveIdsRef = useRef<Set<number>>(new Set());
    const downloadsQuery = useQuery({
        queryKey: activeIngestDownloadsQueryKey,
        queryFn: () => getIngestDownloads({ page: 1, pageSize: 50, state: 'active' }),
        placeholderData: keepPreviousData,
        refetchOnWindowFocus: false,
        refetchInterval: (query) => {
            const hasActiveDownloads = (query.state.data?.items?.length ?? 0) > 0;
            return hasActiveDownloads ? ACTIVE_DOWNLOADS_POLL_INTERVAL_MS : false;
        },
    });
    const presetsQuery = useQuery({
        queryKey: ['ingest', 'presets'],
        queryFn: getIngestPresets,
        refetchOnWindowFocus: false,
    });

    const downloads: IngestDownload[] = Array.isArray(downloadsQuery.data?.items)
        ? downloadsQuery.data.items
        : [];

    useEffect(() => {
        const activeItems = downloadsQuery.data?.items;
        const currentIds = new Set(
            Array.isArray(activeItems) ? activeItems.map((download) => download.id) : []
        );
        const hasLeftActiveList = [...previousActiveIdsRef.current].some(
            (downloadId) => !currentIds.has(downloadId)
        );
        previousActiveIdsRef.current = currentIds;
        if (hasLeftActiveList) {
            void queryClient.invalidateQueries({ queryKey: finishedIngestDownloadsQueryKey });
        }
    }, [downloadsQuery.data?.items, queryClient]);

    const presets = Array.isArray(presetsQuery.data) ? presetsQuery.data : [];
    const presetLabelsByKey = new Map(presets.map((preset) => [preset.key, preset.label]));

    const cancelDownload = (jobId: number) =>
        cancelations.runExclusively(jobId, async () => {
            try {
                await cancelJob(jobId);
                await Promise.all([
                    queryClient.invalidateQueries({ queryKey: activeIngestDownloadsQueryKey }),
                    queryClient.invalidateQueries({ queryKey: finishedIngestDownloadsQueryKey }),
                ]);
            } catch (error) {
                enqueueSnackbar(
                    extractBackendErrorMessage(error) ?? t('INGEST_ACTIVE_CANCEL_ERROR'),
                    { variant: 'error' }
                );
            }
        });

    return {
        downloads,
        cancelDownload,
        isCanceling: cancelations.isInFlight,
        presetLabelFor: (presetKey: string) => presetLabelsByKey.get(presetKey) ?? presetKey,
        isError: downloadsQuery.isError,
        errorMessage: extractBackendErrorMessage(downloadsQuery.error),
        retry: () => {
            void downloadsQuery.refetch();
        },
    };
};

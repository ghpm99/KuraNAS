import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { getIngestDownloads, getIngestPresets } from '@/service/ingest';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { IngestDownload } from '@/types/ingest';
import { activeIngestDownloadsQueryKey } from './ingestQueryKeys';

export const ACTIVE_DOWNLOADS_POLL_INTERVAL_MS = 2000;

export const useActiveIngestDownloads = () => {
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
    const presets = Array.isArray(presetsQuery.data) ? presetsQuery.data : [];
    const presetLabelsByKey = new Map(presets.map((preset) => [preset.key, preset.label]));

    return {
        downloads,
        presetLabelFor: (presetKey: string) => presetLabelsByKey.get(presetKey) ?? presetKey,
        isError: downloadsQuery.isError,
        errorMessage: extractBackendErrorMessage(downloadsQuery.error),
        retry: () => {
            void downloadsQuery.refetch();
        },
    };
};

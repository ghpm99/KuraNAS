import {
    fetchAnalyticsAIUsage,
    fetchAnalyticsDuplicateGroups,
    fetchAnalyticsDuplicates,
    fetchAnalyticsExtensions,
    fetchAnalyticsHealth,
    fetchAnalyticsHotFolders,
    fetchAnalyticsLibrary,
    fetchAnalyticsProcessing,
    fetchAnalyticsRecentFiles,
    fetchAnalyticsStorage,
    fetchAnalyticsTimeSeries,
    fetchAnalyticsTopFolders,
    fetchAnalyticsTypes,
} from '@/service/analytics';
import { AnalyticsOverview, AnalyticsPeriod, DuplicatesType } from '@/types/analytics';
import { useQueries } from '@tanstack/react-query';
import { analyticsStaleTimeMs } from '../queryFreshness';
import { useMemo, useState } from 'react';
import { AnalyticsContext } from './analyticsContext';

const defaultPeriod: AnalyticsPeriod = '7d';

interface AnalyticsProviderProps {
    children: React.ReactNode;
    duplicatesType?: DuplicatesType;
}

export const AnalyticsProvider = ({ children, duplicatesType }: AnalyticsProviderProps) => {
    const [period, setPeriod] = useState<AnalyticsPeriod>(defaultPeriod);

    const results = useQueries({
        queries: [
            { queryKey: ['analytics', 'storage', period], queryFn: () => fetchAnalyticsStorage(period), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'timeseries', period], queryFn: () => fetchAnalyticsTimeSeries(period), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'types'], queryFn: () => fetchAnalyticsTypes(), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'extensions'], queryFn: () => fetchAnalyticsExtensions(), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'recent-files'], queryFn: () => fetchAnalyticsRecentFiles(), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'top-folders'], queryFn: () => fetchAnalyticsTopFolders(), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'hot-folders', period], queryFn: () => fetchAnalyticsHotFolders(period), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'duplicates', duplicatesType ?? 'all'], queryFn: () => fetchAnalyticsDuplicates(duplicatesType), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'duplicate-groups', duplicatesType ?? 'all'], queryFn: () => fetchAnalyticsDuplicateGroups(undefined, duplicatesType), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'library'], queryFn: () => fetchAnalyticsLibrary(), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'processing'], queryFn: () => fetchAnalyticsProcessing(), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'health'], queryFn: () => fetchAnalyticsHealth(), retry: false, staleTime: analyticsStaleTimeMs },
            { queryKey: ['analytics', 'ai-usage'], queryFn: () => fetchAnalyticsAIUsage(), retry: false, staleTime: analyticsStaleTimeMs },
        ],
    });

    const [
        storageQuery,
        timeSeriesQuery,
        typesQuery,
        extensionsQuery,
        recentFilesQuery,
        topFoldersQuery,
        hotFoldersQuery,
        duplicatesQuery,
        duplicateGroupsQuery,
        libraryQuery,
        processingQuery,
        healthQuery,
        aiUsageQuery,
    ] = results;

    const value = useMemo(() => {
        const storage = storageQuery.data;
        const duplicatesSummary = duplicatesQuery.data;

        const data: AnalyticsOverview | null = storage
            ? {
                  period,
                  generated_at: storageQuery.dataUpdatedAt
                      ? new Date(storageQuery.dataUpdatedAt).toISOString()
                      : new Date().toISOString(),
                  storage: storage.storage,
                  counts: storage.counts,
                  time_series: timeSeriesQuery.data ?? [],
                  types: typesQuery.data ?? [],
                  extensions: extensionsQuery.data ?? [],
                  hot_folders: hotFoldersQuery.data ?? [],
                  top_folders: topFoldersQuery.data ?? [],
                  recent_files: recentFilesQuery.data ?? [],
                  duplicates: {
                      groups: duplicatesSummary?.groups ?? 0,
                      files: duplicatesSummary?.files ?? 0,
                      reclaimable_size: duplicatesSummary?.reclaimable_size ?? 0,
                      top_groups: duplicateGroupsQuery.data ?? [],
                  },
                  library: libraryQuery.data ?? {
                      categorized_media: 0,
                      audio_with_metadata: 0,
                      video_with_metadata: 0,
                      image_with_metadata: 0,
                      image_classified: 0,
                  },
                  processing: processingQuery.data ?? {
                      metadata_pending: 0,
                      metadata_failed: 0,
                      thumbnail_pending: 0,
                      thumbnail_failed: 0,
                      recurring_timeouts: 0,
                  },
                  health: healthQuery.data ?? {
                      status: 'ok',
                      last_scan_at: '',
                      last_scan_seconds: 0,
                      indexed_files: 0,
                      errors_last_24h: 0,
                      recent_errors: [],
                  },
                  ai_usage: aiUsageQuery.data ?? {
                      total: 0,
                      success: 0,
                      failure: 0,
                      total_tokens: 0,
                      avg_latency_ms: 0,
                  },
              }
            : null;

        return {
            period,
            duplicatesType,
            data,
            loading: results.some((query) => query.isLoading),
            error: storageQuery.isError ? 'ANALYTICS_ERROR_LOAD_BLOCK' : '',
            setPeriod,
            refresh: async () => {
                await Promise.all(results.map((query) => query.refetch()));
            },
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [period, duplicatesType, ...results.map((query) => query.data), ...results.map((query) => query.isLoading), storageQuery.isError]);

    return <AnalyticsContext.Provider value={value}>{children}</AnalyticsContext.Provider>;
};

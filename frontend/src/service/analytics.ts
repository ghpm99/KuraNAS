import { apiBase } from './index';
import {
    AIUsage,
    AnalyticsPeriod,
    DuplicateGroup,
    DuplicatesSummary,
    ExtensionStat,
    FolderUsage,
    HealthStatus,
    HotFolder,
    LibraryStats,
    ProcessingStats,
    RecentFile,
    StorageStatsResponse,
    TimeSeriesPoint,
    TypeBreakdown,
} from '@/types/analytics';

export const fetchAnalyticsStorage = async (
    period: AnalyticsPeriod
): Promise<StorageStatsResponse> => {
    const response = await apiBase.get<StorageStatsResponse>('/analytics/storage', {
        params: { period },
    });
    return response.data;
};

export const fetchAnalyticsTimeSeries = async (
    period: AnalyticsPeriod
): Promise<TimeSeriesPoint[]> => {
    const response = await apiBase.get<TimeSeriesPoint[]>('/analytics/timeseries', {
        params: { period },
    });
    return response.data;
};

export const fetchAnalyticsTypes = async (): Promise<TypeBreakdown[]> => {
    const response = await apiBase.get<TypeBreakdown[]>('/analytics/types');
    return response.data;
};

export const fetchAnalyticsExtensions = async (limit?: number): Promise<ExtensionStat[]> => {
    const response = await apiBase.get<ExtensionStat[]>('/analytics/extensions', {
        params: limit ? { limit } : undefined,
    });
    return response.data;
};

export const fetchAnalyticsRecentFiles = async (limit?: number): Promise<RecentFile[]> => {
    const response = await apiBase.get<RecentFile[]>('/analytics/recent-files', {
        params: limit ? { limit } : undefined,
    });
    return response.data;
};

export const fetchAnalyticsTopFolders = async (limit?: number): Promise<FolderUsage[]> => {
    const response = await apiBase.get<FolderUsage[]>('/analytics/top-folders', {
        params: limit ? { limit } : undefined,
    });
    return response.data;
};

export const fetchAnalyticsHotFolders = async (
    period: AnalyticsPeriod,
    limit?: number
): Promise<HotFolder[]> => {
    const response = await apiBase.get<HotFolder[]>('/analytics/hot-folders', {
        params: { period, ...(limit ? { limit } : {}) },
    });
    return response.data;
};

export const fetchAnalyticsDuplicates = async (): Promise<DuplicatesSummary> => {
    const response = await apiBase.get<DuplicatesSummary>('/analytics/duplicates');
    return response.data;
};

export const fetchAnalyticsDuplicateGroups = async (
    limit?: number
): Promise<DuplicateGroup[]> => {
    const response = await apiBase.get<DuplicateGroup[]>('/analytics/duplicates/groups', {
        params: limit ? { limit } : undefined,
    });
    return response.data;
};

export const fetchAnalyticsLibrary = async (): Promise<LibraryStats> => {
    const response = await apiBase.get<LibraryStats>('/analytics/library');
    return response.data;
};

export const fetchAnalyticsProcessing = async (): Promise<ProcessingStats> => {
    const response = await apiBase.get<ProcessingStats>('/analytics/processing');
    return response.data;
};

export const fetchAnalyticsHealth = async (): Promise<HealthStatus> => {
    const response = await apiBase.get<HealthStatus>('/analytics/health');
    return response.data;
};

export const fetchAnalyticsAIUsage = async (): Promise<AIUsage> => {
    const response = await apiBase.get<AIUsage>('/analytics/ai-usage');
    return response.data;
};

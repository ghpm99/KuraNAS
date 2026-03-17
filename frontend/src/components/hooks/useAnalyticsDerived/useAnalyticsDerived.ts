import { AnalyticsOverview } from '@/types/analytics';

export const useAnalyticsDerived = (data: AnalyticsOverview | null) => {
    if (!data || data.storage.total_bytes <= 0) {
        return {
            usedPercent: 0,
            reclaimablePercent: 0,
        };
    }

    const usedPercent = (data.storage.used_bytes / data.storage.total_bytes) * 100;
    const reclaimablePercent = (data.duplicates.reclaimable_size / data.storage.total_bytes) * 100;

    return {
        usedPercent,
        reclaimablePercent,
    };
};

import useI18n from '@/components/i18n/provider/i18nContext';
import { useAnalyticsFormatters } from '@/components/hooks/useAnalyticsFormatters/useAnalyticsFormatters';
import { useAnalyticsOverview } from '@/components/providers/analyticsProvider/analyticsContext';
import AnalyticsKpiCard from '@/components/analytics/AnalyticsKpiCard';
import AnalyticsSection from '@/components/analytics/AnalyticsSection';
import { useAnalyticsBlockQuery } from '@/components/analytics/useAnalyticsBlockQuery';
import { fetchAnalyticsStorage } from '@/service/analytics';
import styles from './AnalyticsScreen.module.css';

const AnalyticsStorageIndicators = () => {
    const { t } = useI18n();
    const { formatBytes } = useAnalyticsFormatters();
    const { period } = useAnalyticsOverview();
    const storageBlock = useAnalyticsBlockQuery({
        queryKey: ['analytics', 'storage', period],
        fetchBlock: () => fetchAnalyticsStorage(period),
    });

    const storage = storageBlock.data?.storage;
    const counts = storageBlock.data?.counts;
    const growthPrefix = storage && storage.growth_bytes > 0 ? '+' : '';

    return (
        <AnalyticsSection
            title={t('ANALYTICS_KPI_STORAGE')}
            loading={storageBlock.status === 'pending'}
            errorKey={storageBlock.status === 'error' ? 'ANALYTICS_ERROR_LOAD_BLOCK' : undefined}
            errorMessage={storageBlock.errorMessage}
            onRetry={storageBlock.retry}
        >
            {storage && counts ? (
                <div className={styles.grid}>
                    <div className={styles.span4}>
                        <AnalyticsKpiCard
                            title={t('ANALYTICS_KPI_STORAGE')}
                            value={`${formatBytes(storage.used_bytes)} / ${formatBytes(storage.total_bytes)}`}
                            helpText={`${t('ANALYTICS_FREE')}: ${formatBytes(storage.free_bytes)}`}
                        />
                    </div>
                    <div className={styles.span4}>
                        <AnalyticsKpiCard
                            title={t('ANALYTICS_KPI_GROWTH')}
                            value={`${growthPrefix}${formatBytes(storage.growth_bytes)}`}
                            helpText={t('ANALYTICS_KPI_GROWTH_HELP')}
                        />
                    </div>
                    <div className={styles.span4}>
                        <AnalyticsKpiCard
                            title={t('ANALYTICS_KPI_FILES_ADDED')}
                            value={String(counts.files_added)}
                            helpText={`${t('ANALYTICS_FILE_COUNT')}: ${counts.files_total.toLocaleString()}`}
                        />
                    </div>
                    <div className={styles.span4}>
                        <AnalyticsKpiCard
                            title={t('ANALYTICS_KPI_FOLDERS_TOTAL')}
                            value={counts.folders.toLocaleString()}
                            helpText={t('ANALYTICS_FOLDERS_STORED')}
                        />
                    </div>
                </div>
            ) : null}
        </AnalyticsSection>
    );
};

export default AnalyticsStorageIndicators;

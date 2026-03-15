import useI18n from '@/components/i18n/provider/i18nContext';
import { useAnalyticsDerived } from '@/components/hooks/useAnalyticsDerived/useAnalyticsDerived';
import { useAnalyticsFormatters } from '@/components/hooks/useAnalyticsFormatters/useAnalyticsFormatters';
import { useAnalyticsOverview } from '@/components/providers/analyticsProvider/analyticsContext';
import { useEffect, useMemo, useState } from 'react';

export const useAnalyticsScreenState = () => {
	const { t } = useI18n();
	const { formatBytes, formatPercent, formatDate } = useAnalyticsFormatters();
	const { period, setPeriod, data, loading, error, refresh } = useAnalyticsOverview();
	const { usedPercent, reclaimablePercent } = useAnalyticsDerived(data);
	const [now, setNow] = useState(() => Date.now());

	useEffect(() => {
		const timer = window.setInterval(() => setNow(Date.now()), 60000);
		return () => window.clearInterval(timer);
	}, []);

	const updatedMinutes = (() => {
		if (!data?.generated_at) {
			return '-';
		}

		const generatedTime = new Date(data.generated_at).getTime();
		if (Number.isNaN(generatedTime)) {
			return '-';
		}

		const minutes = Math.max(0, Math.floor((now - generatedTime) / 60000));
		return String(minutes);
	})();

	const healthStatusLabel = useMemo(() => {
		switch (data?.health.status) {
			case 'scanning':
				return t('ANALYTICS_STATUS_SCANNING');
			case 'error':
				return t('ANALYTICS_STATUS_ERROR');
			default:
				return t('ANALYTICS_STATUS_OK');
		}
	}, [data?.health.status, t]);

	const healthStatusColor = useMemo<'error' | 'warning' | 'success'>(() => {
		if (data?.health.status === 'error') {
			return 'error';
		}

		if (data?.health.status === 'scanning') {
			return 'warning';
		}

		return 'success';
	}, [data?.health.status]);

	return {
		t,
		period,
		setPeriod,
		data,
		loading,
		error,
		refresh,
		formatBytes,
		formatPercent,
		formatDate,
		usedPercent,
		reclaimablePercent,
		updatedMinutes,
		healthStatusLabel,
		healthStatusColor,
		processingFailureTotal: (data?.processing.metadata_failed ?? 0) + (data?.processing.thumbnail_failed ?? 0),
	};
};

export type AnalyticsScreenState = ReturnType<typeof useAnalyticsScreenState>;

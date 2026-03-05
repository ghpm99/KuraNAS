import { fetchAnalyticsOverview } from '@/service/analytics';
import { AnalyticsPeriod } from '@/types/analytics';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { AnalyticsContext } from './analyticsContext';

const defaultPeriod: AnalyticsPeriod = '7d';

export const AnalyticsProvider = ({ children }: { children: React.ReactNode }) => {
	const [period, setPeriod] = useState<AnalyticsPeriod>(defaultPeriod);
	const [data, setData] = useState<Awaited<ReturnType<typeof fetchAnalyticsOverview>> | null>(null);
	const [loading, setLoading] = useState<boolean>(true);
	const [error, setError] = useState<string>('');

	const load = useCallback(async () => {
		setLoading(true);
		setError('');
		try {
			const overview = await fetchAnalyticsOverview(period);
			setData(overview);
		} catch {
			setError('ANALYTICS_ERROR_LOAD_BLOCK');
		} finally {
			setLoading(false);
		}
	}, [period]);

	useEffect(() => {
		void load();
	}, [load]);

	const value = useMemo(
		() => ({
			period,
			data,
			loading,
			error,
			setPeriod,
			refresh: load,
		}),
		[period, data, loading, error, load],
	);

	return <AnalyticsContext.Provider value={value}>{children}</AnalyticsContext.Provider>;
};

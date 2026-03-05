import { fetchAnalyticsOverview } from '@/service/analytics';
import { AnalyticsPeriod } from '@/types/analytics';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { AnalyticsContext } from './analyticsContext';

const defaultPeriod: AnalyticsPeriod = '7d';

export const AnalyticsProvider = ({ children }: { children: React.ReactNode }) => {
	const [period, setPeriod] = useState<AnalyticsPeriod>(defaultPeriod);
	const [data, setData] = useState<Awaited<ReturnType<typeof fetchAnalyticsOverview>> | null>(null);
	const [loading, setLoading] = useState<boolean>(true);
	const [error, setError] = useState<string>('');
	const latestRequestIdRef = useRef(0);

	const load = useCallback(async () => {
		const requestId = ++latestRequestIdRef.current;
		setLoading(true);
		setError('');
		try {
			const overview = await fetchAnalyticsOverview(period);
			if (requestId === latestRequestIdRef.current) {
				setData(overview);
			}
		} catch {
			if (requestId === latestRequestIdRef.current) {
				setError('ANALYTICS_ERROR_LOAD_BLOCK');
			}
		} finally {
			if (requestId === latestRequestIdRef.current) {
				setLoading(false);
			}
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

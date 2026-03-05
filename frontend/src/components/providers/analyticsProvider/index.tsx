import { fetchAnalyticsOverview } from '@/service/analytics';
import { AnalyticsPeriod } from '@/types/analytics';
import { useQuery } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import { AnalyticsContext } from './analyticsContext';

const defaultPeriod: AnalyticsPeriod = '7d';

export const AnalyticsProvider = ({ children }: { children: React.ReactNode }) => {
	const [period, setPeriod] = useState<AnalyticsPeriod>(defaultPeriod);
	const { data = null, isLoading, isFetching, isError, refetch } = useQuery({
		queryKey: ['analytics-overview', period],
		queryFn: () => fetchAnalyticsOverview(period),
		retry: false,
	});

	const value = useMemo(
		() => ({
			period,
			data,
			loading: isLoading || isFetching,
			error: isError ? 'ANALYTICS_ERROR_LOAD_BLOCK' : '',
			setPeriod,
			refresh: async () => {
				await refetch();
			},
		}),
		[period, data, isLoading, isFetching, isError, refetch],
	);

	return <AnalyticsContext.Provider value={value}>{children}</AnalyticsContext.Provider>;
};

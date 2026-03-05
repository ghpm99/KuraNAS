import { apiBase } from './index';
import { AnalyticsOverview, AnalyticsPeriod } from '@/types/analytics';

export const fetchAnalyticsOverview = async (period: AnalyticsPeriod): Promise<AnalyticsOverview> => {
	const response = await apiBase.get<AnalyticsOverview>('/analytics/overview', {
		params: { period },
	});
	return response.data;
};

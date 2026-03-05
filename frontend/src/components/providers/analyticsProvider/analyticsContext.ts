import { createContext, useContext } from 'react';
import { AnalyticsOverview, AnalyticsPeriod } from '@/types/analytics';

export interface AnalyticsContextType {
	period: AnalyticsPeriod;
	data: AnalyticsOverview | null;
	loading: boolean;
	error: string;
	setPeriod: (period: AnalyticsPeriod) => void;
	refresh: () => Promise<void>;
}

export const AnalyticsContext = createContext<AnalyticsContextType | undefined>(undefined);

export const useAnalyticsOverview = () => {
	const context = useContext(AnalyticsContext);
	if (!context) {
		throw new Error('useAnalyticsOverview must be used within an AnalyticsProvider');
	}
	return context;
};

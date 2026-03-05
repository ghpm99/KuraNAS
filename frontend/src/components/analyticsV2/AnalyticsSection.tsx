import { Card, CardContent, CardHeader, Skeleton } from '@mui/material';
import AnalyticsEmptyState from './AnalyticsEmptyState';
import AnalyticsErrorState from './AnalyticsErrorState';

interface AnalyticsSectionProps {
	title: string;
	loading: boolean;
	errorKey?: string;
	empty?: boolean;
	emptyKey?: string;
	children: React.ReactNode;
}

export default function AnalyticsSection({ title, loading, errorKey, empty, emptyKey, children }: AnalyticsSectionProps) {
	return (
		<Card>
			<CardHeader title={title} titleTypographyProps={{ variant: 'h6' }} />
			<CardContent>
				{loading ? <Skeleton variant='rectangular' height={160} /> : null}
				{!loading && errorKey ? <AnalyticsErrorState messageKey={errorKey} /> : null}
				{!loading && !errorKey && empty ? <AnalyticsEmptyState messageKey={emptyKey ?? 'ANALYTICS_EMPTY'} /> : null}
				{!loading && !errorKey && !empty ? children : null}
			</CardContent>
		</Card>
	);
}

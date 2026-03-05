import AnalyticsSection from './AnalyticsSection';

interface AnalyticsChartCardProps {
	title: string;
	loading: boolean;
	errorKey?: string;
	empty?: boolean;
	emptyKey?: string;
	children: React.ReactNode;
}

export default function AnalyticsChartCard(props: AnalyticsChartCardProps) {
	return <AnalyticsSection {...props}>{props.children}</AnalyticsSection>;
}

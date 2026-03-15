import AnalyticsContent from '@/components/analytics/AnalyticsContent';
import Layout from '@/components/layout/Layout';
import { AnalyticsProvider } from '@/components/providers/analyticsProvider';

const AnalyticsPage = () => {
	return (
		<Layout>
			<AnalyticsProvider>
				<AnalyticsContent />
			</AnalyticsProvider>
		</Layout>
	);
};

export default AnalyticsPage;

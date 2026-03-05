import { AnalyticsProvider } from '../contexts/AnalyticsContext';
import Layout from '../layout/Layout';

const AnalyticsLayout = ({ children }: { children: React.ReactNode }) => {
	return (
		<AnalyticsProvider>
			<Layout>{children}</Layout>
		</AnalyticsProvider>
	);
};

export default AnalyticsLayout;

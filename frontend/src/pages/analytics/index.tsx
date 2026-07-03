import AnalyticsContent from '@/components/analytics/AnalyticsContent';
import { AnalyticsProvider } from '@/components/providers/analyticsProvider';

const AnalyticsPage = () => {
    return (
        <AnalyticsProvider>
            <AnalyticsContent />
        </AnalyticsProvider>
    );
};

export default AnalyticsPage;

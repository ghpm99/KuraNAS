import AnalyticsContent from '@/components/analytics/AnalyticsContent';
import { AnalyticsProvider } from '@/components/providers/analyticsProvider';
import { parseDuplicatesType } from '@/components/analytics/duplicatesTypeFilter';
import { useLocation } from 'react-router-dom';

const AnalyticsPage = () => {
    const { search } = useLocation();

    return (
        <AnalyticsProvider duplicatesType={parseDuplicatesType(search)}>
            <AnalyticsContent />
        </AnalyticsProvider>
    );
};

export default AnalyticsPage;

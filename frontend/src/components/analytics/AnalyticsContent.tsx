import AnalyticsDomainHeader from '@/components/analytics/AnalyticsDomainHeader';
import AnalyticsLibraryScreen from '@/components/analytics/AnalyticsLibraryScreen';
import AnalyticsOverviewScreen from '@/components/analytics/AnalyticsOverviewScreen';
import AnalyticsSidebar from '@/components/analytics/AnalyticsSidebar';
import AnalyticsToolbar from '@/components/analytics/AnalyticsToolbar';
import { useAnalyticsNavigation } from '@/components/analytics/useAnalyticsNavigation';
import { useAnalyticsScreenState } from '@/components/analytics/useAnalyticsScreenState';
import styles from './AnalyticsContent.module.css';

const AnalyticsContent = () => {
    const state = useAnalyticsScreenState();
    const { currentSection } = useAnalyticsNavigation();

    return (
        <div className={styles.page}>
            <AnalyticsDomainHeader />
            <div className={styles.content}>
                <AnalyticsSidebar />
                <div className={styles.main}>
                    <AnalyticsToolbar state={state} />
                    {currentSection === 'library' ? (
                        <AnalyticsLibraryScreen state={state} />
                    ) : (
                        <AnalyticsOverviewScreen state={state} />
                    )}
                </div>
            </div>
        </div>
    );
};

export default AnalyticsContent;

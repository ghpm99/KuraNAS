import AnalyticsDomainHeader from '@/components/analytics/AnalyticsDomainHeader';
import AnalyticsDomainNav from '@/components/analytics/AnalyticsDomainNav';
import AnalyticsLibraryScreen from '@/components/analytics/AnalyticsLibraryScreen';
import AnalyticsOverviewScreen from '@/components/analytics/AnalyticsOverviewScreen';
import AnalyticsToolbar from '@/components/analytics/AnalyticsToolbar';
import DomainPageLayout from '@/components/layout/DomainPageLayout';
import { useAnalyticsNavigation } from '@/components/analytics/useAnalyticsNavigation';
import { useAnalyticsScreenState } from '@/components/analytics/useAnalyticsScreenState';
import styles from './AnalyticsContent.module.css';

const AnalyticsContent = () => {
    const state = useAnalyticsScreenState();
    const { currentSection } = useAnalyticsNavigation();

    return (
        <DomainPageLayout header={<AnalyticsDomainHeader />} nav={<AnalyticsDomainNav />}>
            <div className={styles.main}>
                <AnalyticsToolbar state={state} />
                {currentSection === 'library' ? (
                    <AnalyticsLibraryScreen state={state} />
                ) : (
                    <AnalyticsOverviewScreen state={state} />
                )}
            </div>
        </DomainPageLayout>
    );
};

export default AnalyticsContent;

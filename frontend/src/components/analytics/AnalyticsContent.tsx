import AnalyticsDomainHeader from '@/components/analytics/AnalyticsDomainHeader';
import AnalyticsDomainNav from '@/components/analytics/AnalyticsDomainNav';
import AnalyticsDuplicatesScreen from '@/components/analytics/AnalyticsDuplicatesScreen';
import AnalyticsLargestFilesScreen from '@/components/analytics/AnalyticsLargestFilesScreen';
import AnalyticsLibraryScreen from '@/components/analytics/AnalyticsLibraryScreen';
import AnalyticsOverviewScreen from '@/components/analytics/AnalyticsOverviewScreen';
import AnalyticsToolbar from '@/components/analytics/AnalyticsToolbar';
import DomainPageLayout from '@/components/layout/DomainPageLayout';
import type { AnalyticsSection } from '@/app/routes';
import { useAnalyticsNavigation } from '@/components/analytics/useAnalyticsNavigation';
import { useAnalyticsScreenState } from '@/components/analytics/useAnalyticsScreenState';
import styles from './AnalyticsContent.module.css';

const sectionsWithAnalyticsToolbar: AnalyticsSection[] = ['overview', 'library'];

const AnalyticsContent = () => {
    const state = useAnalyticsScreenState();
    const { currentSection } = useAnalyticsNavigation();

    return (
        <DomainPageLayout header={<AnalyticsDomainHeader />} nav={<AnalyticsDomainNav />}>
            <div className={styles.main}>
                {sectionsWithAnalyticsToolbar.includes(currentSection) ? (
                    <AnalyticsToolbar state={state} />
                ) : null}
                {currentSection === 'library' ? <AnalyticsLibraryScreen state={state} /> : null}
                {currentSection === 'overview' ? <AnalyticsOverviewScreen state={state} /> : null}
                {currentSection === 'largest' ? <AnalyticsLargestFilesScreen /> : null}
                {currentSection === 'duplicates' ? <AnalyticsDuplicatesScreen /> : null}
            </div>
        </DomainPageLayout>
    );
};

export default AnalyticsContent;

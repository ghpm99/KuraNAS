import AnalyticsLayout from '@/components/analytics/analyticsLayout';
import StorageOverviewCards from '@/components/analytics/StorageOverviewCards/StorageOverviewCards';
import styles from './analytics.module.css';
import DiskUsageChart from '@/components/analytics/DiskUsageChart/DiskUsageChart';
import FileTypesChart from '@/components/analytics/FileTypesChart/FileTypesChart';
import FileTypesTable from '@/components/analytics/FileTypesTable/FileTypesTable';
import SizeRangesChart from '@/components/analytics/SizeRangesChart/SizeRangesChart';
import LargestFilesTable from '@/components/analytics/LargestFilesTable/LargestFilesTable';
import DuplicatesSection from '@/components/analytics/DuplicatesSection/DuplicatesSection';
import ActivityChart from '@/components/analytics/ActivityChart/ActivityChart';
import RecentActivity from '@/components/analytics/RecentActivity/RecentActivity';
import EmptyFoldersSection from '@/components/analytics/EmptyFoldersSection/EmptyFoldersSection';
import CleanupSuggestions from '@/components/analytics/CleanupSuggestions/CleanupSuggestions';
import BackupSection from '@/components/analytics/BackupSection/BackupSection';
import TrashSection from '@/components/analytics/TrashSection/TrashSection';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Button from '@/components/ui/Button/Button';
import useI18n from '@/components/i18n/provider/i18nContext';

const AnalyticsContent = () => {
	const { refreshAnalytics } = useAnalytics();
	const { t } = useI18n();
	return (
		<div className={styles.content}>
			<div className={styles.header}>
				<h1 className={styles.pageTitle}>{t('ANALYTICS_PAGE_TITLE')}</h1>
				<p className={styles.pageDescription}>{t('ANALYTICS_PAGE_DESCRIPTION')}</p>
				<Button className={styles.refreshButton} onClick={refreshAnalytics}>
					{t('ANALYTICS_REFRESH')}
				</Button>
			</div>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_STORAGE_OVERVIEW')}</h2>
				<StorageOverviewCards />
				<div className={styles.chartGrid}>
					<DiskUsageChart />
				</div>
			</section>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_FILE_TYPES_SIZES')}</h2>
				<div className={styles.chartsGrid}>
					<FileTypesChart />
					<FileTypesTable />
					<SizeRangesChart />
					<LargestFilesTable />
				</div>
			</section>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_DUPLICATES_REDUNDANCY')}</h2>
				<DuplicatesSection />
			</section>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_RECENT_ACTIVITY_SECTION')}</h2>
				<div className={styles.activityGrid}>
					<ActivityChart />
					<RecentActivity />
				</div>
			</section>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_EMPTY_FOLDERS_SECTION')}</h2>
				<EmptyFoldersSection />
			</section>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_CLEANUP_SECTION')}</h2>
				<CleanupSuggestions />
			</section>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_BACKUP_SECTION')}</h2>
				<BackupSection />
			</section>

			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>{t('ANALYTICS_TRASH_SECTION')}</h2>
				<TrashSection />
			</section>
		</div>
	);
};

const AnalyticsPage = () => {
	return (
		<AnalyticsLayout>
			<AnalyticsContent />
		</AnalyticsLayout>
	);
};

export default AnalyticsPage;

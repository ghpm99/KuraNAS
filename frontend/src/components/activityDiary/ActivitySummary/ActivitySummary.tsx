import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDuration } from '@/utils';
import styles from './ActivitySummary.module.css';
import Card from '@/components/ui/Card/Card';
import useI18n from '@/components/i18n/provider/i18nContext';

const ActivitySummary = () => {
	const { data } = useActivityDiary();
	const { t } = useI18n();
	return (
		<Card title={t('DAY_SUMMARY_TITLE')}>
			<div className={styles['summary-grid']}>
				<div className={styles['summary-item']}>
					<h3>{t('TOTAL_ACTIVITIES')}</h3>
					<p className={styles['summary-value']}>{data?.summary?.total_activities}</p>
				</div>
				<div className={styles['summary-item']}>
					<h3>{t('TOTAL_WORKED_TIME')}</h3>
					<p className={styles['summary-value']}>{formatDuration(data?.summary?.total_time_spent_seconds)}</p>
				</div>
				<div className={styles['summary-item']}>
					<h3>{t('LONGEST_ACTIVITY')}</h3>
					<p className={styles['summary-value']}>{data?.summary?.longest_activity?.name}</p>
					<p className={styles['summary-subvalue']}>
						{formatDuration(data?.summary?.longest_activity?.duration_seconds)}
					</p>
				</div>
			</div>
		</Card>
	);
};

export default ActivitySummary;

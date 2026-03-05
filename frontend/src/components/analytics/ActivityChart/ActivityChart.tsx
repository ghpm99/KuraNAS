import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import Card from '../../ui/Card/Card';
import styles from './ActivityChart.module.css';

export default function ActivityChart() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { activityChart } = analyticsData.recentActivity;

	const maxValue = Math.max(...activityChart.map((day) => Math.max(day.created, day.modified)));

	return (
		<Card title={t('ANALYTICS_ACTIVITY_CHART')}>
			<div className={styles.chart}>
				<div className={styles.chartArea}>
					{activityChart.map((day) => (
						<div key={day.date} className={styles.dayColumn}>
							<div className={styles.bars}>
								<div
									className={`${styles.bar} ${styles.created}`}
									style={{ height: `${(day.created / maxValue) * 100}%` }}
									title={`${t('ANALYTICS_CREATED')}: ${day.created}`}
								></div>
								<div
									className={`${styles.bar} ${styles.modified}`}
									style={{ height: `${(day.modified / maxValue) * 100}%` }}
									title={`${t('ANALYTICS_MODIFIED')}: ${day.modified}`}
								></div>
							</div>
							<div className={styles.dayLabel}>
								{new Date(day.date).toLocaleDateString(undefined, { day: '2-digit', month: '2-digit' })}
							</div>
						</div>
					))}
				</div>
				<div className={styles.legend}>
					<div className={styles.legendItem}>
						<div className={`${styles.legendColor} ${styles.createdColor}`}></div>
						<span>{t('ANALYTICS_CREATED')}</span>
					</div>
					<div className={styles.legendItem}>
						<div className={`${styles.legendColor} ${styles.modifiedColor}`}></div>
						<span>{t('ANALYTICS_MODIFIED')}</span>
					</div>
				</div>
			</div>
		</Card>
	);
}

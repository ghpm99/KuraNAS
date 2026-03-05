import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import Card from '../../ui/Card/Card';
import styles from './SizeRangesChart.module.css';

export default function SizeRangesChart() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { sizeRanges } = analyticsData;

	const maxCount = Math.max(...sizeRanges.map((range) => range.count));

	return (
		<Card title={t('ANALYTICS_SIZE_DISTRIBUTION')}>
			<div className={styles.chart}>
				{sizeRanges.map((range) => (
					<div key={range.range} className={styles.bar}>
						<div className={styles.barContainer}>
							<div className={styles.barFill} style={{ height: `${(range.count / maxCount) * 100}%` }}></div>
						</div>
						<div className={styles.barLabel}>{range.range}</div>
						<div className={styles.barValue}>{range.count.toLocaleString()}</div>
					</div>
				))}
			</div>
		</Card>
	);
}

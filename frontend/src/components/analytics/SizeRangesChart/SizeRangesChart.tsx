import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Card from '../../ui/Card/Card';
import styles from './SizeRangesChart.module.css';

export default function SizeRangesChart() {
	const { analyticsData } = useAnalytics();
	const { sizeRanges } = analyticsData;

	const maxCount = Math.max(...sizeRanges.map((range) => range.count));

	return (
		<Card title='Distribuição por Tamanho'>
			<div className={styles.chart}>
				{sizeRanges.map((range, index) => (
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

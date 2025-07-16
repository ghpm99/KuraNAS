import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Card from '../../ui/Card/Card';
import styles from './DiskUsageChart.module.css';

export default function DiskUsageChart() {
	const { analyticsData } = useAnalytics();
	const { diskUsage } = analyticsData.storageOverview;

	const circumference = 2 * Math.PI * 45;
	const usedOffset = circumference - (diskUsage.used / 100) * circumference;

	return (
		<Card title='Uso de Disco'>
			<div className={styles.chartContainer}>
				<div className={styles.chart}>
					<svg width='120' height='120' viewBox='0 0 120 120'>
						<circle cx='60' cy='60' r='45' fill='none' stroke='#e5e7eb' strokeWidth='10' />
						<circle
							cx='60'
							cy='60'
							r='45'
							fill='none'
							stroke='#3b82f6'
							strokeWidth='10'
							strokeDasharray={circumference}
							strokeDashoffset={usedOffset}
							strokeLinecap='round'
							transform='rotate(-90 60 60)'
						/>
					</svg>
					<div className={styles.centerText}>
						<div className={styles.percentage}>{diskUsage.used}%</div>
						<div className={styles.label}>Usado</div>
					</div>
				</div>
				<div className={styles.legend}>
					<div className={styles.legendItem}>
						<div className={`${styles.legendColor} ${styles.used}`}></div>
						<span>Usado ({diskUsage.used}%)</span>
					</div>
					<div className={styles.legendItem}>
						<div className={`${styles.legendColor} ${styles.free}`}></div>
						<span>Livre ({diskUsage.free}%)</span>
					</div>
				</div>
			</div>
		</Card>
	);
}

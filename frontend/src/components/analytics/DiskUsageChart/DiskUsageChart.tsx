import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Card from '../../ui/Card/Card';
import styles from './DiskUsageChart.module.css';
import { PieChart } from '@mui/x-charts';

export default function DiskUsageChart() {
	const { analyticsData } = useAnalytics();
	const { diskUsage } = analyticsData.storageOverview;

	const settings = {
		margin: { right: 5 },
		width: 200,
		height: 200,
		hideLegend: true,
	};

	const data = [
		{ label: 'Usado', value: diskUsage.used, color: '#0088FE' },
		{ label: 'Livre', value: diskUsage.free, color: '#00C49F' },
	];

	return (
		<Card title='Uso de Disco'>
			<div className={styles.chartContainer}>
				<div className={styles.chart}>
					<PieChart series={[{ innerRadius: 50, outerRadius: 100, data, arcLabel: 'value' }]} {...settings} />
					<div className={styles.centerText}>
						<div className={styles.percentage}>{diskUsage.used}%</div>
						<div className={styles.label}>Usado</div>
					</div>
				</div>
			</div>
		</Card>
	);
}

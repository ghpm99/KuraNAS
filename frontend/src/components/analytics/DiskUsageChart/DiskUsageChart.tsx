import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import Card from '../../ui/Card/Card';
import styles from './DiskUsageChart.module.css';
import { PieChart } from '@mui/x-charts';

export default function DiskUsageChart() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { diskUsage } = analyticsData.storageOverview;

	const settings = {
		margin: { right: 5 },
		width: 200,
		height: 200,
		hideLegend: true,
	};

	const data = [
		{ label: t('ANALYTICS_USED'), value: diskUsage.used, color: '#0088FE' },
		{ label: t('ANALYTICS_FREE'), value: diskUsage.free, color: '#00C49F' },
	];

	return (
		<Card title={t('ANALYTICS_DISK_USAGE')}>
			<div className={styles.chartContainer}>
				<div className={styles.chart}>
					<PieChart series={[{ innerRadius: 50, outerRadius: 100, data, arcLabel: 'value' }]} {...settings} />
					<div className={styles.centerText}>
						<div className={styles.percentage}>{diskUsage.used}%</div>
						<div className={styles.label}>{t('ANALYTICS_USED')}</div>
					</div>
				</div>
			</div>
		</Card>
	);
}

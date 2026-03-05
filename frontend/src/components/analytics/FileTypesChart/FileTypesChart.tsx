import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import Card from '../../ui/Card/Card';
import styles from './FileTypesChart.module.css';
import { PieChart } from '@mui/x-charts';
import { formatSize } from '@/utils';

export default function FileTypesChart() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { fileTypes } = analyticsData;

	const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

	return (
		<Card title={t('ANALYTICS_FILE_TYPE_DISTRIBUTION')}>
			<div className={styles.container}>
				<div className={styles.chart}>
					<PieChart
						series={[
							{
								data: fileTypes.map((type) => ({
									label: type.format,
									value: type.size,
								})),
							},
						]}
						hideLegend
						width={200}
						height={200}
					/>
				</div>
				<div className={styles.legend}>
					{fileTypes.map((type, index) => (
						<div key={type.format ?? index} className={styles.legendItem}>
							<div className={styles.legendColor} style={{ backgroundColor: colors[index % colors.length] }}></div>
							<div className={styles.legendText}>
								<div className={styles.typeName}>{type.format}</div>
								<div className={styles.typeStats}>
									{type.total} {t('ANALYTICS_FILES_COUNT_LABEL')} • {formatSize(type.size)}
								</div>
							</div>
						</div>
					))}
				</div>
			</div>
		</Card>
	);
}

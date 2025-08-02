import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Card from '../../ui/Card/Card';
import styles from './FileTypesChart.module.css';
import { PieChart } from '@mui/x-charts';
import { formatSize } from '@/utils';

export default function FileTypesChart() {
	const { analyticsData } = useAnalytics();
	const { fileTypes } = analyticsData;

	const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

	console.log('File Types Data:', fileTypes);
	return (
		<Card title='Distribuição por Tipo de Arquivo'>
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
									{type.total} arquivos • {formatSize(type.size)}
								</div>
							</div>
						</div>
					))}
				</div>
			</div>
		</Card>
	);
}

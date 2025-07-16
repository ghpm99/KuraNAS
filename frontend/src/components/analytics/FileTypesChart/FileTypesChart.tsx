import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Card from '../../ui/Card/Card';
import styles from './FileTypesChart.module.css';

export default function FileTypesChart() {
	const { analyticsData } = useAnalytics();
	const { fileTypes } = analyticsData;

	const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

	return (
		<Card title='Distribuição por Tipo de Arquivo'>
			<div className={styles.container}>
				<div className={styles.chart}>
					<svg width='200' height='200' viewBox='0 0 200 200'>
						{fileTypes.map((type, index) => {
							const startAngle = fileTypes.slice(0, index).reduce((sum, t) => sum + (t.percentage * 360) / 100, 0);
							const endAngle = startAngle + (type.percentage * 360) / 100;

							const startAngleRad = (startAngle * Math.PI) / 180;
							const endAngleRad = (endAngle * Math.PI) / 180;

							const largeArcFlag = type.percentage > 50 ? 1 : 0;

							const x1 = 100 + 80 * Math.cos(startAngleRad);
							const y1 = 100 + 80 * Math.sin(startAngleRad);
							const x2 = 100 + 80 * Math.cos(endAngleRad);
							const y2 = 100 + 80 * Math.sin(endAngleRad);

							const pathData = [`M 100 100`, `L ${x1} ${y1}`, `A 80 80 0 ${largeArcFlag} 1 ${x2} ${y2}`, `Z`].join(' ');

							return <path key={type.type} d={pathData} fill={colors[index % colors.length]} />;
						})}
					</svg>
				</div>
				<div className={styles.legend}>
					{fileTypes.map((type, index) => (
						<div key={type.type} className={styles.legendItem}>
							<div className={styles.legendColor} style={{ backgroundColor: colors[index % colors.length] }}></div>
							<div className={styles.legendText}>
								<div className={styles.typeName}>{type.type}</div>
								<div className={styles.typeStats}>
									{type.count.toLocaleString()} arquivos • {type.size}
								</div>
							</div>
						</div>
					))}
				</div>
			</div>
		</Card>
	);
}

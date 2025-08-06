import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Card from '../../ui/Card/Card';
import styles from './ActivityChart.module.css';

export default function ActivityChart() {
	const { analyticsData } = useAnalytics();
	const { activityChart } = analyticsData.recentActivity;

	const maxValue = Math.max(...activityChart.map((day) => Math.max(day.created, day.modified)));

	return (
		<Card title='Atividade por Dia (Últimos 7 dias)'>
			<div className={styles.chart}>
				<div className={styles.chartArea}>
					{activityChart.map((day) => (
						<div key={day.date} className={styles.dayColumn}>
							<div className={styles.bars}>
								<div
									className={`${styles.bar} ${styles.created}`}
									style={{ height: `${(day.created / maxValue) * 100}%` }}
									title={`Criados: ${day.created}`}
								></div>
								<div
									className={`${styles.bar} ${styles.modified}`}
									style={{ height: `${(day.modified / maxValue) * 100}%` }}
									title={`Modificados: ${day.modified}`}
								></div>
							</div>
							<div className={styles.dayLabel}>
								{new Date(day.date).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })}
							</div>
						</div>
					))}
				</div>
				<div className={styles.legend}>
					<div className={styles.legendItem}>
						<div className={`${styles.legendColor} ${styles.createdColor}`}></div>
						<span>Criados</span>
					</div>
					<div className={styles.legendItem}>
						<div className={`${styles.legendColor} ${styles.modifiedColor}`}></div>
						<span>Modificados</span>
					</div>
				</div>
			</div>
		</Card>
	);
}

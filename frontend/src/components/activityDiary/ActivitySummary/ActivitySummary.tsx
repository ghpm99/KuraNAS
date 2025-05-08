import { useActivityDiary } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDuration } from '@/utils';
import styles from './ActivitySummary.module.css';
import Card from '@/components/ui/Card/Card';

const ActivitySummary = () => {
	const { data } = useActivityDiary();
	return (
		<Card title='Resumo do Dia'>
			<div className={styles['summary-grid']}>
				<div className={styles['summary-item']}>
					<h3>Total de Atividades</h3>
					<p className={styles['summary-value']}>{data?.summary?.total_activities}</p>
				</div>
				<div className={styles['summary-item']}>
					<h3>Tempo Total Trabalhado</h3>
					<p className={styles['summary-value']}>{formatDuration(data?.summary?.total_time_spent_seconds)}</p>
				</div>
				<div className={styles['summary-item']}>
					<h3>Atividade Mais Longa</h3>
					<p className={styles['summary-value']}>{data?.summary?.longest_activity?.name}</p>
					<p className={styles['summary-subvalue']}>
						{formatDuration(data?.summary?.longest_activity?.duration_seconds)}
					</p>
				</div>
			</div>
		</Card>
	);
};

export default ActivitySummary;

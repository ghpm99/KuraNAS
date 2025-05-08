import { useActivityDiary } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDate, formatDuration } from '@/utils';
import styles from './list.module.css';
import Card from '@/components/ui/Card/Card';

const ActivityList = () => {
	const { data, getCurrentDuration } = useActivityDiary();
	return (
		<Card title='Atividades Registradas' className={styles['content']}>
			{data?.entries?.length === 0 ? (
				<p className={styles.noActivities}>Nenhuma atividade registrada ainda.</p>
			) : (
				<div className={styles.tableContainer}>
					<table className={styles.table}>
						<thead>
							<tr>
								<th>Nome</th>
								<th>Descrição</th>
								<th>Início</th>
								<th>Fim</th>
								<th>Duração</th>
							</tr>
						</thead>
						<tbody>
							{data?.entries?.map((activity) => (
								<tr key={activity.id} className={activity.end_time === null ? styles.activeRow : ''}>
									<td>{activity.name}</td>
									<td>{activity.description || '-'}</td>
									<td>{formatDate(activity.start_time)}</td>
									<td>{activity.end_time ? formatDate(activity.end_time) : 'Em andamento'}</td>
									<td>
										{activity.end_time
											? activity.duration_formatted
											: formatDuration(getCurrentDuration(activity.start_time))}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}
		</Card>
	);
};

export default ActivityList;

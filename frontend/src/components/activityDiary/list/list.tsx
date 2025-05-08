import { useActivityDiary } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDate, formatDuration } from '@/utils';
import styles from './list.module.css';

const List = () => {
	const { data, getCurrentDuration } = useActivityDiary();
	return (
		<div className={styles['activity-list-card']}>
			<h2 className={styles['card-title']}>Atividades Registradas</h2>
			{data?.entries?.length === 0 ? (
				<p className={styles['no-activities']}>Nenhuma atividade registrada ainda.</p>
			) : (
				<div className={styles['table-container']}>
					<table className={styles['activity-table']}>
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
								<tr key={activity.id} className={activity.end_time === null ? styles['active-row'] : ''}>
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
		</div>
	);
};

export default List;

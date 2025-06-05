import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDate, formatDuration } from '@/utils';
import styles from './list.module.css';
import Card from '@/components/ui/Card/Card';
import { Copy } from 'lucide-react';

const ActivityList = () => {
	const { data, getCurrentDuration, copyActivity } = useActivityDiary();

	return (
		<Card title='Atividades Registradas' className={styles['content']}>
			{data?.entries?.items.length === 0 ? (
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
								<th>Ação</th>
							</tr>
						</thead>
						<tbody>
							{data?.entries?.items.map((activity) => (
								<tr key={activity.id} className={activity.end_time === null ? styles.activeRow : ''}>
									<td>{activity.name}</td>
									<td>{activity.description || '-'}</td>
									<td>{formatDate(activity.start_time)}</td>
									<td>{activity.end_time.HasValue ? formatDate(activity.end_time.Value) : 'Em andamento'}</td>
									<td>
										{activity.end_time.HasValue
											? formatDuration(activity.duration)
											: formatDuration(getCurrentDuration(activity.start_time))}
									</td>
									<td>
										<div onClick={() => copyActivity(activity)} className={styles.copyButton}>
											<Copy />
										</div>
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

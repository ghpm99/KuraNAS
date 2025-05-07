import { ActivityDiaryData } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDuration } from '@/utils';

const List = (activities: ActivityDiaryData[]) => {
	return (
		<div className='activity-list-card'>
			<h2 className='card-title'>Atividades Registradas</h2>
			{activities.length === 0 ? (
				<p className='no-activities'>Nenhuma atividade registrada ainda.</p>
			) : (
				<div className='table-container'>
					<table className='activity-table'>
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
							{activities.map((activity) => (
								<tr key={activity.id} className={activity.end_time === null ? 'active-row' : ''}>
									<td>{activity.name}</td>
									<td>{activity.description || '-'}</td>
									<td>{formatDateTime(activity.startTime)}</td>
									<td>{activity.endTime ? formatDateTime(activity.endTime) : 'Em andamento'}</td>
									<td>
										{activity.endTime
											? formatDuration(activity.duration)
											: formatDuration(getCurrentDuration(activity.startTime))}
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

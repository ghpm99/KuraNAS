import { ActivityDiarySummary } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDuration } from '@/utils';

const Summary = ({ total_activities, total_time_spent_seconds, longest_activity }: ActivityDiarySummary) => {
	return (
		<div className='activity-summary-card'>
			<h2 className='card-title'>Resumo do Dia</h2>
			<div className='summary-grid'>
				<div className='summary-item'>
					<h3>Total de Atividades</h3>
					<p className='summary-value'>{total_activities}</p>
				</div>
				<div className='summary-item'>
					<h3>Tempo Total Trabalhado</h3>
					<p className='summary-value'>{formatDuration(total_time_spent_seconds)}</p>
				</div>
				<div className='summary-item'>
					<h3>Atividade Mais Longa</h3>
					<p className='summary-value'>{longest_activity?.name}</p>
					<p className='summary-subvalue'>{formatDuration(longest_activity?.duration_seconds)}</p>
				</div>
			</div>
		</div>
	);
};

export default Summary;

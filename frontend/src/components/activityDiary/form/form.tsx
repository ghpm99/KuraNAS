import { Plus } from 'lucide-react';

const ActivityDiaryForm = ({
	name,
	description,
	setName,
	setDescription,
	submit,
}: {
	name: string;
	description: string;
	setName: (name: string) => void;
	setDescription: (description: string) => void;
	submit: () => void;
}) => {
	return (
		<div className='activity-form-card'>
			<h2 className='card-title'>Nova Atividade</h2>
			<div className='activity-form'>
				<div className='form-group'>
					<label htmlFor='activity-name'>Nome da Atividade *</label>
					<input
						type='text'
						id='activity-name'
						value={name}
						onChange={(e) => setName(e.target.value)}
						placeholder='Ex: Estudar, Reunião com cliente'
						required
					/>
				</div>

				<div className='form-group'>
					<label htmlFor='activity-description'>Descrição</label>
					<textarea
						id='activity-description'
						value={description}
						onChange={(e) => setDescription(e.target.value)}
						placeholder='Ex: Estudando React, Chamando cliente sobre projeto X'
						rows={3}
					/>
				</div>

				<button className='button primary-button' onClick={submit}>
					<Plus className='icon' />
					Adicionar Atividade
				</button>
			</div>
		</div>
	);
};

export default ActivityDiaryForm;

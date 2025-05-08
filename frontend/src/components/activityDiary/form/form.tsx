import { useActivityDiary } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';
import { Plus } from 'lucide-react';
import styles from './form.module.css';

const ActivityDiaryForm = () => {
	const {
		form: { name, description },
		setForm,
		submitForm,
	} = useActivityDiary();
	return (
		<div className={styles['activity-form-card']}>
			<h2 className={styles['card-title']}>Nova Atividade</h2>
			<div className={styles['activity-form']}>
				<div className={styles['form-group']}>
					<label htmlFor='activity-name'>Nome da Atividade *</label>
					<input
						type='text'
						id='activity-name'
						value={name}
						onChange={(e) => setForm({ type: 'SET_NAME', payload: e.target.value })}
						placeholder='Ex: Estudar, Reunião com cliente'
						required
					/>
				</div>

				<div className={styles['form-group']}>
					<label htmlFor='activity-description'>Descrição</label>
					<textarea
						id='activity-description'
						value={description}
						onChange={(e) => setForm({ type: 'SET_DESCRIPTION', payload: e.target.value })}
						placeholder='Ex: Estudando React, Chamando cliente sobre projeto X'
						rows={3}
					/>
				</div>

				<button className={styles['button primary-button']} onClick={submitForm}>
					<Plus className={styles['icon']} />
					Adicionar Atividade
				</button>
			</div>
		</div>
	);
};

export default ActivityDiaryForm;

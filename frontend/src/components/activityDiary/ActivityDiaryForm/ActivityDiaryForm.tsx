import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import Button from '@/components/ui/Button/Button';
import Card from '@/components/ui/Card/Card';
import { Plus } from 'lucide-react';
import styles from './form.module.css';

const ActivityDiaryForm = () => {
	const { form, handleSubmit, handleNameChange, handleDescriptionChange } = useActivityDiary();

	return (
		<Card title='Nova Atividade'>
			<form className={styles.form} onSubmit={handleSubmit}>
				<div className={styles.formGroup}>
					<label htmlFor='activity-name' className={styles.label}>
						Nome da Atividade *
					</label>
					<input
						type='text'
						id='activity-name'
						value={form.name}
						onChange={handleNameChange}
						placeholder='Ex: Estudar, Reunião com cliente'
						className={styles.input}
						required
					/>
				</div>

				<div className={styles.formGroup}>
					<label htmlFor='activity-description' className={styles.label}>
						Descrição
					</label>
					<textarea
						id='activity-description'
						value={form.description}
						onChange={handleDescriptionChange}
						placeholder='Ex: Estudando React, Chamando cliente sobre projeto X'
						className={styles.textarea}
						rows={3}
					/>
				</div>

				<Button type='submit' icon={Plus}>
					Adicionar Atividade
				</Button>
			</form>
		</Card>
	);
};

export default ActivityDiaryForm;

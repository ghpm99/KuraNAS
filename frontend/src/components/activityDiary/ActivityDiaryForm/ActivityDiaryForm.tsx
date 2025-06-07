import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import Button from '@/components/ui/Button/Button';
import Card from '@/components/ui/Card/Card';
import { Plus } from 'lucide-react';
import styles from './form.module.css';
import useI18n from '@/components/i18n/provider/i18nContext';

const ActivityDiaryForm = () => {
	const { form, handleSubmit, handleNameChange, handleDescriptionChange } = useActivityDiary();
	const { t } = useI18n();

	return (
		<Card title={t('NEW_ACTIVITY_TITLE')}>
			<form className={styles.form} onSubmit={handleSubmit}>
				<div className={styles.formGroup}>
					<label htmlFor='activity-name' className={styles.label}>
						{t('ACTIVITY_NAME_LABEL')}
					</label>
					<input
						type='text'
						id='activity-name'
						value={form.name}
						onChange={handleNameChange}
						placeholder={t('ACTIVITY_NAME_PLACEHOLDER')}
						className={styles.input}
						required
					/>
				</div>

				<div className={styles.formGroup}>
					<label htmlFor='activity-description' className={styles.label}>
						{t('ACTIVITY_DESCRIPTION_LABEL')}
					</label>
					<textarea
						id='activity-description'
						value={form.description}
						onChange={handleDescriptionChange}
						placeholder={t('ACTIVITY_DESCRIPTION_PLACEHOLDER')}
						className={styles.textarea}
						rows={3}
					/>
				</div>

				<Button type='submit' icon={Plus}>
					{t('ADD_ACTIVITY')}
				</Button>
			</form>
		</Card>
	);
};

export default ActivityDiaryForm;

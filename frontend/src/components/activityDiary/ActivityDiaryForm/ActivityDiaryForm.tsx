import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import Card from '@/components/ui/Card/Card';
import { Box, Button, TextField } from '@mui/material';
import { Plus } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ChangeEvent } from 'react';

const ActivityDiaryForm = () => {
	const { form, handleSubmit, handleNameChange, handleDescriptionChange } = useActivityDiary();
	const { t } = useI18n();

	return (
		<Card title={t('NEW_ACTIVITY_TITLE')}>
			<Box component='form' onSubmit={handleSubmit} sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
				<TextField
					id='activity-name'
					label={t('ACTIVITY_NAME_LABEL')}
					value={form.name}
					onChange={handleNameChange as (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void}
					placeholder={t('ACTIVITY_NAME_PLACEHOLDER')}
					required
					size='small'
					fullWidth
				/>
				<TextField
					id='activity-description'
					label={t('ACTIVITY_DESCRIPTION_LABEL')}
					value={form.description}
					onChange={handleDescriptionChange as (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void}
					placeholder={t('ACTIVITY_DESCRIPTION_PLACEHOLDER')}
					multiline
					rows={3}
					size='small'
					fullWidth
				/>
				<Button type='submit' variant='contained' startIcon={<Plus size={16} />} sx={{ alignSelf: 'flex-start' }}>
					{t('ADD_ACTIVITY')}
				</Button>
			</Box>
		</Card>
	);
};

export default ActivityDiaryForm;

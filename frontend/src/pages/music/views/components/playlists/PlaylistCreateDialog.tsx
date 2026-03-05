import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type PlaylistCreateDialogProps = {
	open: boolean;
	newName: string;
	newDescription: string;
	isSubmitting: boolean;
	onClose: () => void;
	onNameChange: (value: string) => void;
	onDescriptionChange: (value: string) => void;
	onSubmit: () => void;
};

export default function PlaylistCreateDialog({
	open,
	newName,
	newDescription,
	isSubmitting,
	onClose,
	onNameChange,
	onDescriptionChange,
	onSubmit,
}: PlaylistCreateDialogProps) {
	const { t } = useI18n();

	return (
		<Dialog open={open} onClose={onClose} maxWidth='sm' fullWidth>
			<DialogTitle>{t('MUSIC_CREATE_PLAYLIST')}</DialogTitle>
			<DialogContent>
				<TextField
					autoFocus
					fullWidth
					label={t('NAME')}
					value={newName}
					onChange={(event) => onNameChange(event.target.value)}
					sx={{ mt: 1, mb: 2 }}
				/>
				<TextField
					fullWidth
					label={t('MUSIC_DESCRIPTION_OPTIONAL')}
					value={newDescription}
					onChange={(event) => onDescriptionChange(event.target.value)}
					multiline
					rows={2}
				/>
			</DialogContent>
			<DialogActions>
				<Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
				<Button variant='contained' onClick={onSubmit} disabled={!newName.trim() || isSubmitting}>
					{isSubmitting ? <CircularProgress size={20} /> : t('ACTION_CREATE')}
				</Button>
			</DialogActions>
		</Dialog>
	);
}

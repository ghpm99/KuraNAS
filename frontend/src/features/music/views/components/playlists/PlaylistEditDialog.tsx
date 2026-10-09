import { useState } from 'react';
import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type PlaylistEditDialogProps = {
    currentName: string;
    currentDescription: string;
    isSubmitting: boolean;
    onClose: () => void;
    onSubmit: (name: string, description: string) => void;
};

export default function PlaylistEditDialog({
    currentName,
    currentDescription,
    isSubmitting,
    onClose,
    onSubmit,
}: PlaylistEditDialogProps) {
    const { t } = useI18n();
    const [draftName, setDraftName] = useState(currentName);
    const [draftDescription, setDraftDescription] = useState(currentDescription);

    return (
        <Dialog open onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('MUSIC_PLAYLIST_EDIT')}</DialogTitle>
            <DialogContent>
                <TextField
                    autoFocus
                    fullWidth
                    label={t('NAME')}
                    value={draftName}
                    onChange={(event) => setDraftName(event.target.value)}
                    sx={{ mt: 1, mb: 2 }}
                />
                <TextField
                    fullWidth
                    label={t('MUSIC_DESCRIPTION_OPTIONAL')}
                    value={draftDescription}
                    onChange={(event) => setDraftDescription(event.target.value)}
                    multiline
                    rows={2}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button
                    variant="contained"
                    onClick={() => onSubmit(draftName.trim(), draftDescription.trim())}
                    disabled={!draftName.trim() || isSubmitting}
                >
                    {isSubmitting ? <CircularProgress size={20} /> : t('MUSIC_PLAYLIST_SAVE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}

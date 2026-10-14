import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
} from '@mui/material';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';

type VideoPlaylistNameDialogProps = {
    isOpen: boolean;
    isSubmitting: boolean;
    confirmLabelKey: string;
    onClose: () => void;
    onSubmit: (name: string) => void;
};

export default function VideoPlaylistNameDialog({
    isOpen,
    isSubmitting,
    confirmLabelKey,
    onClose,
    onSubmit,
}: VideoPlaylistNameDialogProps) {
    const { t } = useI18n();
    const [name, setName] = useState('');
    const trimmedName = name.trim();
    const canSubmit = trimmedName !== '' && !isSubmitting;

    const closeDialog = () => {
        setName('');
        onClose();
    };

    const submit = () => {
        if (!canSubmit) return;
        onSubmit(trimmedName);
        setName('');
    };

    return (
        <Dialog open={isOpen} onClose={closeDialog} maxWidth="sm" fullWidth>
            <DialogTitle>{t('VIDEO_NEW_PLAYLIST')}</DialogTitle>
            <DialogContent>
                <TextField
                    autoFocus
                    fullWidth
                    label={t('VIDEO_NEW_PLAYLIST_NAME')}
                    value={name}
                    onChange={(event) => setName(event.target.value)}
                    onKeyDown={(event) => {
                        if (event.key === 'Enter') submit();
                    }}
                    sx={{ mt: 1 }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={closeDialog}>{t('ACTION_CANCEL')}</Button>
                <Button variant="contained" onClick={submit} disabled={!canSubmit}>
                    {isSubmitting ? <CircularProgress size={20} /> : t(confirmLabelKey)}
                </Button>
            </DialogActions>
        </Dialog>
    );
}

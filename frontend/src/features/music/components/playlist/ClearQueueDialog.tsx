import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type ClearQueueDialogProps = {
    isOpen: boolean;
    onConfirm: () => void;
    onCancel: () => void;
};

export default function ClearQueueDialog({ isOpen, onConfirm, onCancel }: ClearQueueDialogProps) {
    const { t } = useI18n();

    return (
        <Dialog open={isOpen} onClose={onCancel} maxWidth="xs" fullWidth>
            <DialogTitle>{t('MUSIC_QUEUE_CLEAR_CONFIRM_TITLE')}</DialogTitle>
            <DialogContent>
                <DialogContentText>{t('MUSIC_QUEUE_CLEAR_CONFIRM_MESSAGE')}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={onCancel}>{t('ACTION_CANCEL')}</Button>
                <Button color="error" variant="contained" onClick={onConfirm}>
                    {t('MUSIC_QUEUE_CLEAR')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}

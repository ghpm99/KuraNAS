import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type VideoConfirmDialogProps = {
    isOpen: boolean;
    title: string;
    message: string;
    confirmLabel: string;
    onConfirm: () => void;
    onClose: () => void;
};

export default function VideoConfirmDialog({
    isOpen,
    title,
    message,
    confirmLabel,
    onConfirm,
    onClose,
}: VideoConfirmDialogProps) {
    const { t } = useI18n();

    return (
        <Dialog open={isOpen} onClose={onClose} maxWidth="xs" fullWidth>
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <DialogContentText>{message}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button
                    variant="contained"
                    onClick={() => {
                        onClose();
                        onConfirm();
                    }}
                >
                    {confirmLabel}
                </Button>
            </DialogActions>
        </Dialog>
    );
}

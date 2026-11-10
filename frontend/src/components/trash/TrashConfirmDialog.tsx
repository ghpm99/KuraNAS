import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type TrashConfirmDialogProps = {
    isOpen: boolean;
    title: string;
    message: string;
    onCancel: () => void;
    onConfirm: () => void;
};

const TrashConfirmDialog = ({
    isOpen,
    title,
    message,
    onCancel,
    onConfirm,
}: TrashConfirmDialogProps) => {
    const { t } = useI18n();

    return (
        <Dialog open={isOpen} onClose={onCancel} maxWidth="xs" fullWidth>
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <Typography variant="body2">{message}</Typography>
            </DialogContent>
            <DialogActions>
                <Button onClick={onCancel}>{t('ACTION_CANCEL')}</Button>
                <Button onClick={onConfirm} variant="contained" color="error">
                    {title}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default TrashConfirmDialog;

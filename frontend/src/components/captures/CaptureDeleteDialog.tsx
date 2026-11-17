import useI18n from '@/components/i18n/provider/i18nContext';
import type { Capture } from '@/types/captures';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Typography,
} from '@mui/material';

type CaptureDeleteDialogProps = {
    capture: Capture | null;
    onCancel: () => void;
    onConfirm: () => void;
};

const isPromoted = (capture: Capture): boolean =>
    capture.status === 'promoted' || capture.file_id != null;

const CaptureDeleteDialog = ({ capture, onCancel, onConfirm }: CaptureDeleteDialogProps) => {
    const { t } = useI18n();
    const messageKey =
        capture && isPromoted(capture)
            ? 'CAPTURE_DELETE_PROMOTED_MESSAGE'
            : 'CAPTURE_DELETE_UNPROMOTED_MESSAGE';

    return (
        <Dialog open={capture !== null} onClose={onCancel} maxWidth="xs" fullWidth>
            <DialogTitle>{t('CAPTURE_DELETE_TITLE')}</DialogTitle>
            <DialogContent>
                <Typography variant="body2">{t(messageKey)}</Typography>
            </DialogContent>
            <DialogActions>
                <Button onClick={onCancel}>{t('ACTION_CANCEL')}</Button>
                <Button onClick={onConfirm} variant="contained" color="error">
                    {t('CAPTURES_DELETE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default CaptureDeleteDialog;

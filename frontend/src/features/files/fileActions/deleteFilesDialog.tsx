import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

type DeleteFilesDialogProps = {
    files: FileData[];
    isOpen: boolean;
    onClose: () => void;
    onConfirm: (files: FileData[]) => void;
};

const DeleteFilesDialog = ({ files, isOpen, onClose, onConfirm }: DeleteFilesDialogProps) => {
    const { t } = useI18n();
    const confirmationMessage =
        files.length > 1
            ? t('FILES_CONFIRM_DELETE_MANY', { count: String(files.length) })
            : t('CONFIRM_DELETE');

    return (
        <Dialog open={isOpen} onClose={onClose} maxWidth="xs" fullWidth>
            <DialogTitle>{t('DELETE')}</DialogTitle>
            <DialogContent>
                <Typography variant="body2">{confirmationMessage}</Typography>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button onClick={() => onConfirm(files)} variant="contained" color="error">
                    {t('DELETE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default DeleteFilesDialog;

import { useState } from 'react';
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

type RenameFileDialogProps = {
    file: FileData | null;
    onClose: () => void;
    onConfirm: (file: FileData, newName: string) => void;
};

const RenameFileDialog = ({ file, onClose, onConfirm }: RenameFileDialogProps) => {
    const { t } = useI18n();
    const [typedName, setTypedName] = useState(file?.name ?? '');
    const trimmedName = typedName.trim();
    const isUnchangedOrEmpty = trimmedName === '' || trimmedName === file?.name;

    const confirmRename = () => {
        if (!file || isUnchangedOrEmpty) return;
        onConfirm(file, trimmedName);
    };

    return (
        <Dialog open={file !== null} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('RENAME')}</DialogTitle>
            <DialogContent>
                <TextField
                    autoFocus
                    margin="dense"
                    label={t('NAME')}
                    fullWidth
                    value={typedName}
                    onChange={(event) => setTypedName(event.target.value)}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button onClick={confirmRename} variant="contained" disabled={isUnchangedOrEmpty}>
                    {t('RENAME')}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default RenameFileDialog;

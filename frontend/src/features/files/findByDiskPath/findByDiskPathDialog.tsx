import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { getFileByDiskPath } from '@/service/files';
import {
    Alert,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    TextField,
} from '@mui/material';
import { useState } from 'react';
import type { FormEvent } from 'react';

type FindByDiskPathDialogProps = {
    open: boolean;
    onClose: () => void;
    onFileFound?: (file: FileData) => void;
};

const extractBackendMessage = (error: unknown): string | undefined => {
    if (typeof error !== 'object' || error === null || !('response' in error)) return undefined;
    const response = (error as { response?: { data?: { error?: string } } }).response;
    return response?.data?.error;
};

const FindByDiskPathDialog = ({ open, onClose, onFileFound }: FindByDiskPathDialogProps) => {
    const { t } = useI18n();
    const [diskPath, setDiskPath] = useState('');
    const [errorMessage, setErrorMessage] = useState<string | null>(null);
    const [isSearching, setIsSearching] = useState(false);

    const closeDialog = () => {
        setDiskPath('');
        setErrorMessage(null);
        onClose();
    };

    const submitSearch = async (event: FormEvent) => {
        event.preventDefault();
        const trimmedPath = diskPath.trim();
        if (!trimmedPath) return;

        setIsSearching(true);
        setErrorMessage(null);
        try {
            const file = await getFileByDiskPath(trimmedPath);
            onFileFound?.(file);
            closeDialog();
        } catch (error) {
            setErrorMessage(extractBackendMessage(error) ?? t('FILES_FIND_BY_DISK_PATH_FAILED'));
        } finally {
            setIsSearching(false);
        }
    };

    return (
        <Dialog open={open} onClose={closeDialog} fullWidth maxWidth="sm">
            <form onSubmit={submitSearch}>
                <DialogTitle>{t('FILES_FIND_BY_DISK_PATH')}</DialogTitle>
                <DialogContent>
                    <DialogContentText sx={{ mb: 2 }}>
                        {t('FILES_FIND_BY_DISK_PATH_DESCRIPTION')}
                    </DialogContentText>
                    <TextField
                        autoFocus
                        fullWidth
                        size="small"
                        value={diskPath}
                        onChange={(event) => setDiskPath(event.target.value)}
                        label={t('FILES_FIND_BY_DISK_PATH_LABEL')}
                        slotProps={{ htmlInput: { style: { fontFamily: 'monospace' } } }}
                    />
                    {errorMessage ? (
                        <Alert severity="error" sx={{ mt: 2 }}>
                            {errorMessage}
                        </Alert>
                    ) : null}
                </DialogContent>
                <DialogActions>
                    <Button onClick={closeDialog}>{t('FILES_FIND_BY_DISK_PATH_CANCEL')}</Button>
                    <Button
                        type="submit"
                        variant="contained"
                        disabled={isSearching || !diskPath.trim()}
                    >
                        {t('FILES_FIND_BY_DISK_PATH_SUBMIT')}
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
    );
};

export default FindByDiskPathDialog;

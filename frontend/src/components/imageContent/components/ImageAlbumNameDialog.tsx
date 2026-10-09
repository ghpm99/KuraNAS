import { useState } from 'react';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type ImageAlbumNameDialogProps = {
    isOpen: boolean;
    title: string;
    confirmLabel: string;
    initialName?: string;
    onClose: () => void;
    onSubmit: (name: string) => void;
};

const ImageAlbumNameDialogBody = ({
    title,
    confirmLabel,
    initialName = '',
    onClose,
    onSubmit,
}: Omit<ImageAlbumNameDialogProps, 'isOpen'>) => {
    const { t } = useI18n();
    const [name, setName] = useState(initialName);
    const trimmedName = name.trim();

    const submitName = () => {
        if (trimmedName) {
            onSubmit(trimmedName);
        }
    };

    return (
        <>
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <TextField
                    autoFocus
                    fullWidth
                    margin="dense"
                    label={t('IMAGES_ALBUM_NAME_LABEL')}
                    value={name}
                    onChange={(event) => setName(event.target.value)}
                    onKeyDown={(event) => {
                        if (event.key === 'Enter') {
                            submitName();
                        }
                    }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button variant="contained" disabled={!trimmedName} onClick={submitName}>
                    {confirmLabel}
                </Button>
            </DialogActions>
        </>
    );
};

export default function ImageAlbumNameDialog({ isOpen, ...bodyProps }: ImageAlbumNameDialogProps) {
    return (
        <Dialog open={isOpen} onClose={bodyProps.onClose} fullWidth maxWidth="xs">
            {isOpen && <ImageAlbumNameDialogBody {...bodyProps} />}
        </Dialog>
    );
}

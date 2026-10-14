import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
} from '@mui/material';
import { Trash2 } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { type VideoPlaylistDto } from '@/service/videoPlayback';
import styles from '../videoContent.module.css';

type VideoDeletePlaylistButtonProps = {
    playlist: Pick<VideoPlaylistDto, 'name' | 'is_auto'>;
    isDeleting: boolean;
    onDelete?: () => void;
};

export default function VideoDeletePlaylistButton({
    playlist,
    isDeleting,
    onDelete,
}: VideoDeletePlaylistButtonProps) {
    const { t } = useI18n();
    const [isConfirmOpen, setIsConfirmOpen] = useState(false);

    if (playlist.is_auto !== false || !onDelete) return null;

    return (
        <>
            <button
                type="button"
                className={styles.actionBtn}
                disabled={isDeleting}
                onClick={() => setIsConfirmOpen(true)}
            >
                <Trash2 size={14} />
                {t('VIDEO_DELETE_PLAYLIST')}
            </button>
            <Dialog
                open={isConfirmOpen}
                onClose={() => setIsConfirmOpen(false)}
                maxWidth="xs"
                fullWidth
            >
                <DialogTitle>{t('VIDEO_DELETE_PLAYLIST_CONFIRM_TITLE')}</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        {t('VIDEO_DELETE_PLAYLIST_CONFIRM_MESSAGE', { name: playlist.name })}
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setIsConfirmOpen(false)}>{t('ACTION_CANCEL')}</Button>
                    <Button
                        color="error"
                        variant="contained"
                        disabled={isDeleting}
                        onClick={() => {
                            setIsConfirmOpen(false);
                            onDelete();
                        }}
                    >
                        {t('VIDEO_DELETE_PLAYLIST')}
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
}

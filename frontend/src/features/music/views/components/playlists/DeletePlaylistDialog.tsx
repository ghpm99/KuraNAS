import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type DeletePlaylistDialogProps = {
    playlistName: string | null;
    onConfirm: () => void;
    onCancel: () => void;
};

export default function DeletePlaylistDialog({
    playlistName,
    onConfirm,
    onCancel,
}: DeletePlaylistDialogProps) {
    const { t } = useI18n();

    return (
        <Dialog open={playlistName !== null} onClose={onCancel} maxWidth="xs" fullWidth>
            <DialogTitle>{t('MUSIC_PLAYLIST_DELETE_CONFIRM_TITLE')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('MUSIC_PLAYLIST_DELETE_CONFIRM_MESSAGE', { name: playlistName ?? '' })}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={onCancel}>{t('ACTION_CANCEL')}</Button>
                <Button color="error" variant="contained" onClick={onConfirm}>
                    {t('DELETE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}

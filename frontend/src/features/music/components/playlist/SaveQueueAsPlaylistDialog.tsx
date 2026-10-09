import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
} from '@mui/material';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { saveTracksAsPlaylist } from './saveTracksAsPlaylist';

type SaveQueueAsPlaylistDialogProps = {
    fileIds: number[];
    onClose: () => void;
};

export default function SaveQueueAsPlaylistDialog({
    fileIds,
    onClose,
}: SaveQueueAsPlaylistDialogProps) {
    const [playlistName, setPlaylistName] = useState('');
    const queryClient = useQueryClient();
    const { enqueueSnackbar } = useSnackbar();
    const { t } = useI18n();

    const saveMutation = useMutation({
        mutationFn: () => saveTracksAsPlaylist(playlistName.trim(), fileIds),
        onSuccess: (summary) => {
            queryClient.invalidateQueries({ queryKey: ['playlists'] });
            queryClient.invalidateQueries({ queryKey: ['playlists-menu'] });
            enqueueSnackbar(
                t('MUSIC_QUEUE_SAVED_AS_PLAYLIST', {
                    name: summary.playlistName,
                    added: String(summary.addedTracks),
                    total: String(summary.totalTracks),
                }),
                { variant: summary.addedTracks === summary.totalTracks ? 'success' : 'warning' }
            );
            onClose();
        },
        onError: () => {
            enqueueSnackbar(t('MUSIC_QUEUE_SAVE_FAILED'), { variant: 'error' });
        },
    });

    return (
        <Dialog open onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('MUSIC_QUEUE_SAVE_AS_PLAYLIST')}</DialogTitle>
            <DialogContent>
                <TextField
                    autoFocus
                    fullWidth
                    label={t('MUSIC_PLAYLIST_NAME')}
                    value={playlistName}
                    onChange={(event) => setPlaylistName(event.target.value)}
                    sx={{ mt: 1 }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button
                    variant="contained"
                    onClick={() => saveMutation.mutate()}
                    disabled={!playlistName.trim() || saveMutation.isPending}
                >
                    {saveMutation.isPending ? (
                        <CircularProgress size={20} />
                    ) : (
                        t('MUSIC_QUEUE_SAVE_AS_PLAYLIST_ACTION')
                    )}
                </Button>
            </DialogActions>
        </Dialog>
    );
}

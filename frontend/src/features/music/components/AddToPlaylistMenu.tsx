import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    TextField,
} from '@mui/material';
import { ListPlus, ListMusic, Plus } from 'lucide-react';
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getPlaylists, addTrackToPlaylist, createPlaylist } from '@/service/playlist';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';

interface AddToPlaylistMenuProps {
    fileId: number;
    anchorEl: HTMLElement | null;
    onClose: () => void;
}

const AddToPlaylistMenu = ({ fileId, anchorEl, onClose }: AddToPlaylistMenuProps) => {
    const [createOpen, setCreateOpen] = useState(false);
    const [newName, setNewName] = useState('');
    const queryClient = useQueryClient();
    const { enqueueSnackbar } = useSnackbar();
    const { t } = useI18n();

    const { data, isLoading } = useQuery({
        queryKey: ['playlists-menu'],
        queryFn: () => getPlaylists(1, 100),
        enabled: !!anchorEl,
    });

    const addMutation = useMutation({
        mutationFn: (playlistId: number) => addTrackToPlaylist(playlistId, fileId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['playlists'] });
            queryClient.invalidateQueries({ queryKey: ['playlist-tracks'] });
            enqueueSnackbar(t('MUSIC_TRACK_ADDED'), { variant: 'success' });
            onClose();
        },
        onError: () => {
            enqueueSnackbar(t('MUSIC_TRACK_ADD_FAILED'), { variant: 'warning' });
        },
    });

    const createAndAddMutation = useMutation({
        mutationFn: async () => {
            const playlist = await createPlaylist({ name: newName });
            await addTrackToPlaylist(playlist.id, fileId);
            return playlist;
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['playlists'] });
            queryClient.invalidateQueries({ queryKey: ['playlists-menu'] });
            setCreateOpen(false);
            setNewName('');
            enqueueSnackbar(t('MUSIC_PLAYLIST_CREATED_ADDED'), {
                variant: 'success',
            });
            onClose();
        },
        onError: () => {
            enqueueSnackbar(t('MUSIC_PLAYLIST_CREATE_FAILED'), { variant: 'error' });
        },
    });

    const playlists = data?.items?.filter((p) => !p.is_system) ?? [];
    const menuItems = [
        <MenuItem
            key="create"
            onClick={() => {
                setCreateOpen(true);
                onClose();
            }}
        >
            <ListItemIcon>
                <Plus size={18} />
            </ListItemIcon>
            <ListItemText primary={t('MUSIC_NEW_PLAYLIST')} />
        </MenuItem>,
        ...playlists.map((playlist) => (
            <MenuItem
                key={playlist.id}
                onClick={() => addMutation.mutate(playlist.id)}
                disabled={addMutation.isPending}
            >
                <ListItemIcon>
                    <ListMusic size={18} />
                </ListItemIcon>
                <ListItemText primary={playlist.name} />
            </MenuItem>
        )),
    ];

    if (playlists.length === 0) {
        menuItems.push(
            <MenuItem key="empty" disabled>
                <ListItemText primary={t('MUSIC_NO_PLAYLISTS')} />
            </MenuItem>
        );
    }

    return (
        <>
            <Menu anchorEl={anchorEl} open={!!anchorEl} onClose={onClose}>
                {isLoading ? (
                    <MenuItem disabled>
                        <CircularProgress size={20} sx={{ mr: 1 }} /> {t('LOADING')}
                    </MenuItem>
                ) : (
                    menuItems
                )}
            </Menu>

            <Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth="sm" fullWidth>
                <DialogTitle>{t('MUSIC_CREATE_PLAYLIST_ADD')}</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        fullWidth
                        label={t('MUSIC_PLAYLIST_NAME')}
                        value={newName}
                        onChange={(e) => setNewName(e.target.value)}
                        sx={{ mt: 1 }}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setCreateOpen(false)}>{t('ACTION_CANCEL')}</Button>
                    <Button
                        variant="contained"
                        onClick={() => createAndAddMutation.mutate()}
                        disabled={!newName.trim() || createAndAddMutation.isPending}
                    >
                        {createAndAddMutation.isPending ? (
                            <CircularProgress size={20} />
                        ) : (
                            t('ACTION_CREATE_ADD')
                        )}
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
};

export default AddToPlaylistMenu;

export const AddToPlaylistButton = ({ fileId }: { fileId: number }) => {
    const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

    return (
        <>
            <ListItemButton
                sx={{ px: 1, py: 0.5, borderRadius: 1, maxWidth: 'fit-content' }}
                onClick={(e) => {
                    e.stopPropagation();
                    setAnchorEl(e.currentTarget);
                }}
            >
                <ListItemIcon sx={{ minWidth: 28 }}>
                    <ListPlus size={16} />
                </ListItemIcon>
            </ListItemButton>
            <AddToPlaylistMenu
                fileId={fileId}
                anchorEl={anchorEl}
                onClose={() => setAnchorEl(null)}
            />
        </>
    );
};

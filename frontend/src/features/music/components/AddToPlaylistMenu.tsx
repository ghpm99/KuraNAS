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
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getPlaylists, addTrackToPlaylist, createPlaylist } from '@/service/playlist';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

const PLAYLIST_MENU_PAGE_SIZE = 30;
const PLAYLIST_SEARCH_DEBOUNCE_MS = 300;

interface AddToPlaylistMenuProps {
    fileId: number;
    anchorEl: HTMLElement | null;
    onClose: () => void;
}

const AddToPlaylistMenu = ({ fileId, anchorEl, onClose }: AddToPlaylistMenuProps) => {
    const [createOpen, setCreateOpen] = useState(false);
    const [newName, setNewName] = useState('');
    const [searchInput, setSearchInput] = useState('');
    const playlistSearch = useDebouncedValue(searchInput.trim(), PLAYLIST_SEARCH_DEBOUNCE_MS);
    const queryClient = useQueryClient();
    const { enqueueSnackbar } = useSnackbar();
    const { t } = useI18n();

    const closeMenu = () => {
        setSearchInput('');
        onClose();
    };

    const { data, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } = useInfiniteQuery({
        queryKey: ['playlists-menu', playlistSearch],
        queryFn: ({ pageParam }) =>
            getPlaylists(pageParam, PLAYLIST_MENU_PAGE_SIZE, playlistSearch),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
        enabled: !!anchorEl,
    });

    const addMutation = useMutation({
        mutationFn: (playlistId: number) => addTrackToPlaylist(playlistId, fileId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['playlists'] });
            queryClient.invalidateQueries({ queryKey: ['playlist-tracks'] });
            enqueueSnackbar(t('MUSIC_TRACK_ADDED'), { variant: 'success' });
            closeMenu();
        },
        onError: (error) => {
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('MUSIC_TRACK_ADD_FAILED'), {
                variant: 'warning',
            });
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
            closeMenu();
        },
        onError: () => {
            enqueueSnackbar(t('MUSIC_PLAYLIST_CREATE_FAILED'), { variant: 'error' });
        },
    });

    const playlists =
        data?.pages
            ?.flatMap((page) => page.items ?? [])
            .filter((playlist) => !playlist.is_system && !playlist.is_ai_generated) ?? [];
    const hasSearch = playlistSearch !== '';
    const menuItems = [
        <MenuItem
            key="create"
            onClick={() => {
                setCreateOpen(true);
                closeMenu();
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
                <ListItemText
                    primary={t(hasSearch ? 'MUSIC_PLAYLIST_SEARCH_EMPTY' : 'MUSIC_NO_PLAYLISTS')}
                />
            </MenuItem>
        );
    }

    return (
        <>
            <Menu
                anchorEl={anchorEl}
                open={!!anchorEl}
                onClose={closeMenu}
                slotProps={{ list: { autoFocusItem: false } }}
            >
                <TextField
                    key="search"
                    autoFocus
                    fullWidth
                    size="small"
                    placeholder={t('MUSIC_PLAYLIST_SEARCH')}
                    value={searchInput}
                    onChange={(event) => setSearchInput(event.target.value)}
                    onKeyDown={(event) => event.stopPropagation()}
                    slotProps={{ htmlInput: { 'aria-label': t('MUSIC_PLAYLIST_SEARCH') } }}
                    sx={{ px: 1.5, pb: 1 }}
                />
                {isLoading ? (
                    <MenuItem key="loading" disabled>
                        <CircularProgress size={20} sx={{ mr: 1 }} /> {t('LOADING')}
                    </MenuItem>
                ) : (
                    menuItems
                )}
                <LoadMoreSentinel
                    key="sentinel"
                    hasNextPage={hasNextPage}
                    isFetchingNextPage={isFetchingNextPage}
                    fetchNextPage={() => void fetchNextPage()}
                />
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

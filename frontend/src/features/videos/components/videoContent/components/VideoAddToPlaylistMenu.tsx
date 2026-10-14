import {
    CircularProgress,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    TextField,
} from '@mui/material';
import { Check, ListVideo, Plus } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';
import VideoPlaylistNameDialog from './VideoPlaylistNameDialog';
import {
    useVideoPlaylistSearch,
    useVideoPlaylistsOfVideo,
} from '../../../providers/videoContentProvider/useVideoQueries';

const PLAYLIST_MENU_PAGE_SIZE = 30;
const PLAYLIST_SEARCH_DEBOUNCE_MS = 300;

type VideoAddToPlaylistMenuProps = {
    videoId: number;
    anchorEl: HTMLElement | null;
    onClose: () => void;
    onSelectPlaylist: (playlistId: number) => void;
    onCreatePlaylist?: (name: string) => void;
    isCreatingPlaylist?: boolean;
};

export default function VideoAddToPlaylistMenu({
    videoId,
    anchorEl,
    onClose,
    onSelectPlaylist,
    onCreatePlaylist,
    isCreatingPlaylist = false,
}: VideoAddToPlaylistMenuProps) {
    const { t } = useI18n();
    const [searchInput, setSearchInput] = useState('');
    const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
    const playlistSearch = useDebouncedValue(searchInput.trim(), PLAYLIST_SEARCH_DEBOUNCE_MS);
    const isOpen = !!anchorEl;
    const { data, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } =
        useVideoPlaylistSearch(playlistSearch, PLAYLIST_MENU_PAGE_SIZE, isOpen);
    const { data: playlistsOfVideo = [] } = useVideoPlaylistsOfVideo(videoId, isOpen);

    const playlists = data?.pages?.flatMap((page) => page.items ?? []) ?? [];
    const playlistIdsContainingVideo = new Set(
        playlistsOfVideo.map((playlistOfVideo) => playlistOfVideo.id)
    );
    const hasSearch = playlistSearch !== '';

    const closeMenu = () => {
        setSearchInput('');
        onClose();
    };

    const openCreateDialog = () => {
        setIsCreateDialogOpen(true);
        closeMenu();
    };

    const renderCreatePlaylistItem = () => {
        if (!onCreatePlaylist) return null;
        return (
            <MenuItem onClick={openCreateDialog}>
                <ListItemIcon>
                    <Plus size={18} />
                </ListItemIcon>
                <ListItemText primary={t('VIDEO_NEW_PLAYLIST')} />
            </MenuItem>
        );
    };

    const renderPlaylistItems = () => {
        if (playlists.length === 0) {
            return (
                <MenuItem disabled>
                    <ListItemText
                        primary={t(
                            hasSearch ? 'VIDEO_PLAYLIST_SEARCH_EMPTY' : 'VIDEO_NO_PLAYLISTS'
                        )}
                    />
                </MenuItem>
            );
        }
        return playlists.map((playlist) => {
            const isAlreadyInPlaylist = playlistIdsContainingVideo.has(playlist.id);
            return (
                <MenuItem
                    key={playlist.id}
                    disabled={isAlreadyInPlaylist}
                    onClick={() => {
                        if (isAlreadyInPlaylist) return;
                        setSearchInput('');
                        onSelectPlaylist(playlist.id);
                    }}
                >
                    <ListItemIcon>
                        {isAlreadyInPlaylist ? <Check size={18} /> : <ListVideo size={18} />}
                    </ListItemIcon>
                    <ListItemText
                        primary={playlist.name}
                        secondary={isAlreadyInPlaylist ? t('VIDEO_ALREADY_ADDED') : undefined}
                    />
                </MenuItem>
            );
        });
    };

    return (
        <>
            <Menu
                anchorEl={anchorEl}
                open={isOpen}
                onClose={closeMenu}
                slotProps={{ list: { autoFocusItem: false } }}
            >
                <TextField
                    autoFocus
                    fullWidth
                    size="small"
                    placeholder={t('VIDEO_PLAYLIST_SEARCH')}
                    value={searchInput}
                    onChange={(event) => setSearchInput(event.target.value)}
                    onKeyDown={(event) => event.stopPropagation()}
                    slotProps={{ htmlInput: { 'aria-label': t('VIDEO_PLAYLIST_SEARCH') } }}
                    sx={{ px: 1.5, pb: 1 }}
                />
                {renderCreatePlaylistItem()}
                {isLoading ? (
                    <MenuItem disabled>
                        <CircularProgress size={20} sx={{ mr: 1 }} /> {t('LOADING')}
                    </MenuItem>
                ) : (
                    renderPlaylistItems()
                )}
                <LoadMoreSentinel
                    hasNextPage={hasNextPage}
                    isFetchingNextPage={isFetchingNextPage}
                    fetchNextPage={() => void fetchNextPage()}
                />
            </Menu>
            <VideoPlaylistNameDialog
                isOpen={isCreateDialogOpen}
                isSubmitting={isCreatingPlaylist}
                confirmLabelKey="ACTION_CREATE_ADD"
                onClose={() => setIsCreateDialogOpen(false)}
                onSubmit={(name) => {
                    setIsCreateDialogOpen(false);
                    onCreatePlaylist?.(name);
                }}
            />
        </>
    );
}

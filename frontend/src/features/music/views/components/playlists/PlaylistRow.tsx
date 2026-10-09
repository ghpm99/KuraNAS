import {
    Box,
    IconButton,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
} from '@mui/material';
import { ListMusic, Play, Trash2 } from 'lucide-react';
import CollectionContextMenu from '@/features/music/components/contextMenu/CollectionContextMenu';
import { createPlaylistPlaybackContext } from '@/features/music/components/playbackContext';
import { queueToTracks } from '@/features/music/components/musicQueueTracks';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { getPlaylistQueue } from '@/service/playlist';
import { Playlist } from '@/types/playlist';
import useI18n from '@/components/i18n/provider/i18nContext';

const loadPlaylistTracks = (playlistId: number) => getPlaylistQueue(playlistId).then(queueToTracks);

type PlaylistRowProps = {
    playlist: Playlist;
    canDelete: boolean;
    onSelect: (playlist: Playlist) => void;
    onDeleteRequest: (playlist: Playlist) => void;
};

export default function PlaylistRow({
    playlist,
    canDelete,
    onSelect,
    onDeleteRequest,
}: PlaylistRowProps) {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();

    const handleKeyDown = (event: React.KeyboardEvent<HTMLElement>) => {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            onSelect(playlist);
        }
    };

    const handlePlayPlaylist = async (event: React.MouseEvent) => {
        event.stopPropagation();
        const tracks = await loadPlaylistTracks(playlist.id);
        if (tracks.length > 0) replaceQueue(tracks, 0, createPlaylistPlaybackContext(playlist));
    };

    return (
        <CollectionContextMenu
            collectionName={playlist.name}
            playbackContext={createPlaylistPlaybackContext(playlist)}
            loadTracks={() => loadPlaylistTracks(playlist.id)}
            layout="row"
        >
            <ListItem disablePadding sx={{ '&:hover .playlist-actions': { opacity: 1 } }}>
                <ListItemButton
                    component="div"
                    role="button"
                    tabIndex={0}
                    onClick={() => onSelect(playlist)}
                    onKeyDown={handleKeyDown}
                    sx={{ borderRadius: 1.5, py: 1, pl: 1.5, pr: 6, gap: 1 }}
                >
                    <ListItemIcon sx={{ minWidth: 40 }}>
                        <Box
                            sx={{
                                width: 40,
                                height: 40,
                                borderRadius: 1,
                                bgcolor: playlist.is_system
                                    ? 'rgba(167, 139, 250, 0.15)'
                                    : 'rgba(var(--app-color-primary-rgb), 0.12)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}
                        >
                            <ListMusic
                                size={20}
                                color={playlist.is_system ? '#a78bfa' : '#6366f1'}
                            />
                        </Box>
                    </ListItemIcon>
                    <ListItemText
                        primary={playlist.name}
                        secondary={`${playlist.track_count} ${t('MUSIC_TRACKS_COUNT')}${playlist.description ? ` · ${playlist.description}` : ''}`}
                        primaryTypographyProps={{ fontWeight: 500 }}
                    />
                    <Box
                        className="playlist-actions"
                        sx={{
                            display: 'flex',
                            gap: 0.5,
                            opacity: 0,
                            transition: 'opacity 0.2s ease',
                        }}
                    >
                        <IconButton
                            size="small"
                            onClick={handlePlayPlaylist}
                            sx={{ color: 'primary.main' }}
                        >
                            <Play size={16} fill="var(--app-color-primary)" />
                        </IconButton>
                        {canDelete && (
                            <IconButton
                                size="small"
                                aria-label={t('DELETE')}
                                onClick={(event) => {
                                    event.stopPropagation();
                                    onDeleteRequest(playlist);
                                }}
                                sx={{
                                    color: 'text.secondary',
                                    '&:hover': { color: 'error.main' },
                                }}
                            >
                                <Trash2 size={16} />
                            </IconButton>
                        )}
                    </Box>
                </ListItemButton>
            </ListItem>
        </CollectionContextMenu>
    );
}

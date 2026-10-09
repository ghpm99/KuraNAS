import {
    Box,
    CircularProgress,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    Typography,
} from '@mui/material';
import { useState } from 'react';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import { GripVertical, ListMusic, Pause, Pencil, Play, Trash2 } from 'lucide-react';
import { createPlaylistPlaybackContext } from '@/features/music/components/playbackContext';
import { Playlist, PlaylistTrack } from '@/types/playlist';
import useI18n from '@/components/i18n/provider/i18nContext';
import { usePlaylistTrackHandlers } from '@/features/music/hooks/usePlaylistTrackHandlers/usePlaylistTrackHandlers';
import CategoryHeader from '@/features/music/components/CategoryHeader';
import { formatMusicDuration, getTrackDurationSeconds } from '@/utils/music';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { getPlaylistQueue } from '@/service/playlist';
import { findStartIndex, queueToTracks } from '@/features/music/components/musicQueueTracks';
import { shuffleItems } from '@/utils/shuffleItems';
import PlaylistEditDialog from './PlaylistEditDialog';
import PlaylistTrackMoveButtons from './PlaylistTrackMoveButtons';

type PlaylistDetailSectionProps = {
    playlist: Playlist;
    tracks: PlaylistTrack[];
    isLoading: boolean;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onBack: () => void;
    onRemoveTrack: (fileId: number) => void;
    onLoadMore: () => void;
    isRenaming?: boolean;
    onRenamePlaylist?: (name: string, description: string, onSaved: () => void) => void;
    onMoveTrack?: (fileId: number, targetPosition: number) => void;
};

export default function PlaylistDetailSection({
    playlist,
    tracks,
    isLoading,
    hasNextPage,
    isFetchingNextPage,
    onBack,
    onRemoveTrack,
    onLoadMore,
    isRenaming = false,
    onRenamePlaylist,
    onMoveTrack,
}: PlaylistDetailSectionProps) {
    const { t } = useI18n();
    const [isEditOpen, setIsEditOpen] = useState(false);
    const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
    const { getMusicArtist, getMusicTitle } = usePlaylistTrackHandlers();
    const { currentTrack, isPlaying, replaceQueue } = useGlobalMusic();
    const handleListItemKeyDown = (
        event: React.KeyboardEvent<HTMLElement>,
        onActivate: () => void
    ) => {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            onActivate();
        }
    };

    const playbackContext = createPlaylistPlaybackContext(playlist);
    const canRemoveTracks = !playlist.is_system && !playlist.is_auto;
    const isUserPlaylist = canRemoveTracks && !playlist.is_ai_generated;
    const canRename = isUserPlaylist && onRenamePlaylist !== undefined;
    const canReorder = isUserPlaylist && onMoveTrack !== undefined;

    const moveTrackToIndex = (fileId: number, fromIndex: number, toIndex: number) => {
        if (!onMoveTrack || fromIndex === toIndex) return;
        onMoveTrack(fileId, toIndex + 1);
    };

    const handleDrop = (targetIndex: number) => {
        const sourceIndex = draggedIndex;
        setDraggedIndex(null);
        const draggedTrack = sourceIndex === null ? undefined : tracks[sourceIndex];
        if (sourceIndex === null || !draggedTrack) return;
        moveTrackToIndex(draggedTrack.file.id, sourceIndex, targetIndex);
    };

    const startPlaylistQueue = async (startFileId?: number, shouldShuffle = false) => {
        const queueTracks = queueToTracks(await getPlaylistQueue(playlist.id));
        if (queueTracks.length === 0) return;

        if (shouldShuffle) {
            replaceQueue(shuffleItems(queueTracks), 0, playbackContext);
            return;
        }
        replaceQueue(queueTracks, findStartIndex(queueTracks, startFileId), playbackContext);
    };

    const handlePlayAll = () => void startPlaylistQueue();
    const handleShuffleAll = () => void startPlaylistQueue(undefined, true);

    return (
        <Box sx={{ p: 2 }}>
            <CategoryHeader
                title={playlist.name}
                subtitle={playlist.description || undefined}
                trackCount={playlist.track_count}
                icon={<ListMusic size={48} opacity={0.7} />}
                gradientFrom="var(--app-color-primary)"
                onBack={onBack}
                onPlayAll={handlePlayAll}
                onShuffleAll={handleShuffleAll}
                actions={
                    canRename ? (
                        <IconButton
                            aria-label={t('MUSIC_PLAYLIST_EDIT')}
                            onClick={() => setIsEditOpen(true)}
                            sx={{ color: 'text.secondary', '&:hover': { color: 'text.primary' } }}
                        >
                            <Pencil size={18} />
                        </IconButton>
                    ) : undefined
                }
            />

            {canRename && isEditOpen && (
                <PlaylistEditDialog
                    currentName={playlist.name}
                    currentDescription={playlist.description ?? ''}
                    isSubmitting={isRenaming}
                    onClose={() => setIsEditOpen(false)}
                    onSubmit={(name, description) =>
                        onRenamePlaylist?.(name, description, () => setIsEditOpen(false))
                    }
                />
            )}

            {isLoading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                    <CircularProgress />
                </Box>
            ) : tracks.length === 0 ? (
                <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ textAlign: 'center', p: 4 }}
                >
                    {t('MUSIC_PLAYLIST_EMPTY')}
                </Typography>
            ) : (
                <List sx={{ width: '100%' }}>
                    {tracks.map((track, index) => {
                        const isCurrentTrack = currentTrack?.id === track.file.id;
                        const duration = getTrackDurationSeconds(track.file.metadata);

                        return (
                            <ListItem
                                key={track.id}
                                disablePadding
                                draggable={canReorder}
                                onDragStart={() => setDraggedIndex(index)}
                                onDragEnd={() => setDraggedIndex(null)}
                                onDragOver={(event) => {
                                    if (draggedIndex !== null) event.preventDefault();
                                }}
                                onDrop={(event) => {
                                    event.preventDefault();
                                    handleDrop(index);
                                }}
                                sx={{
                                    opacity: draggedIndex === index ? 0.5 : 1,
                                    '&:hover .remove-btn': { opacity: 1 },
                                }}
                            >
                                <ListItemButton
                                    component="div"
                                    role="button"
                                    tabIndex={0}
                                    onClick={() => void startPlaylistQueue(track.file.id)}
                                    onKeyDown={(event) =>
                                        handleListItemKeyDown(
                                            event,
                                            () => void startPlaylistQueue(track.file.id)
                                        )
                                    }
                                    sx={{
                                        borderRadius: 1,
                                        py: 0.5,
                                        px: 1,
                                        gap: 1.5,
                                        bgcolor: isCurrentTrack
                                            ? 'rgba(var(--app-color-primary-rgb), 0.08)'
                                            : 'transparent',
                                        '&:hover': {
                                            bgcolor: isCurrentTrack
                                                ? 'rgba(var(--app-color-primary-rgb), 0.12)'
                                                : undefined,
                                        },
                                        '&:hover .track-index': { display: 'none' },
                                        '&:hover .track-play-icon': { display: 'flex' },
                                    }}
                                >
                                    {canReorder && (
                                        <GripVertical
                                            size={14}
                                            aria-hidden
                                            style={{ cursor: 'grab', flexShrink: 0 }}
                                        />
                                    )}
                                    <Box
                                        sx={{
                                            width: 32,
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            flexShrink: 0,
                                        }}
                                    >
                                        {isCurrentTrack && isPlaying ? (
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    alignItems: 'flex-end',
                                                    gap: '2px',
                                                    height: 16,
                                                }}
                                            >
                                                {[1, 2, 3].map((bar) => (
                                                    <Box
                                                        key={bar}
                                                        sx={{
                                                            width: 3,
                                                            bgcolor: 'primary.main',
                                                            borderRadius: 1,
                                                            height: '10px',
                                                            animation: `equalizer ${0.4 + bar * 0.15}s ease-in-out infinite alternate`,
                                                            '@keyframes equalizer': {
                                                                '0%': { height: '4px' },
                                                                '100%': { height: '14px' },
                                                            },
                                                        }}
                                                    />
                                                ))}
                                            </Box>
                                        ) : isCurrentTrack ? (
                                            <Pause size={14} color="var(--app-color-primary)" />
                                        ) : (
                                            <>
                                                <Typography
                                                    className="track-index"
                                                    variant="body2"
                                                    color="text.secondary"
                                                    sx={{ fontVariantNumeric: 'tabular-nums' }}
                                                >
                                                    {index + 1}
                                                </Typography>
                                                <Box
                                                    className="track-play-icon"
                                                    sx={{ display: 'none', alignItems: 'center' }}
                                                >
                                                    <Play size={14} />
                                                </Box>
                                            </>
                                        )}
                                    </Box>

                                    <Box sx={{ flex: 1, minWidth: 0 }}>
                                        <Typography
                                            variant="body2"
                                            noWrap
                                            fontWeight={isCurrentTrack ? 600 : 400}
                                            color={isCurrentTrack ? 'primary.main' : 'text.primary'}
                                        >
                                            {getMusicTitle(track.file)}
                                        </Typography>
                                        <Typography
                                            variant="caption"
                                            color="text.secondary"
                                            noWrap
                                            component="div"
                                        >
                                            {getMusicArtist(track.file, t('MUSIC_UNKNOWN_ARTIST'))}
                                        </Typography>
                                    </Box>

                                    {canReorder && (
                                        <PlaylistTrackMoveButtons
                                            canMoveUp={index > 0}
                                            canMoveDown={index < tracks.length - 1 || hasNextPage}
                                            onMoveUp={() =>
                                                moveTrackToIndex(track.file.id, index, index - 1)
                                            }
                                            onMoveDown={() =>
                                                moveTrackToIndex(track.file.id, index, index + 1)
                                            }
                                        />
                                    )}

                                    {canRemoveTracks && (
                                        <IconButton
                                            className="remove-btn"
                                            size="small"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                onRemoveTrack(track.file.id);
                                            }}
                                            sx={{
                                                opacity: 0,
                                                color: 'text.secondary',
                                                '&:hover': { color: 'error.main' },
                                            }}
                                        >
                                            <Trash2 size={14} />
                                        </IconButton>
                                    )}

                                    {duration ? (
                                        <Typography
                                            variant="caption"
                                            color="text.secondary"
                                            sx={{
                                                flexShrink: 0,
                                                fontVariantNumeric: 'tabular-nums',
                                                minWidth: 36,
                                                textAlign: 'right',
                                            }}
                                        >
                                            {formatMusicDuration(duration)}
                                        </Typography>
                                    ) : (
                                        <Box sx={{ width: 36 }} />
                                    )}
                                </ListItemButton>
                            </ListItem>
                        );
                    })}
                </List>
            )}

            <LoadMoreSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={onLoadMore}
            />
        </Box>
    );
}

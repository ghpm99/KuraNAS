import { Box, Button, Drawer, IconButton, List, Typography } from '@mui/material';
import { ChevronDown, ChevronRight, ListMusic, Play, Pause, Save, Trash2, X } from 'lucide-react';
import { useState } from 'react';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import {
    getMusicTitle,
    getMusicArtist,
    formatMusicDuration,
    getTrackDurationSeconds,
} from '@/utils/music';
import useI18n from '@/components/i18n/provider/i18nContext';
import ClearQueueDialog from './ClearQueueDialog';
import QueueTrackRow from './QueueTrackRow';
import SaveQueueAsPlaylistDialog from './SaveQueueAsPlaylistDialog';

const DRAWER_WIDTH = 360;

const QueueDrawer = () => {
    const {
        queue = [],
        currentIndex,
        queueOpen,
        setQueueOpen,
        playTrackFromQueue,
        removeFromQueue,
        moveQueueItem,
        clearQueue,
        isPlaying,
        playbackContext,
    } = useGlobalMusic();
    const { t } = useI18n();
    const playbackContextLabel = playbackContext
        ? t(playbackContext.labelKey, playbackContext.labelParams)
        : '';

    const [isPlayedSectionOpen, setIsPlayedSectionOpen] = useState(false);
    const [isClearDialogOpen, setIsClearDialogOpen] = useState(false);
    const [isSaveDialogOpen, setIsSaveDialogOpen] = useState(false);
    const [draggedQueueIndex, setDraggedQueueIndex] = useState<number | undefined>(undefined);

    const currentTrack = currentIndex !== undefined ? queue[currentIndex] : undefined;
    const firstUpcomingIndex = currentIndex !== undefined ? currentIndex + 1 : 0;
    const playedTracks = queue.slice(0, currentIndex ?? 0);
    const upcomingTracks = queue.slice(firstUpcomingIndex);
    const hasQueueEntries = queue.length > 0;

    const dropOnQueueIndex = (targetIndex: number) => {
        if (draggedQueueIndex !== undefined) {
            moveQueueItem(draggedQueueIndex, targetIndex);
        }
        setDraggedQueueIndex(undefined);
    };

    const confirmClearQueue = () => {
        setIsClearDialogOpen(false);
        clearQueue();
    };

    return (
        <Drawer
            anchor="right"
            open={queueOpen}
            onClose={() => setQueueOpen(false)}
            variant="persistent"
            sx={{
                '& .MuiDrawer-paper': {
                    width: `min(${DRAWER_WIDTH}px, 100vw)`,
                    bgcolor: 'background.paper',
                    borderLeft: '1px solid',
                    borderColor: 'divider',
                    pt: '0px',
                    pb: '80px',
                },
            }}
        >
            <Box
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    p: 2,
                    pb: 1,
                }}
            >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <ListMusic size={20} />
                    <Typography variant="subtitle1" fontWeight={700}>
                        {t('MUSIC_QUEUE')}
                    </Typography>
                </Box>
                <Box sx={{ display: 'flex', gap: 0.5 }}>
                    <IconButton
                        size="small"
                        aria-label={t('MUSIC_QUEUE_SAVE_AS_PLAYLIST')}
                        disabled={!hasQueueEntries}
                        onClick={() => setIsSaveDialogOpen(true)}
                        sx={{ color: 'text.secondary', '&:hover': { color: 'primary.main' } }}
                    >
                        <Save size={16} />
                    </IconButton>
                    <IconButton
                        size="small"
                        aria-label={t('MUSIC_QUEUE_CLEAR')}
                        disabled={!hasQueueEntries}
                        onClick={() => setIsClearDialogOpen(true)}
                        sx={{ color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                    >
                        <Trash2 size={16} />
                    </IconButton>
                    <IconButton
                        size="small"
                        aria-label={t('MUSIC_QUEUE_CLOSE')}
                        onClick={() => setQueueOpen(false)}
                    >
                        <X size={18} />
                    </IconButton>
                </Box>
            </Box>

            {currentTrack && (
                <Box sx={{ px: 2, pb: 1 }}>
                    <Typography
                        variant="overline"
                        color="primary.main"
                        fontWeight={600}
                        sx={{ fontSize: '0.65rem' }}
                    >
                        {t('MUSIC_NOW_PLAYING')}
                    </Typography>
                    <Box
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 1.5,
                            p: 1,
                            borderRadius: 2,
                            bgcolor: 'rgba(var(--app-color-primary-rgb), 0.08)',
                            border: '1px solid',
                            borderColor: 'rgba(var(--app-color-primary-rgb), 0.2)',
                        }}
                    >
                        <Box
                            sx={{
                                width: 40,
                                height: 40,
                                borderRadius: 1,
                                bgcolor: 'primary.dark',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                flexShrink: 0,
                            }}
                        >
                            {isPlaying ? (
                                <Pause size={16} color="white" />
                            ) : (
                                <Play size={16} color="white" />
                            )}
                        </Box>
                        <Box sx={{ minWidth: 0, flex: 1 }}>
                            <Typography variant="body2" fontWeight={600} noWrap>
                                {getMusicTitle(currentTrack)}
                            </Typography>
                            <Typography variant="caption" color="text.secondary" noWrap>
                                {getMusicArtist(currentTrack, t('MUSIC_UNKNOWN_ARTIST'))}
                            </Typography>
                            {playbackContextLabel && (
                                <Typography
                                    variant="caption"
                                    color="text.secondary"
                                    noWrap
                                    component="div"
                                >
                                    {t('MUSIC_PLAYBACK_FROM', { context: playbackContextLabel })}
                                </Typography>
                            )}
                        </Box>
                        {getTrackDurationSeconds(currentTrack.metadata) > 0 && (
                            <Typography
                                variant="caption"
                                color="text.secondary"
                                sx={{ flexShrink: 0 }}
                            >
                                {formatMusicDuration(
                                    getTrackDurationSeconds(currentTrack.metadata)
                                )}
                            </Typography>
                        )}
                    </Box>
                </Box>
            )}

            {upcomingTracks.length > 0 && (
                <Box sx={{ px: 2, pt: 1 }}>
                    <Typography
                        variant="overline"
                        color="text.secondary"
                        sx={{ fontSize: '0.65rem' }}
                    >
                        {t('MUSIC_NEXT_IN_QUEUE')}
                    </Typography>
                </Box>
            )}

            <List sx={{ overflowY: 'auto', px: 1, pt: 0 }}>
                {upcomingTracks.map((track, upcomingPosition) => {
                    const queueIndex = firstUpcomingIndex + upcomingPosition;
                    return (
                        <QueueTrackRow
                            key={track.queueEntryId}
                            track={track}
                            isReorderable
                            canMoveUp={upcomingPosition > 0}
                            canMoveDown={upcomingPosition < upcomingTracks.length - 1}
                            isDragged={draggedQueueIndex === queueIndex}
                            onPlay={() => playTrackFromQueue(queueIndex)}
                            onRemove={() => removeFromQueue(track.queueEntryId)}
                            onMoveUp={() => moveQueueItem(queueIndex, queueIndex - 1)}
                            onMoveDown={() => moveQueueItem(queueIndex, queueIndex + 1)}
                            onDragStart={() => setDraggedQueueIndex(queueIndex)}
                            onDragEnd={() => setDraggedQueueIndex(undefined)}
                            onDrop={() => dropOnQueueIndex(queueIndex)}
                        />
                    );
                })}
            </List>

            {playedTracks.length > 0 && (
                <Box sx={{ px: 1, pt: 1 }}>
                    <Button
                        size="small"
                        color="inherit"
                        aria-expanded={isPlayedSectionOpen}
                        onClick={() => setIsPlayedSectionOpen((isOpen) => !isOpen)}
                        startIcon={
                            isPlayedSectionOpen ? (
                                <ChevronDown size={14} />
                            ) : (
                                <ChevronRight size={14} />
                            )
                        }
                        sx={{ color: 'text.secondary', fontSize: '0.65rem' }}
                    >
                        {t('MUSIC_QUEUE_PLAYED')} ({playedTracks.length})
                    </Button>
                    {isPlayedSectionOpen && (
                        <List sx={{ pt: 0 }}>
                            {playedTracks.map((track, playedPosition) => (
                                <QueueTrackRow
                                    key={track.queueEntryId}
                                    track={track}
                                    isReorderable={false}
                                    canMoveUp={false}
                                    canMoveDown={false}
                                    isDragged={false}
                                    onPlay={() => playTrackFromQueue(playedPosition)}
                                    onRemove={() => removeFromQueue(track.queueEntryId)}
                                    onMoveUp={() => undefined}
                                    onMoveDown={() => undefined}
                                    onDragStart={() => undefined}
                                    onDragEnd={() => undefined}
                                    onDrop={() => undefined}
                                />
                            ))}
                        </List>
                    )}
                </Box>
            )}

            <ClearQueueDialog
                isOpen={isClearDialogOpen}
                onConfirm={confirmClearQueue}
                onCancel={() => setIsClearDialogOpen(false)}
            />
            {isSaveDialogOpen && (
                <SaveQueueAsPlaylistDialog
                    fileIds={queue.map((track) => track.id)}
                    onClose={() => setIsSaveDialogOpen(false)}
                />
            )}
        </Drawer>
    );
};

export default QueueDrawer;

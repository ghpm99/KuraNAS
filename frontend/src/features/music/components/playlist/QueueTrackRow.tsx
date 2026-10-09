import { Box, IconButton, ListItem, ListItemButton, ListItemText, Typography } from '@mui/material';
import { ChevronDown, ChevronUp, GripVertical, Trash2 } from 'lucide-react';
import type { DragEvent } from 'react';
import type { QueueTrack } from '@/features/music/providers/globalMusic/queueEntries';
import {
    getMusicTitle,
    getMusicArtist,
    formatMusicDuration,
    getTrackDurationSeconds,
} from '@/utils/music';
import useI18n from '@/components/i18n/provider/i18nContext';

type QueueTrackRowProps = {
    track: QueueTrack;
    isReorderable: boolean;
    canMoveUp: boolean;
    canMoveDown: boolean;
    isDragged: boolean;
    onPlay: () => void;
    onRemove: () => void;
    onMoveUp: () => void;
    onMoveDown: () => void;
    onDragStart: () => void;
    onDragEnd: () => void;
    onDrop: () => void;
};

const actionButtonSx = {
    color: 'text.secondary',
    '&:hover': { color: 'primary.main' },
};

export default function QueueTrackRow({
    track,
    isReorderable,
    canMoveUp,
    canMoveDown,
    isDragged,
    onPlay,
    onRemove,
    onMoveUp,
    onMoveDown,
    onDragStart,
    onDragEnd,
    onDrop,
}: QueueTrackRowProps) {
    const { t } = useI18n();
    const trackTitle = getMusicTitle(track);
    const durationSeconds = getTrackDurationSeconds(track.metadata);

    const handleDragOver = (event: DragEvent<HTMLElement>) => {
        if (!isReorderable) return;
        event.preventDefault();
    };

    const handleDrop = (event: DragEvent<HTMLElement>) => {
        if (!isReorderable) return;
        event.preventDefault();
        onDrop();
    };

    return (
        <ListItem
            disablePadding
            draggable={isReorderable}
            onDragStart={isReorderable ? onDragStart : undefined}
            onDragEnd={onDragEnd}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
            sx={{
                borderRadius: 1,
                opacity: isDragged ? 0.4 : 1,
                '& .queue-row-actions': { opacity: 0 },
                '&:hover .queue-row-actions, &:focus-within .queue-row-actions': { opacity: 1 },
            }}
        >
            {isReorderable && (
                <Box
                    aria-hidden
                    sx={{ display: 'flex', color: 'text.disabled', cursor: 'grab', pl: 0.5 }}
                >
                    <GripVertical size={14} />
                </Box>
            )}
            <ListItemButton onClick={onPlay} sx={{ borderRadius: 1, py: 0.5, px: 1, minWidth: 0 }}>
                <ListItemText
                    primary={trackTitle}
                    secondary={getMusicArtist(track)}
                    primaryTypographyProps={{ variant: 'body2', noWrap: true, fontWeight: 500 }}
                    secondaryTypographyProps={{ variant: 'caption', noWrap: true }}
                />
                {durationSeconds > 0 && (
                    <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{ ml: 1, flexShrink: 0 }}
                    >
                        {formatMusicDuration(durationSeconds)}
                    </Typography>
                )}
            </ListItemButton>
            <Box className="queue-row-actions" sx={{ display: 'flex', flexShrink: 0 }}>
                {isReorderable && (
                    <>
                        <IconButton
                            size="small"
                            aria-label={t('MUSIC_QUEUE_MOVE_UP', { name: trackTitle })}
                            disabled={!canMoveUp}
                            onClick={onMoveUp}
                            sx={actionButtonSx}
                        >
                            <ChevronUp size={14} />
                        </IconButton>
                        <IconButton
                            size="small"
                            aria-label={t('MUSIC_QUEUE_MOVE_DOWN', { name: trackTitle })}
                            disabled={!canMoveDown}
                            onClick={onMoveDown}
                            sx={actionButtonSx}
                        >
                            <ChevronDown size={14} />
                        </IconButton>
                    </>
                )}
                <IconButton
                    size="small"
                    aria-label={t('MUSIC_QUEUE_REMOVE_TRACK', { name: trackTitle })}
                    onClick={onRemove}
                    sx={{ color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                >
                    <Trash2 size={14} />
                </IconButton>
            </Box>
        </ListItem>
    );
}

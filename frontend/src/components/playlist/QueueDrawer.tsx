import {
    Box,
    Drawer,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemText,
    Typography,
} from '@mui/material';
import { ListMusic, Play, Pause, Trash2, X } from 'lucide-react';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { getMusicTitle, getMusicArtist, formatMusicDuration } from '@/utils/music';
import useI18n from '@/components/i18n/provider/i18nContext';

const DRAWER_WIDTH = 360;

const QueueDrawer = () => {
    const {
        queue = [],
        currentIndex,
        queueOpen,
        setQueueOpen,
        playTrackFromQueue,
        removeFromQueue,
        clearQueue,
        isPlaying,
        playbackContext,
    } = useGlobalMusic();
    const { t } = useI18n();
    const playbackContextLabel = playbackContext
        ? t(playbackContext.labelKey, playbackContext.labelParams)
        : '';

    const currentTrack = currentIndex !== undefined ? queue[currentIndex] : undefined;
    const upcomingTracks = queue
        .map((track, index) => ({ track, index }))
        .filter(({ index }) => index !== currentIndex);

    return (
        <Drawer
            anchor="right"
            open={queueOpen}
            onClose={() => setQueueOpen(false)}
            variant="persistent"
            sx={{
                '& .MuiDrawer-paper': {
                    width: DRAWER_WIDTH,
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
                        onClick={clearQueue}
                        sx={{ color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                    >
                        <Trash2 size={16} />
                    </IconButton>
                    <IconButton size="small" onClick={() => setQueueOpen(false)}>
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
                            bgcolor: 'rgba(99, 102, 241, 0.08)',
                            border: '1px solid',
                            borderColor: 'rgba(99, 102, 241, 0.2)',
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
                                {getMusicArtist(currentTrack)}
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
                        {currentTrack.metadata?.duration && (
                            <Typography
                                variant="caption"
                                color="text.secondary"
                                sx={{ flexShrink: 0 }}
                            >
                                {formatMusicDuration(currentTrack.metadata.duration)}
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

            <List sx={{ flex: 1, overflowY: 'auto', px: 1, pt: 0 }}>
                {upcomingTracks.map(({ track, index }) => (
                    <ListItem
                        key={`${track.id}-${index}`}
                        disablePadding
                        secondaryAction={
                            <IconButton
                                edge="end"
                                size="small"
                                onClick={() => removeFromQueue(index)}
                                sx={{
                                    color: 'text.secondary',
                                    opacity: 0,
                                    '&:hover': { color: 'error.main', opacity: 1 },
                                }}
                            >
                                <Trash2 size={14} />
                            </IconButton>
                        }
                        sx={{
                            '&:hover .MuiIconButton-root': { opacity: 1 },
                            borderRadius: 1,
                        }}
                    >
                        <ListItemButton
                            onClick={() => playTrackFromQueue(index)}
                            sx={{ borderRadius: 1, py: 0.5, px: 1 }}
                        >
                            <ListItemText
                                primary={getMusicTitle(track)}
                                secondary={getMusicArtist(track)}
                                primaryTypographyProps={{
                                    variant: 'body2',
                                    noWrap: true,
                                    fontWeight: 500,
                                }}
                                secondaryTypographyProps={{ variant: 'caption', noWrap: true }}
                            />
                            {track.metadata?.duration && (
                                <Typography
                                    variant="caption"
                                    color="text.secondary"
                                    sx={{ ml: 1, flexShrink: 0 }}
                                >
                                    {formatMusicDuration(track.metadata.duration)}
                                </Typography>
                            )}
                        </ListItemButton>
                    </ListItem>
                ))}
            </List>
        </Drawer>
    );
};

export default QueueDrawer;

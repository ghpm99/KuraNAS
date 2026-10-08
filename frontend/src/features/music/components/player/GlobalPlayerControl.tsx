import {
    Box,
    Card,
    CardContent,
    IconButton,
    LinearProgress,
    Slider,
    Typography,
} from '@mui/material';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useEffect, useState } from 'react';
import type { KeyboardEvent, MouseEvent } from 'react';
import {
    ListMusic,
    Pause,
    Play,
    Repeat,
    Repeat1,
    Shuffle,
    SkipBack,
    SkipForward,
    Volume2,
    VolumeX,
} from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import QueueDrawer from '@/features/music/components/playlist/QueueDrawer';
import { getMusicTitle, getMusicArtist } from '@/utils/music';
import ExpandedPlayerSheet from './ExpandedPlayerSheet';
import { formatPlaybackTime } from './formatPlaybackTime';
import { nextRepeatMode } from './nextRepeatMode';
import { viewportMediaQueries } from '@/theme/visualTokens';
import styles from './GlobalPlayerControl.module.css';
import '../playerControl/playerControl.css';

const COMPACT_PLAYER_MEDIA_QUERY = viewportMediaQueries.belowPhone;
const GLOBAL_PLAYER_VISIBILITY_ATTRIBUTE = 'data-global-player';
const compactTouchTargetSx = { width: { xs: 44, sm: 'auto' }, height: { xs: 44, sm: 'auto' } };

const GlobalPlayerControl = () => {
    const { t } = useI18n();
    const {
        isPlaying,
        currentTime,
        duration,
        volume,
        shuffle,
        repeatMode,
        togglePlayPause,
        next,
        previous,
        seek,
        setVolume,
        toggleShuffle,
        setRepeatMode,
        currentTrack,
        hasQueue,
        playbackContext,
        toggleQueue,
        queueOpen,
    } = useGlobalMusic();
    const isCompactPlayer = useMediaQuery(COMPACT_PLAYER_MEDIA_QUERY);
    const [isExpandedSheetOpen, setIsExpandedSheetOpen] = useState(false);

    useEffect(() => {
        if (!hasQueue) return;
        const rootElement = document.documentElement;
        rootElement.setAttribute(GLOBAL_PLAYER_VISIBILITY_ATTRIBUTE, 'visible');
        return () => rootElement.removeAttribute(GLOBAL_PLAYER_VISIBILITY_ATTRIBUTE);
    }, [hasQueue]);

    if (!hasQueue) return null;

    const currentTrackTitle = getMusicTitle
        ? getMusicTitle(currentTrack!)
        : currentTrack?.metadata?.title || currentTrack?.name || '';
    const currentTrackArtist = getMusicArtist
        ? getMusicArtist(currentTrack!)
        : currentTrack?.metadata?.artist || '';
    const playbackContextLabel = playbackContext
        ? t(playbackContext.labelKey, playbackContext.labelParams)
        : '';
    const safeCurrentTime = Number.isFinite(currentTime) ? currentTime : 0;
    const safeDuration = Number.isFinite(duration) && duration > 0 ? duration : 0;

    const progressPercent = safeDuration > 0 ? (safeCurrentTime / safeDuration) * 100 : 0;

    const openExpandedSheet = () => setIsExpandedSheetOpen(true);
    const closeExpandedSheet = () => setIsExpandedSheetOpen(false);

    const openExpandedSheetFromCard = (event: MouseEvent<HTMLElement>) => {
        if (!isCompactPlayer) return;
        if ((event.target as HTMLElement).closest('button')) return;
        openExpandedSheet();
    };

    const openExpandedSheetFromKeyboard = (event: KeyboardEvent<HTMLElement>) => {
        if (event.key !== 'Enter' && event.key !== ' ') return;
        event.preventDefault();
        openExpandedSheet();
    };

    const cycleRepeatMode = () => setRepeatMode(nextRepeatMode(repeatMode));

    const RepeatIcon = repeatMode === 'one' ? Repeat1 : Repeat;

    return (
        <>
            <Card
                className={`player-control ${styles.player}`}
                onClick={openExpandedSheetFromCard}
                sx={{ zIndex: (theme) => theme.zIndex.appBar + 50 }}
            >
                {isCompactPlayer && (
                    <LinearProgress
                        className={styles.progressBar}
                        variant="determinate"
                        value={progressPercent}
                        aria-label={t('PLAYER_ARIA_PROGRESS')}
                    />
                )}
                <CardContent
                    sx={{
                        display: 'grid',
                        gridTemplateColumns: { xs: 'minmax(0, 1fr) auto', sm: '1fr 2fr 1fr' },
                        alignItems: 'center',
                        gap: { xs: 1, sm: 2 },
                        p: { xs: 1, sm: 2 },
                    }}
                >
                    <Box
                        className={isCompactPlayer ? styles.expandableInfo : undefined}
                        role={isCompactPlayer ? 'button' : undefined}
                        tabIndex={isCompactPlayer ? 0 : undefined}
                        aria-label={isCompactPlayer ? t('PLAYER_ARIA_OPEN_EXPANDED') : undefined}
                        onKeyDown={isCompactPlayer ? openExpandedSheetFromKeyboard : undefined}
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: { xs: 1, sm: 1.5 },
                            minWidth: 0,
                        }}
                    >
                        <Box
                            sx={{
                                width: { xs: 38, sm: 46 },
                                height: { xs: 38, sm: 46 },
                                bgcolor: 'primary.dark',
                                borderRadius: 1.5,
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                flexShrink: 0,
                            }}
                        >
                            {isPlaying ? (
                                <Box
                                    sx={{
                                        display: 'flex',
                                        alignItems: 'flex-end',
                                        gap: '2px',
                                        height: 18,
                                    }}
                                >
                                    {[1, 2, 3].map((bar) => (
                                        <Box
                                            key={bar}
                                            sx={{
                                                width: 3,
                                                bgcolor: 'white',
                                                borderRadius: 1,
                                                height: '10px',
                                                animation: `eqPlayer ${0.4 + bar * 0.15}s ease-in-out infinite alternate`,
                                                '@keyframes eqPlayer': {
                                                    '0%': { height: '4px' },
                                                    '100%': { height: '16px' },
                                                },
                                            }}
                                        />
                                    ))}
                                </Box>
                            ) : (
                                <Volume2 size={22} color="white" />
                            )}
                        </Box>
                        <Box sx={{ minWidth: 0 }}>
                            <Typography
                                variant="body2"
                                fontWeight={600}
                                noWrap
                                sx={{ fontSize: { xs: '0.8rem', sm: '0.875rem' } }}
                            >
                                {currentTrackTitle}
                            </Typography>
                            <Typography
                                variant="caption"
                                color="text.secondary"
                                noWrap
                                component="div"
                            >
                                {currentTrackArtist}
                            </Typography>
                            {playbackContextLabel && (
                                <Typography
                                    variant="caption"
                                    color="text.secondary"
                                    noWrap
                                    component="div"
                                    sx={{ display: { xs: 'none', md: 'block' } }}
                                >
                                    {t('MUSIC_PLAYBACK_FROM', { context: playbackContextLabel })}
                                </Typography>
                            )}
                        </Box>
                    </Box>

                    <Box
                        sx={{
                            display: 'flex',
                            flexDirection: { xs: 'row', sm: 'column' },
                            alignItems: 'center',
                            gap: 0.5,
                        }}
                    >
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 0, sm: 1 } }}>
                            <IconButton
                                onClick={toggleShuffle}
                                size="small"
                                aria-label={t('PLAYER_ARIA_SHUFFLE')}
                                sx={{
                                    opacity: shuffle ? 1 : 0.4,
                                    display: { xs: 'none', sm: 'inline-flex' },
                                }}
                            >
                                <Shuffle size={16} />
                            </IconButton>
                            <IconButton
                                onClick={previous}
                                size="small"
                                aria-label={t('PLAYER_ARIA_PREVIOUS')}
                                sx={compactTouchTargetSx}
                            >
                                <SkipBack size={18} />
                            </IconButton>
                            <IconButton
                                onClick={togglePlayPause}
                                aria-label={
                                    isPlaying ? t('PLAYER_ARIA_PAUSE') : t('PLAYER_ARIA_PLAY')
                                }
                                sx={{
                                    bgcolor: 'text.primary',
                                    color: 'background.paper',
                                    width: { xs: 44, sm: 34 },
                                    height: { xs: 44, sm: 34 },
                                    '&:hover': {
                                        bgcolor: 'rgba(var(--app-color-ink-rgb), 0.85)',
                                        transform: 'scale(1.05)',
                                    },
                                    transition: 'all 0.15s ease',
                                }}
                            >
                                {isPlaying ? (
                                    <Pause size={18} />
                                ) : (
                                    <Play size={18} style={{ marginLeft: 2 }} />
                                )}
                            </IconButton>
                            <IconButton
                                onClick={next}
                                size="small"
                                aria-label={t('PLAYER_ARIA_NEXT')}
                                sx={compactTouchTargetSx}
                            >
                                <SkipForward size={18} />
                            </IconButton>
                            <IconButton
                                onClick={cycleRepeatMode}
                                size="small"
                                aria-label={t('PLAYER_ARIA_REPEAT')}
                                sx={{
                                    opacity: repeatMode !== 'none' ? 1 : 0.4,
                                    color: repeatMode !== 'none' ? 'primary.main' : undefined,
                                    display: { xs: 'none', sm: 'inline-flex' },
                                }}
                            >
                                <RepeatIcon size={16} />
                            </IconButton>
                        </Box>
                        <Box
                            sx={{
                                display: { xs: 'none', sm: 'flex' },
                                alignItems: 'center',
                                gap: 1,
                                width: '100%',
                                maxWidth: 500,
                            }}
                        >
                            <Typography
                                variant="caption"
                                sx={{ minWidth: 36, textAlign: 'right', fontSize: '0.7rem' }}
                            >
                                {formatPlaybackTime(currentTime)}
                            </Typography>
                            <Slider
                                size="small"
                                aria-label={t('PLAYER_ARIA_SEEK')}
                                value={safeCurrentTime}
                                max={safeDuration || 100}
                                onChange={(_, value) => seek(value as number)}
                                sx={{
                                    flexGrow: 1,
                                    height: 4,
                                    '& .MuiSlider-thumb': {
                                        width: 0,
                                        height: 0,
                                        transition: 'width 0.15s, height 0.15s',
                                    },
                                    '&:hover .MuiSlider-thumb': {
                                        width: 12,
                                        height: 12,
                                    },
                                }}
                            />
                            <Typography variant="caption" sx={{ minWidth: 36, fontSize: '0.7rem' }}>
                                {formatPlaybackTime(duration)}
                            </Typography>
                        </Box>
                    </Box>

                    <Box
                        sx={{
                            display: { xs: 'none', sm: 'flex' },
                            alignItems: 'center',
                            justifyContent: 'flex-end',
                            gap: 1,
                        }}
                    >
                        <IconButton
                            size="small"
                            onClick={toggleQueue}
                            aria-label={t('PLAYER_ARIA_QUEUE')}
                            sx={{
                                color: queueOpen ? 'primary.main' : 'text.secondary',
                                '&:hover': {
                                    color: queueOpen ? 'primary.light' : 'text.primary',
                                },
                            }}
                        >
                            <ListMusic size={18} />
                        </IconButton>
                        <Box
                            sx={{
                                display: { xs: 'none', md: 'flex' },
                                alignItems: 'center',
                                gap: 0.5,
                                width: 120,
                            }}
                        >
                            <IconButton
                                size="small"
                                onClick={() => setVolume(volume > 0 ? 0 : 0.7)}
                                aria-label={
                                    volume === 0 ? t('PLAYER_ARIA_UNMUTE') : t('PLAYER_ARIA_MUTE')
                                }
                            >
                                {volume === 0 ? <VolumeX size={16} /> : <Volume2 size={16} />}
                            </IconButton>
                            <Slider
                                size="small"
                                aria-label={t('PLAYER_ARIA_VOLUME')}
                                value={volume}
                                max={1}
                                step={0.01}
                                onChange={(_, value) => setVolume(value as number)}
                                sx={{
                                    width: 80,
                                    height: 4,
                                    '& .MuiSlider-thumb': {
                                        width: 0,
                                        height: 0,
                                        transition: 'width 0.15s, height 0.15s',
                                    },
                                    '&:hover .MuiSlider-thumb': {
                                        width: 10,
                                        height: 10,
                                    },
                                }}
                            />
                        </Box>
                    </Box>
                </CardContent>
            </Card>
            {isCompactPlayer && (
                <ExpandedPlayerSheet
                    isOpen={isExpandedSheetOpen}
                    onOpen={openExpandedSheet}
                    onClose={closeExpandedSheet}
                />
            )}
            <QueueDrawer />
        </>
    );
};

export default GlobalPlayerControl;

import { Box, IconButton, Slider, SwipeableDrawer, Typography } from '@mui/material';
import {
    ChevronDown,
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
import CoverArt from '@/features/music/components/CoverArt';
import { getTrackCoverUrl } from '@/service/musicCover';
import { getMusicArtist, getMusicTitle } from '@/utils/music';
import { formatPlaybackTime } from './formatPlaybackTime';
import { nextRepeatMode } from './nextRepeatMode';
import { supportsProgrammaticVolume } from './supportsProgrammaticVolume';
import styles from './ExpandedPlayerSheet.module.css';

interface ExpandedPlayerSheetProps {
    isOpen: boolean;
    onOpen: () => void;
    onClose: () => void;
}

const touchTargetSx = { width: 44, height: 44 };

const isReducedMotionEnabled = (): boolean =>
    document.documentElement.getAttribute('data-app-motion') === 'reduced';

const ExpandedPlayerSheet = ({ isOpen, onOpen, onClose }: ExpandedPlayerSheetProps) => {
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
        setQueueOpen,
    } = useGlobalMusic();

    const trackTitle = currentTrack ? getMusicTitle(currentTrack) : '';
    const trackArtist = currentTrack ? getMusicArtist(currentTrack) : '';
    const safeCurrentTime = Number.isFinite(currentTime) ? currentTime : 0;
    const safeDuration = Number.isFinite(duration) && duration > 0 ? duration : 0;
    const RepeatIcon = repeatMode === 'one' ? Repeat1 : Repeat;
    const isRepeatActive = repeatMode !== 'none';

    const playbackIcon = isPlaying ? (
        <Pause size={48} color="white" />
    ) : (
        <Play size={48} color="white" />
    );

    const openQueue = () => {
        onClose();
        setQueueOpen(true);
    };

    return (
        <SwipeableDrawer
            anchor="bottom"
            open={isOpen}
            onOpen={onOpen}
            onClose={onClose}
            disableSwipeToOpen
            transitionDuration={isReducedMotionEnabled() ? 0 : undefined}
        >
            <Box className={styles.sheet}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <IconButton
                        sx={touchTargetSx}
                        onClick={onClose}
                        aria-label={t('PLAYER_ARIA_CLOSE_EXPANDED')}
                    >
                        <ChevronDown size={22} />
                    </IconButton>
                    <IconButton
                        sx={touchTargetSx}
                        onClick={openQueue}
                        aria-label={t('PLAYER_ARIA_QUEUE')}
                    >
                        <ListMusic size={20} />
                    </IconButton>
                </Box>

                <Box className={styles.artwork}>
                    {currentTrack ? (
                        <CoverArt
                            src={getTrackCoverUrl(currentTrack.id, 512)}
                            fallback={playbackIcon}
                        />
                    ) : (
                        playbackIcon
                    )}
                </Box>

                <Box className={styles.trackText}>
                    <Typography variant="h3" noWrap>
                        {trackTitle}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap>
                        {trackArtist}
                    </Typography>
                </Box>

                <Box className={styles.timeRow}>
                    <Typography variant="caption">{formatPlaybackTime(safeCurrentTime)}</Typography>
                    <Slider
                        aria-label={t('PLAYER_ARIA_SEEK')}
                        value={safeCurrentTime}
                        max={safeDuration || 100}
                        onChange={(_, seekPosition) => seek(seekPosition as number)}
                    />
                    <Typography variant="caption">{formatPlaybackTime(safeDuration)}</Typography>
                </Box>

                <Box className={styles.controlsRow}>
                    <IconButton
                        onClick={toggleShuffle}
                        aria-label={t('PLAYER_ARIA_SHUFFLE')}
                        sx={{ ...touchTargetSx, opacity: shuffle ? 1 : 0.4 }}
                    >
                        <Shuffle size={20} />
                    </IconButton>
                    <IconButton
                        sx={touchTargetSx}
                        onClick={previous}
                        aria-label={t('PLAYER_ARIA_PREVIOUS')}
                    >
                        <SkipBack size={24} />
                    </IconButton>
                    <IconButton
                        onClick={togglePlayPause}
                        aria-label={isPlaying ? t('PLAYER_ARIA_PAUSE') : t('PLAYER_ARIA_PLAY')}
                        sx={{
                            width: 56,
                            height: 56,
                            bgcolor: 'primary.main',
                            color: 'common.white',
                            '&:hover': { bgcolor: 'primary.light' },
                        }}
                    >
                        {isPlaying ? <Pause size={26} /> : <Play size={26} />}
                    </IconButton>
                    <IconButton
                        sx={touchTargetSx}
                        onClick={next}
                        aria-label={t('PLAYER_ARIA_NEXT')}
                    >
                        <SkipForward size={24} />
                    </IconButton>
                    <IconButton
                        onClick={() => setRepeatMode(nextRepeatMode(repeatMode))}
                        aria-label={t('PLAYER_ARIA_REPEAT')}
                        sx={{
                            ...touchTargetSx,
                            opacity: isRepeatActive ? 1 : 0.4,
                            color: isRepeatActive ? 'primary.main' : undefined,
                        }}
                    >
                        <RepeatIcon size={20} />
                    </IconButton>
                </Box>

                {supportsProgrammaticVolume() && (
                    <Box className={styles.volumeRow}>
                        <IconButton
                            sx={touchTargetSx}
                            onClick={() => setVolume(volume > 0 ? 0 : 0.7)}
                            aria-label={
                                volume === 0 ? t('PLAYER_ARIA_UNMUTE') : t('PLAYER_ARIA_MUTE')
                            }
                        >
                            {volume === 0 ? <VolumeX size={20} /> : <Volume2 size={20} />}
                        </IconButton>
                        <Slider
                            aria-label={t('PLAYER_ARIA_VOLUME')}
                            value={volume}
                            max={1}
                            step={0.01}
                            onChange={(_, nextVolume) => setVolume(nextVolume as number)}
                        />
                    </Box>
                )}
            </Box>
        </SwipeableDrawer>
    );
};

export default ExpandedPlayerSheet;

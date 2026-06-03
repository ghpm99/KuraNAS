import { Box, Card, CardContent, IconButton, Slider, Typography } from '@mui/material';
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
import { getMusicTitle, getMusicArtist } from '@/utils/music';
import '../playerControl/playerControl.css';

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

    const formatTime = (time: number): string => {
        if (isNaN(time)) return '0:00';
        const minutes = Math.floor(time / 60);
        const seconds = Math.floor(time % 60);
        return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    };

    const cycleRepeatMode = () => {
        const modes = ['none', 'all', 'one'] as const;
        const currentIdx = modes.indexOf(repeatMode);
        const nextMode =
            currentIdx === -1 ? modes[0] : (modes[(currentIdx + 1) % modes.length] ?? modes[0]);
        setRepeatMode(nextMode);
    };

    const RepeatIcon = repeatMode === 'one' ? Repeat1 : Repeat;

    return (
        <Card
            className="player-control"
            sx={{ position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 1300 }}
        >
            <CardContent
                sx={{
                    display: 'grid',
                    gridTemplateColumns: { xs: 'minmax(0, 1fr) auto', sm: '1fr 2fr 1fr' },
                    alignItems: 'center',
                    gap: { xs: 1, sm: 2 },
                    p: { xs: 1, sm: 2 },
                }}
            >
                {/* Left: Track Info */}
                <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1, sm: 1.5 }, minWidth: 0 }}>
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
                        <Typography variant="body2" fontWeight={600} noWrap sx={{ fontSize: { xs: '0.8rem', sm: '0.875rem' } }}>
                            {currentTrackTitle}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" noWrap component="div">
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

                {/* Center: Controls + Progress */}
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
                            aria-label="toggle shuffle"
                            sx={{ opacity: shuffle ? 1 : 0.4, display: { xs: 'none', sm: 'inline-flex' } }}
                        >
                            <Shuffle size={16} />
                        </IconButton>
                        <IconButton onClick={previous} size="small" aria-label="previous track">
                            <SkipBack size={18} />
                        </IconButton>
                        <IconButton
                            onClick={togglePlayPause}
                            aria-label={isPlaying ? 'pause playback' : 'play playback'}
                            sx={{
                                bgcolor: 'white',
                                color: 'black',
                                width: 34,
                                height: 34,
                                '&:hover': {
                                    bgcolor: 'rgba(255,255,255,0.85)',
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
                        <IconButton onClick={next} size="small" aria-label="next track">
                            <SkipForward size={18} />
                        </IconButton>
                        <IconButton
                            onClick={cycleRepeatMode}
                            size="small"
                            aria-label="change repeat mode"
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
                            {formatTime(currentTime)}
                        </Typography>
                        <Slider
                            size="small"
                            aria-label="seek playback"
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
                            {formatTime(duration)}
                        </Typography>
                    </Box>
                </Box>

                {/* Right: Volume + Queue */}
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
                        aria-label="toggle queue"
                        sx={{
                            color: queueOpen ? 'primary.main' : 'text.secondary',
                            '&:hover': {
                                color: queueOpen ? 'primary.light' : 'text.primary',
                            },
                        }}
                    >
                        <ListMusic size={18} />
                    </IconButton>
                    <Box sx={{ display: { xs: 'none', md: 'flex' }, alignItems: 'center', gap: 0.5, width: 120 }}>
                        <IconButton
                            size="small"
                            onClick={() => setVolume(volume > 0 ? 0 : 0.7)}
                            aria-label={volume === 0 ? 'unmute volume' : 'mute volume'}
                        >
                            {volume === 0 ? <VolumeX size={16} /> : <Volume2 size={16} />}
                        </IconButton>
                        <Slider
                            size="small"
                            aria-label="set volume"
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
    );
};

export default GlobalPlayerControl;

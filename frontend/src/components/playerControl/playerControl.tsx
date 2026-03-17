import { Box, Card, CardContent, IconButton, Slider, Typography } from '@mui/material';
import { Pause, Play, SkipBack, SkipForward, Volume2 } from 'lucide-react';
import './playerControl.css';
import { useGlobalMusic } from '../providers/GlobalMusicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';

const PlayerControl = () => {
    const {
        queue,
        currentIndex,
        isPlaying,
        currentTime,
        duration,
        volume,
        next,
        previous,
        seek,
        setVolume,
        togglePlayPause,
    } = useGlobalMusic();
    const { t } = useI18n();
    const safeCurrentTime = Number.isFinite(currentTime) ? currentTime : 0;
    const safeDuration = Number.isFinite(duration) && duration > 0 ? duration : 0;
    const safeVolume = Number.isFinite(volume) ? volume : 0;

    const formatTime = (time: number): string => {
        if (isNaN(time)) return '0:00';
        const minutes = Math.floor(time / 60);
        const seconds = Math.floor(time % 60);
        return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    };

    const getTrackTitle = (): string => {
        if (currentIndex === undefined) return t('PLAYER_NO_TRACK');
        return (
            queue[currentIndex]?.metadata?.title ||
            queue[currentIndex]?.name ||
            t('PLAYER_UNKNOWN_TITLE')
        );
    };

    const getTrackArtist = (): string => {
        if (currentIndex === undefined) return '';
        return queue[currentIndex]?.metadata?.artist || t('PLAYER_UNKNOWN_ARTIST');
    };

    return (
        <>
            <Card
                className="player-control"
                sx={{ position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 1000 }}
            >
                <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2, p: 2 }}>
                    {/* Track Info */}
                    <Box
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 2,
                            minWidth: 200,
                        }}
                    >
                        <Box
                            sx={{
                                width: 50,
                                height: 50,
                                bgcolor: 'primary.main',
                                borderRadius: 1,
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}
                        >
                            <Volume2 size={24} color="white" />
                        </Box>
                        <Box sx={{ minWidth: 0 }}>
                            <Typography variant="subtitle2" noWrap>
                                {getTrackTitle()}
                            </Typography>
                            <Typography variant="caption" color="text.secondary" noWrap>
                                {getTrackArtist()}
                            </Typography>
                        </Box>
                    </Box>

                    {/* Controls */}
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <IconButton onClick={previous} size="small">
                            <SkipBack size={20} />
                        </IconButton>
                        <IconButton
                            onClick={togglePlayPause}
                            sx={{
                                bgcolor: 'primary.main',
                                '&:hover': { bgcolor: 'primary.dark' },
                            }}
                        >
                            {!isPlaying && <Play size={20} color="white" />}
                            {isPlaying && <Pause size={20} color="white" />}
                        </IconButton>
                        <IconButton onClick={next} size="small">
                            <SkipForward size={20} />
                        </IconButton>
                    </Box>

                    {/* Progress */}
                    <Box
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 2,
                            flexGrow: 1,
                            minWidth: 200,
                        }}
                    >
                        <Typography variant="caption" sx={{ minWidth: 40 }}>
                            {formatTime(currentTime)}
                        </Typography>
                        <Slider
                            size="small"
                            value={safeCurrentTime}
                            max={safeDuration || 100}
                            onChange={(_, value) => seek(value as number)}
                            sx={{ flexGrow: 1 }}
                        />
                        <Typography variant="caption" sx={{ minWidth: 40 }}>
                            {formatTime(duration)}
                        </Typography>
                    </Box>

                    {/* Volume */}
                    <Box
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 1,
                            minWidth: 120,
                        }}
                    >
                        <Volume2 size={20} />
                        <Slider
                            size="small"
                            value={safeVolume}
                            max={1}
                            step={0.1}
                            onChange={(_, value) => setVolume(value as number)}
                            sx={{ width: 80 }}
                        />
                    </Box>
                </CardContent>
            </Card>
        </>
    );
};

export default PlayerControl;

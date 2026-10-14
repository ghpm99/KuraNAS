import { IconButton, Slider, Box, Typography, Menu, MenuItem, Tooltip } from '@mui/material';
import {
    Play,
    Pause,
    VolumeOff,
    Maximize,
    Minimize,
    Settings,
    Rewind,
    FastForward,
    SkipBack,
    SkipForward,
    Volume2,
    Keyboard,
    PictureInPicture2,
} from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { usePlayerPortalContainer } from '@/features/videos/videoPlayer/playerPortalContainer';
import './videoControls.css';

interface VideoControlsProps {
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    volume: number;
    isMuted: boolean;
    playbackRate: number;
    isFullscreen: boolean;
    isPictureInPictureSupported?: boolean;
    isPictureInPicture?: boolean;
    areControlsVisible?: boolean;
    seekTo: (time: number) => void;
    setVolume: (volume: number) => void;
    toggleMute: () => void;
    setPlaybackRate: (rate: number) => void;
    toggleFullscreen: () => void;
    togglePictureInPicture?: () => void;
    togglePlayPause: () => void;
    nextVideo: () => void;
    previousVideo: () => void;
    canGoNext?: boolean;
    canGoPrevious?: boolean;
    onOpenShortcutsHelp?: () => void;
}
const VideoControls = ({
    isPlaying,
    currentTime,
    duration,
    volume,
    isMuted,
    playbackRate,
    isFullscreen,
    isPictureInPictureSupported = false,
    isPictureInPicture = false,
    areControlsVisible = true,
    seekTo,
    setVolume,
    toggleMute,
    setPlaybackRate,
    toggleFullscreen,
    togglePictureInPicture,
    togglePlayPause,
    nextVideo,
    previousVideo,
    canGoNext = true,
    canGoPrevious = true,
    onOpenShortcutsHelp,
}: VideoControlsProps) => {
    const { t } = useI18n();
    const [settingsAnchor, setSettingsAnchor] = useState<null | HTMLElement>(null);
    const portalContainer = usePlayerPortalContainer();

    const formatTime = (time: number): string => {
        if (isNaN(time)) return '0:00';
        const hours = Math.floor(time / 3600);
        const minutes = Math.floor((time % 3600) / 60);
        const seconds = Math.floor(time % 60);

        if (hours > 0) {
            return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
        }
        return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    };

    const handleSeek = (_event: Event, newValue: number | number[]) => {
        seekTo(newValue as number);
    };

    const handleVolumeChange = (_event: Event, newValue: number | number[]) => {
        setVolume(newValue as number);
    };

    const handleSettingsClick = (event: React.MouseEvent<HTMLElement>) => {
        setSettingsAnchor(event.currentTarget);
    };

    const handleSettingsClose = () => {
        setSettingsAnchor(null);
    };

    const handlePlaybackRateChange = (rate: number) => {
        setPlaybackRate(rate);
        handleSettingsClose();
    };

    const skipForward = () => {
        seekTo(Math.min(currentTime + 10, duration));
    };

    const skipBackward = () => {
        seekTo(Math.max(currentTime - 10, 0));
    };

    const fullscreenLabel = isFullscreen ? t('VIDEO_EXIT_FULLSCREEN') : t('VIDEO_FULLSCREEN');
    const pictureInPictureLabel = isPictureInPicture
        ? t('VIDEO_EXIT_PICTURE_IN_PICTURE')
        : t('VIDEO_PICTURE_IN_PICTURE');
    const isSilent = isMuted || volume === 0;
    const muteLabel = isSilent ? t('VIDEO_UNMUTE') : t('VIDEO_MUTE');
    const tooltipSlotProps = { popper: { container: portalContainer } };
    const playbackRates = [0.5, 0.75, 1, 1.25, 1.5, 2];

    if (!areControlsVisible) {
        return null;
    }

    return (
        <div className="video-controls">
            {/* Central Play/Pause Button */}
            <Box className="center-controls">
                <IconButton onClick={skipBackward} size="large" className="skip-button">
                    <Rewind size={24} />
                </IconButton>

                <IconButton onClick={togglePlayPause} size="large" className="play-pause-button">
                    {isPlaying ? <Pause size={32} /> : <Play size={32} />}
                </IconButton>

                <IconButton onClick={skipForward} size="large" className="skip-button">
                    <FastForward size={24} />
                </IconButton>
            </Box>

            {/* Bottom Controls Bar */}
            <Box className="bottom-controls">
                {/* Left: Time and Skip */}
                <Box className="left-controls">
                    <Typography variant="caption" className="time-display">
                        {formatTime(currentTime)}
                    </Typography>
                    <IconButton onClick={previousVideo} size="small" disabled={!canGoPrevious}>
                        <SkipBack size={20} />
                    </IconButton>
                    <IconButton onClick={nextVideo} size="small" disabled={!canGoNext}>
                        <SkipForward size={20} />
                    </IconButton>
                </Box>

                {/* Center: Progress Bar */}
                <Box className="progress-container">
                    <Slider
                        size="small"
                        value={currentTime}
                        max={duration || 100}
                        onChange={handleSeek}
                        className="progress-bar"
                        sx={{ flexGrow: 1 }}
                    />
                    <Typography variant="caption" className="duration-display">
                        {formatTime(duration)}
                    </Typography>
                </Box>

                {/* Right: Volume and Settings */}
                <Box className="right-controls">
                    <IconButton
                        onClick={toggleMute}
                        size="small"
                        aria-label={muteLabel}
                        aria-pressed={isSilent}
                    >
                        {isSilent ? <VolumeOff size={20} /> : <Volume2 size={20} />}
                    </IconButton>

                    <Slider
                        size="small"
                        value={volume}
                        max={1}
                        step={0.1}
                        onChange={handleVolumeChange}
                        className="volume-slider"
                        sx={{ width: 80 }}
                    />

                    <IconButton onClick={handleSettingsClick} size="small">
                        <Settings size={20} />
                    </IconButton>

                    {onOpenShortcutsHelp ? (
                        <Tooltip
                            title={t('VIDEO_SHORTCUTS_HELP_BUTTON')}
                            slotProps={tooltipSlotProps}
                        >
                            <IconButton
                                onClick={onOpenShortcutsHelp}
                                size="small"
                                aria-label={t('VIDEO_SHORTCUTS_HELP_BUTTON')}
                            >
                                <Keyboard size={20} />
                            </IconButton>
                        </Tooltip>
                    ) : null}

                    {isPictureInPictureSupported ? (
                        <Tooltip title={pictureInPictureLabel} slotProps={tooltipSlotProps}>
                            <IconButton
                                onClick={togglePictureInPicture}
                                size="small"
                                aria-label={pictureInPictureLabel}
                                aria-pressed={isPictureInPicture}
                            >
                                <PictureInPicture2 size={20} />
                            </IconButton>
                        </Tooltip>
                    ) : null}

                    <Tooltip title={fullscreenLabel} slotProps={tooltipSlotProps}>
                        <IconButton
                            onClick={toggleFullscreen}
                            size="small"
                            aria-label={fullscreenLabel}
                        >
                            {isFullscreen ? <Minimize size={20} /> : <Maximize size={20} />}
                        </IconButton>
                    </Tooltip>
                </Box>
            </Box>

            {/* Settings Menu */}
            <Menu
                anchorEl={settingsAnchor}
                open={Boolean(settingsAnchor)}
                container={portalContainer}
                onClose={handleSettingsClose}
                className="settings-menu"
            >
                <Typography variant="subtitle2" sx={{ px: 2, py: 1, fontWeight: 'bold' }}>
                    {t('VIDEO_PLAYBACK_SPEED')}
                </Typography>
                {playbackRates.map((rate) => (
                    <MenuItem
                        key={rate}
                        onClick={() => handlePlaybackRateChange(rate)}
                        selected={playbackRate === rate}
                    >
                        {rate}x
                    </MenuItem>
                ))}
            </Menu>
        </div>
    );
};

export default VideoControls;

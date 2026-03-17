import { Divider, ListItemIcon, ListItemText, Menu, MenuItem, Typography } from '@mui/material';
import { Gauge, Hd, Proportions, Settings, Subtitles, Volume2 } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import './videoSettings.css';

interface VideoSettingsProps {
    anchorEl: HTMLElement | null;
    onClose: () => void;
    playbackRate: number;
    setPlaybackRate: (rate: number) => void;
    quality: string;
    setQuality: (quality: string) => void;
}

const VideoSettings = ({
    anchorEl,
    onClose,
    playbackRate,
    setPlaybackRate,
    quality,
    setQuality,
}: VideoSettingsProps) => {
    const { t } = useI18n();
    const open = Boolean(anchorEl);

    const playbackRates = [
        { value: 0.5, label: '0.5x' },
        { value: 0.75, label: '0.75x' },
        { value: 1, label: t('VIDEO_NORMAL') },
        { value: 1.25, label: '1.25x' },
        { value: 1.5, label: '1.5x' },
        { value: 2, label: '2x' },
    ];

    const qualityOptions = [
        { value: 'auto', label: t('VIDEO_AUTO') },
        { value: '1080p', label: '1080p' },
        { value: '720p', label: '720p' },
        { value: '480p', label: '480p' },
        { value: '360p', label: '360p' },
    ];

    const handlePlaybackRateChange = (rate: number) => {
        setPlaybackRate(rate);
        onClose();
    };

    const handleQualityChange = (newQuality: string) => {
        setQuality(newQuality);
        onClose();
    };

    const handleClose = () => {
        onClose();
    };

    return (
        <Menu
            anchorEl={anchorEl}
            open={open}
            onClose={handleClose}
            classes={{
                paper: 'video-settings-menu',
                list: 'video-settings-list',
            }}
            anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
            }}
            transformOrigin={{
                vertical: 'bottom',
                horizontal: 'right',
            }}
        >
            {/* Playback Speed Section */}
            <div className="settings-section">
                <Typography variant="subtitle2" className="section-title">
                    <Gauge size={16} className="section-icon" />
                    {t('VIDEO_PLAYBACK_SPEED')}
                </Typography>
                {playbackRates.map((rate) => (
                    <MenuItem
                        key={rate.value}
                        onClick={() => handlePlaybackRateChange(rate.value)}
                        selected={playbackRate === rate.value}
                        className="settings-menu-item"
                    >
                        <ListItemText
                            primary={rate.label}
                            primaryTypographyProps={{
                                className: 'item-text',
                            }}
                        />
                        {playbackRate === rate.value && <div className="selected-indicator">✓</div>}
                    </MenuItem>
                ))}
            </div>

            <Divider className="settings-divider" />

            {/* Quality Section */}
            <div className="settings-section">
                <Typography variant="subtitle2" className="section-title">
                    <Hd size={16} className="section-icon" />
                    {t('VIDEO_QUALITY')}
                </Typography>
                {qualityOptions.map((option) => (
                    <MenuItem
                        key={option.value}
                        onClick={() => handleQualityChange(option.value)}
                        selected={quality === option.value}
                        className="settings-menu-item"
                    >
                        <ListItemText
                            primary={option.label}
                            primaryTypographyProps={{
                                className: 'item-text',
                            }}
                        />
                        {quality === option.value && <div className="selected-indicator">✓</div>}
                    </MenuItem>
                ))}
            </div>

            <Divider className="settings-divider" />

            {/* Future Features (disabled for now) */}
            <div className="settings-section">
                <Typography variant="subtitle2" className="section-title">
                    <Settings size={16} className="section-icon" />
                    {t('VIDEO_MORE_OPTIONS')}
                </Typography>

                <MenuItem disabled className="settings-menu-item disabled">
                    <ListItemIcon>
                        <Subtitles size={16} />
                    </ListItemIcon>
                    <ListItemText
                        primary={t('VIDEO_SUBTITLES')}
                        secondary={t('VIDEO_NOT_AVAILABLE_YET')}
                        primaryTypographyProps={{
                            className: 'item-text',
                        }}
                        secondaryTypographyProps={{
                            className: 'item-secondary',
                        }}
                    />
                </MenuItem>

                <MenuItem disabled className="settings-menu-item disabled">
                    <ListItemIcon>
                        <Proportions size={16} />
                    </ListItemIcon>
                    <ListItemText
                        primary={t('VIDEO_ASPECT_RATIO')}
                        secondary={t('VIDEO_AUTO_DETECTED')}
                        primaryTypographyProps={{
                            className: 'item-text',
                        }}
                        secondaryTypographyProps={{
                            className: 'item-secondary',
                        }}
                    />
                </MenuItem>

                <MenuItem disabled className="settings-menu-item disabled">
                    <ListItemIcon>
                        <Volume2 size={16} />
                    </ListItemIcon>
                    <ListItemText
                        primary={t('VIDEO_AUDIO_TRACK')}
                        secondary={t('VIDEO_DEFAULT')}
                        primaryTypographyProps={{
                            className: 'item-text',
                        }}
                        secondaryTypographyProps={{
                            className: 'item-secondary',
                        }}
                    />
                </MenuItem>
            </div>
        </Menu>
    );
};

export default VideoSettings;

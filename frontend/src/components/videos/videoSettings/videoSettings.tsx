import { useVideoPlayer } from '@/components/hooks/videoPlayerProvider/videoPlayerProvider';
import { 
	Menu,
	MenuItem,
	ListItemText,
	ListItemIcon,
	Divider,
	Tooltip,
	Typography
} from '@mui/material';
import { 
	Speed,
	HighQuality,
	Subtitles,
	VolumeUp,
	Settings,
	AspectRatio
} from 'lucide-react';
import { useState } from 'react';
import './videoSettings.css';

interface VideoSettingsProps {
	anchorEl: HTMLElement | null;
	onClose: () => void;
}

const VideoSettings = ({ anchorEl, onClose }: VideoSettingsProps) => {
	const { playbackRate, setPlaybackRate, quality, setQuality } = useVideoPlayer();
	const open = Boolean(anchorEl);

	const playbackRates = [
		{ value: 0.5, label: '0.5x' },
		{ value: 0.75, label: '0.75x' },
		{ value: 1, label: 'Normal' },
		{ value: 1.25, label: '1.25x' },
		{ value: 1.5, label: '1.5x' },
		{ value: 2, label: '2x' }
	];

	const qualityOptions = [
		{ value: 'auto', label: 'Auto' },
		{ value: '1080p', label: '1080p' },
		{ value: '720p', label: '720p' },
		{ value: '480p', label: '480p' },
		{ value: '360p', label: '360p' }
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
				list: 'video-settings-list'
			}}
			anchorOrigin={{
				vertical: 'top',
				horizontal: 'right'
			}}
			transformOrigin={{
				vertical: 'bottom',
				horizontal: 'right'
			}}
		>
			{/* Playback Speed Section */}
			<div className="settings-section">
				<Typography variant="subtitle2" className="section-title">
					<Speed size={16} className="section-icon" />
					Playback Speed
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
								className: 'item-text'
							}}
						/>
						{playbackRate === rate.value && (
							<div className="selected-indicator">✓</div>
						)}
					</MenuItem>
				))}
			</div>

			<Divider className="settings-divider" />

			{/* Quality Section */}
			<div className="settings-section">
				<Typography variant="subtitle2" className="section-title">
					<HighQuality size={16} className="section-icon" />
					Quality
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
								className: 'item-text'
							}}
						/>
						{quality === option.value && (
							<div className="selected-indicator">✓</div>
						)}
					</MenuItem>
				))}
			</div>

			<Divider className="settings-divider" />

			{/* Future Features (disabled for now) */}
			<div className="settings-section">
				<Typography variant="subtitle2" className="section-title">
					<Settings size={16} className="section-icon" />
					More Options
				</Typography>
				
				<MenuItem disabled className="settings-menu-item disabled">
					<ListItemIcon>
						<Subtitles size={16} />
					</ListItemIcon>
					<ListItemText 
						primary="Subtitles"
						secondary="Not available yet"
						primaryTypographyProps={{
							className: 'item-text'
						}}
						secondaryTypographyProps={{
							className: 'item-secondary'
						}}
					/>
				</MenuItem>

				<MenuItem disabled className="settings-menu-item disabled">
					<ListItemIcon>
						<AspectRatio size={16} />
					</ListItemIcon>
					<ListItemText 
						primary="Aspect Ratio"
						secondary="Auto detected"
						primaryTypographyProps={{
							className: 'item-text'
						}}
						secondaryTypographyProps={{
							className: 'item-secondary'
						}}
					/>
				</MenuItem>

				<MenuItem disabled className="settings-menu-item disabled">
					<ListItemIcon>
						<VolumeUp size={16} />
					</ListItemIcon>
					<ListItemText 
						primary="Audio Track"
						secondary="Default"
						primaryTypographyProps={{
							className: 'item-text'
						}}
						secondaryTypographyProps={{
							className: 'item-secondary'
						}}
					/>
				</MenuItem>
			</div>
		</Menu>
	);
};

export default VideoSettings;

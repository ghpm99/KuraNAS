import { useVideoPlayer } from '@/components/hooks/videoPlayerProvider/videoPlayerProvider';
import { IconButton, Slider, Box, Typography, Menu, MenuItem, Tooltip } from '@mui/material';
import {
	Play,
	Pause,
	VolumeOff,
	Fullscreen,
	Settings,
	Rewind,
	FastForward,
	SkipBack,
	SkipForward,
	Volume2,
} from 'lucide-react';
import { useState, useEffect, useRef } from 'react';
import './videoControls.css';

interface VideoControlsProps {
	isPlaying: boolean;
	currentTime: number;
	duration: number;
	volume: number;
	playbackRate: number;
	isFullscreen: boolean;
	pause: () => void;
	resume: () => void;
	seekTo: (time: number) => void;
	setVolume: (volume: number) => void;
	setPlaybackRate: (rate: number) => void;
	toggleFullscreen: () => void;
	togglePlayPause: () => void;
	nextVideo: () => void;
	previousVideo: () => void;
}
const VideoControls = ({
	isPlaying,
	currentTime,
	duration,
	volume,
	playbackRate,
	isFullscreen,
	pause,
	resume,
	seekTo,
	setVolume,
	setPlaybackRate,
	toggleFullscreen,
	togglePlayPause,
	nextVideo,
	previousVideo,
}: VideoControlsProps) => {
	const [showControls, setShowControls] = useState(true);
	const [settingsAnchor, setSettingsAnchor] = useState<null | HTMLElement>(null);
	const timeoutRef = useRef<NodeJS.Timeout>();

	// Auto-hide controls
	useEffect(() => {
		const resetTimer = () => {
			if (timeoutRef.current) {
				clearTimeout(timeoutRef.current);
			}
			timeoutRef.current = setTimeout(() => {
				if (isPlaying) {
					setShowControls(false);
				}
			}, 3000);
		};

		const handleMouseMove = () => {
			setShowControls(true);
			resetTimer();
		};

		const handleKeyPress = () => {
			setShowControls(true);
			resetTimer();
		};

		document.addEventListener('mousemove', handleMouseMove);
		document.addEventListener('keypress', handleKeyPress);

		return () => {
			document.removeEventListener('mousemove', handleMouseMove);
			document.removeEventListener('keypress', handleKeyPress);
			if (timeoutRef.current) {
				clearTimeout(timeoutRef.current);
			}
		};
	}, [isPlaying]);

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

	const handleSeek = (event: Event, newValue: number | number[]) => {
		seekTo(newValue as number);
	};

	const handleVolumeChange = (event: Event, newValue: number | number[]) => {
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

	const playbackRates = [0.5, 0.75, 1, 1.25, 1.5, 2];

	if (!showControls) {
		return null;
	}

	return (
		<div className='video-controls'>
			{/* Central Play/Pause Button */}
			<Box className='center-controls'>
				<IconButton onClick={skipBackward} size='large' className='skip-button'>
					<Rewind size={24} />
				</IconButton>

				<IconButton onClick={togglePlayPause} size='large' className='play-pause-button'>
					{isPlaying ? <Pause size={32} /> : <Play size={32} />}
				</IconButton>

				<IconButton onClick={skipForward} size='large' className='skip-button'>
					<FastForward size={24} />
				</IconButton>
			</Box>

			{/* Bottom Controls Bar */}
			<Box className='bottom-controls'>
				{/* Left: Time and Skip */}
				<Box className='left-controls'>
					<Typography variant='caption' className='time-display'>
						{formatTime(currentTime)}
					</Typography>
					<IconButton onClick={previousVideo} size='small'>
						<SkipBack size={20} />
					</IconButton>
					<IconButton onClick={nextVideo} size='small'>
						<SkipForward size={20} />
					</IconButton>
				</Box>

				{/* Center: Progress Bar */}
				<Box className='progress-container'>
					<Slider
						size='small'
						value={currentTime}
						max={duration || 100}
						onChange={handleSeek}
						className='progress-bar'
						sx={{ flexGrow: 1 }}
					/>
					<Typography variant='caption' className='duration-display'>
						{formatTime(duration)}
					</Typography>
				</Box>

				{/* Right: Volume and Settings */}
				<Box className='right-controls'>
					<IconButton onClick={() => setVolume(volume === 0 ? 0.5 : 0)} size='small'>
						{volume === 0 ? <VolumeOff size={20} /> : <Volume2 size={20} />}
					</IconButton>

					<Slider
						size='small'
						value={volume}
						max={1}
						step={0.1}
						onChange={handleVolumeChange}
						className='volume-slider'
						sx={{ width: 80 }}
					/>

					<IconButton onClick={handleSettingsClick} size='small'>
						<Settings size={20} />
					</IconButton>

					<Tooltip title={isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'}>
						<IconButton onClick={toggleFullscreen} size='small'>
							{isFullscreen ? <Fullscreen size={20} /> : <Fullscreen size={20} />}
						</IconButton>
					</Tooltip>
				</Box>
			</Box>

			{/* Settings Menu */}
			<Menu
				anchorEl={settingsAnchor}
				open={Boolean(settingsAnchor)}
				onClose={handleSettingsClose}
				className='settings-menu'
			>
				<Typography variant='subtitle2' sx={{ px: 2, py: 1, fontWeight: 'bold' }}>
					Playback Speed
				</Typography>
				{playbackRates.map((rate) => (
					<MenuItem key={rate} onClick={() => handlePlaybackRateChange(rate)} selected={playbackRate === rate}>
						{rate}x
					</MenuItem>
				))}
			</Menu>
		</div>
	);
};

export default VideoControls;

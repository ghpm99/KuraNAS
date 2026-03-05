import { Box, Card, CardContent, IconButton, Slider, Typography } from '@mui/material';
import { Pause, Play, Repeat, Repeat1, Shuffle, SkipBack, SkipForward, Volume2 } from 'lucide-react';
import { useGlobalMusic } from '../providers/GlobalMusicProvider';
import '../playerControl/playerControl.css';

const GlobalPlayerControl = () => {
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
	} = useGlobalMusic();

	if (!hasQueue) return null;

	const formatTime = (time: number): string => {
		if (isNaN(time)) return '0:00';
		const minutes = Math.floor(time / 60);
		const seconds = Math.floor(time % 60);
		return `${minutes}:${seconds.toString().padStart(2, '0')}`;
	};

	const cycleRepeatMode = () => {
		const modes = ['none', 'all', 'one'] as const;
		const currentIdx = modes.indexOf(repeatMode);
		const nextMode = currentIdx === -1 ? modes[0] : (modes[(currentIdx + 1) % modes.length] ?? modes[0]);
		setRepeatMode(nextMode);
	};

	const RepeatIcon = repeatMode === 'one' ? Repeat1 : Repeat;

	return (
		<Card
			className='player-control'
			sx={{ position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 1300 }}
		>
			<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2, p: 2 }}>
				{/* Track Info */}
				<Box sx={{ display: 'flex', alignItems: 'center', gap: 2, minWidth: 200 }}>
					<Box
						sx={{
							width: 50,
							height: 50,
							bgcolor: 'primary.main',
							borderRadius: 1,
							display: 'flex',
							alignItems: 'center',
							justifyContent: 'center',
							flexShrink: 0,
						}}
					>
						<Volume2 size={24} color='white' />
					</Box>
					<Box sx={{ minWidth: 0 }}>
						<Typography variant='subtitle2' noWrap>
							{currentTrack?.metadata?.title || currentTrack?.name || 'No track playing'}
						</Typography>
						<Typography variant='caption' color='text.secondary' noWrap>
							{currentTrack?.metadata?.artist || 'Unknown Artist'}
						</Typography>
					</Box>
				</Box>

				{/* Controls */}
				<Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
					<IconButton onClick={toggleShuffle} size='small' sx={{ opacity: shuffle ? 1 : 0.5 }}>
						<Shuffle size={18} />
					</IconButton>
					<IconButton onClick={previous} size='small'>
						<SkipBack size={20} />
					</IconButton>
					<IconButton
						onClick={togglePlayPause}
						sx={{ bgcolor: 'primary.main', '&:hover': { bgcolor: 'primary.dark' } }}
					>
						{isPlaying ? <Pause size={20} color='white' /> : <Play size={20} color='white' />}
					</IconButton>
					<IconButton onClick={next} size='small'>
						<SkipForward size={20} />
					</IconButton>
					<IconButton
						onClick={cycleRepeatMode}
						size='small'
						sx={{ opacity: repeatMode !== 'none' ? 1 : 0.5 }}
					>
						<RepeatIcon size={18} />
					</IconButton>
				</Box>

				{/* Progress */}
				<Box sx={{ display: 'flex', alignItems: 'center', gap: 2, flexGrow: 1, minWidth: 200 }}>
					<Typography variant='caption' sx={{ minWidth: 40 }}>
						{formatTime(currentTime)}
					</Typography>
					<Slider
						size='small'
						value={currentTime}
						max={duration || 100}
						onChange={(_, value) => seek(value as number)}
						sx={{ flexGrow: 1 }}
					/>
					<Typography variant='caption' sx={{ minWidth: 40 }}>
						{formatTime(duration)}
					</Typography>
				</Box>

				{/* Volume */}
				<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 120 }}>
					<Volume2 size={20} />
					<Slider
						size='small'
						value={volume}
						max={1}
						step={0.1}
						onChange={(_, value) => setVolume(value as number)}
						sx={{ width: 80 }}
					/>
				</Box>
			</CardContent>
		</Card>
	);
};

export default GlobalPlayerControl;

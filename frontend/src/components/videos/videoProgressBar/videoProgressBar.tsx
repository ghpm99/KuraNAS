import { Box, Typography } from '@mui/material';
import { useEffect, useRef, useState } from 'react';
import './videoProgressBar.css';

interface VideoProgressBarProps {
	className?: string;
	currentTime: number;
	duration: number;
	seekTo: (time: number) => void;
}

const VideoProgressBar = ({ className = '', currentTime, duration, seekTo }: VideoProgressBarProps) => {
	const [isDragging, setIsDragging] = useState(false);
	const [hoverTime, setHoverTime] = useState<number | null>(null);
	const [hoverPosition, setHoverPosition] = useState<number | null>(null);
	const progressBarRef = useRef<HTMLDivElement>(null);

	const formatTime = (time: number): string => {
		if (isNaN(time)) return '0:00';
		const minutes = Math.floor(time / 60);
		const seconds = Math.floor(time % 60);
		return `${minutes}:${seconds.toString().padStart(2, '0')}`;
	};

	const handleMouseDown = () => {
		setIsDragging(true);
	};

	const handleMouseUp = () => {
		setIsDragging(false);
	};

	const handleMouseMove = (event: React.MouseEvent<HTMLDivElement>) => {
		if (!progressBarRef.current) return;

		const rect = progressBarRef.current.getBoundingClientRect();
		const x = event.clientX - rect.left;
		const percentage = Math.max(0, Math.min(1, x / rect.width));
		const time = percentage * (duration || 0);

		setHoverTime(time);
		setHoverPosition(x);
	};

	const handleMouseLeave = () => {
		setHoverTime(null);
		setHoverPosition(null);
	};

	const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
		if (!progressBarRef.current) return;

		const rect = progressBarRef.current.getBoundingClientRect();
		const x = event.clientX - rect.left;
		const percentage = Math.max(0, Math.min(1, x / rect.width));
		const time = percentage * (duration || 0);

		seekTo(time);
	};

	// Global mouse events for dragging
	useEffect(() => {
		if (!isDragging) return;

		const handleGlobalMouseMove = (event: MouseEvent) => {
			if (!progressBarRef.current) return;

			const rect = progressBarRef.current.getBoundingClientRect();
			const x = event.clientX - rect.left;
			const percentage = Math.max(0, Math.min(1, x / rect.width));
			const time = percentage * (duration || 0);

			seekTo(time);
			setHoverTime(time);
			setHoverPosition(x);
		};

		const handleGlobalMouseUp = () => {
			setIsDragging(false);
		};

		document.addEventListener('mousemove', handleGlobalMouseMove);
		document.addEventListener('mouseup', handleGlobalMouseUp);

		return () => {
			document.removeEventListener('mousemove', handleGlobalMouseMove);
			document.removeEventListener('mouseup', handleGlobalMouseUp);
		};
	}, [isDragging, duration, seekTo]);

	const progressPercentage = duration ? (currentTime / duration) * 100 : 0;

	return (
		<Box
			className={`video-progress-bar ${className}`}
			ref={progressBarRef}
			onMouseMove={handleMouseMove}
			onMouseLeave={handleMouseLeave}
			onMouseDown={handleMouseDown}
			onMouseUp={handleMouseUp}
			onClick={handleClick}
		>
			{/* Progress Track */}
			<Box className='progress-track'>
				{/* Buffered portion (simplified - could be enhanced with actual buffering data) */}
				<Box className='progress-buffered' sx={{ width: `${progressPercentage}%` }} />

				{/* Played portion */}
				<Box className='progress-played' sx={{ width: `${progressPercentage}%` }} />
			</Box>

			{/* Hover Time Tooltip */}
			{hoverTime !== null && hoverPosition !== null && (
				<Box className='hover-tooltip' sx={{ left: `${hoverPosition}px` }}>
					<Typography variant='caption' className='tooltip-time'>
						{formatTime(hoverTime)}
					</Typography>
				</Box>
			)}

			{/* Current Time Indicator */}
			<Box className='progress-handle' sx={{ left: `${progressPercentage}%` }} />

			{/* Time Display */}
			<Box className='time-display'>
				<Typography variant='caption' className='current-time'>
					{formatTime(currentTime)}
				</Typography>
				<Typography variant='caption' className='total-time'>
					{formatTime(duration)}
				</Typography>
			</Box>
		</Box>
	);
};

export default VideoProgressBar;

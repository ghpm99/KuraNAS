import { useRef, useState } from 'react';

type Status = 'waiting' | 'playing' | 'paused' | 'stopped';

const useVideoPlayer = ({ videoId }: { videoId: string }) => {
	const [status, setStatus] = useState<Status>('waiting');
	const [currentTime, setCurrentTime] = useState(0);
	const [duration, setDuration] = useState(0);
	const [volume, setVolume] = useState(1);
	const [playbackRate, setPlaybackRate] = useState(1);
	const [quality, setQuality] = useState('auto');
	const [isFullscreen, setIsFullscreen] = useState(false);
	const videoRef = useRef<HTMLVideoElement>(null);

	const playVideo = () => {
		console.log('Playing video:', videoId, videoRef.current);
		if (videoRef.current) {
			videoRef.current.src = `${import.meta.env.VITE_API_URL}/api/v1/files/video-stream/${videoId}`;
			videoRef.current.play();
			setStatus('playing');
		}
	};

	const pause = () => {
		if (videoRef.current) {
			videoRef.current.pause();
			setStatus('paused');
		}
	};

	const resume = () => {
		console.log('Resuming video:', videoId, videoRef.current);
		if (videoRef.current) {
			videoRef.current.play();
			setStatus('playing');
		}
	};

	const seekTo = (time: number) => {
		if (videoRef.current) {
			videoRef.current.currentTime = time;
		}
	};

	const setVolumeHandler = (newVolume: number) => {
		if (videoRef.current) {
			videoRef.current.volume = Math.max(0, Math.min(1, newVolume));
			setVolume(Math.max(0, Math.min(1, newVolume)));
		}
	};

	const setPlaybackRateHandler = (rate: number) => {
		if (videoRef.current) {
			videoRef.current.playbackRate = rate;
			setPlaybackRate(rate);
		}
	};

	const toggleFullscreen = () => {
		if (!document.fullscreenElement) {
			videoRef.current?.requestFullscreen();
			setIsFullscreen(true);
		} else {
			document.exitFullscreen();
			setIsFullscreen(false);
		}
	};

	const togglePlayPause = () => {
		console.log('Toggling play/pause. Current status:', status);
		if (status === 'playing') {
			pause();
		} else {
			resume();
		}
	};

	return {
		videoRef,
		playVideo,
		pause,
		resume,
		seekTo,
		setVolume: setVolumeHandler,
		setPlaybackRate: setPlaybackRateHandler,
		toggleFullscreen,
		togglePlayPause,
		status,
		currentTime,
		duration,
		volume,
		playbackRate,
		isFullscreen,
		setCurrentTime,
		setDuration,
		quality,
		setQuality,
	};
};

export default useVideoPlayer;

import { createContext, useContext, useState, ReactNode, useRef } from 'react';
import { IVideoData, IVideoPlayerContext } from '@/types/video';
import { useNavigate } from 'react-router-dom';

const VideoPlayerContext = createContext<IVideoPlayerContext | undefined>(undefined);

export const VideoPlayerProvider = ({ children }: { children: ReactNode }) => {
	const navigate = useNavigate();
	const [currentVideo, setCurrentVideo] = useState<IVideoData | null>(null);
	const [isPlaying, setIsPlaying] = useState(false);
	const [currentTime, setCurrentTime] = useState(0);
	const [duration, setDuration] = useState(0);
	const [volume, setVolume] = useState(1);
	const [playbackRate, setPlaybackRate] = useState(1);
	const [quality, setQuality] = useState('auto');
	const [isFullscreen, setIsFullscreen] = useState(false);
	const [playlist, setPlaylist] = useState<IVideoData[]>([]);
	const videoRef = useRef<HTMLVideoElement>(null);

	const playVideo = (video: IVideoData) => {
		console.log('Playing video:', video, videoRef.current);
		navigate(`/video/${video.id}`);
		if (videoRef.current) {
			setCurrentVideo(video);
			videoRef.current.src = `${import.meta.env.VITE_API_URL}/api/v1/files/video-stream/${video.id}`;
			videoRef.current.play();
			setIsPlaying(true);
		}
	};

	const pause = () => {
		if (videoRef.current) {
			videoRef.current.pause();
			setIsPlaying(false);
		}
	};

	const resume = () => {
		if (videoRef.current) {
			videoRef.current.play();
			setIsPlaying(true);
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

	const nextVideo = () => {
		if (playlist.length === 0) return;

		const currentIndex = currentVideo ? playlist.findIndex((video) => video.id === currentVideo.id) : -1;
		const nextIndex = currentIndex === playlist.length - 1 ? 0 : currentIndex + 1;
		playVideo(playlist[nextIndex]);
	};

	const previousVideo = () => {
		if (playlist.length === 0) return;

		const currentIndex = currentVideo ? playlist.findIndex((video) => video.id === currentVideo.id) : -1;
		const prevIndex = currentIndex === 0 ? playlist.length - 1 : currentIndex - 1;
		playVideo(playlist[prevIndex]);
	};

	const togglePlayPause = () => {
		if (isPlaying) {
			pause();
		} else {
			resume();
		}
	};

	const setVideoRef = (ref: HTMLVideoElement | null) => {
		console.log('Setting video ref:', ref);
		videoRef.current = ref;
	};

	const value: IVideoPlayerContext = {
		currentVideo,
		isPlaying,
		currentTime,
		duration,
		volume,
		playbackRate,
		quality,
		isFullscreen,
		playlist,
		playVideo,
		pause,
		resume,
		seekTo,
		setVolume: setVolumeHandler,
		setPlaybackRate: setPlaybackRateHandler,
		toggleFullscreen,
		nextVideo,
		previousVideo,
		togglePlayPause,
		setVideoRef,
		setCurrentTime,
		setDuration,
		setPlaylist,
	};

	return <VideoPlayerContext.Provider value={value}>{children}</VideoPlayerContext.Provider>;
};

export const useVideoPlayer = () => {
	const context = useContext(VideoPlayerContext);
	if (!context) {
		throw new Error('useVideoPlayer must be used within a VideoPlayerProvider');
	}
	return context;
};

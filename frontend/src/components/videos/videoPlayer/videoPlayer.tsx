import { IVideoData } from '@/types/video';
import { useEffect } from 'react';
import './videoPlayer.css';

interface VideoPlayerProps {
	currentVideo: IVideoData | null;

	volume: number;
	playbackRate: number;
	videoRef: React.RefObject<HTMLVideoElement | null>;
	setCurrentTime: (time: number) => void;
	setDuration: (duration: number) => void;
	nextVideo: () => void;
}
const VideoPlayer = ({
	currentVideo,
	volume,
	playbackRate,
	videoRef,
	setCurrentTime,
	setDuration,
	nextVideo,
}: VideoPlayerProps) => {
	useEffect(() => {
		const video = videoRef.current;
		if (!video) return;

		const updateTime = () => setCurrentTime(video.currentTime);
		const updateDuration = () => setDuration(video.duration);
		const handleEnded = () => nextVideo();
		const handleCanPlay = () => {
			// Video está pronto para tocar
		};
		const handleWaiting = () => {
			// Video está carregando/buffering
		};
		const handlePlaying = () => {
			// Video começou a tocar
		};

		video.addEventListener('timeupdate', updateTime);
		video.addEventListener('loadedmetadata', updateDuration);
		video.addEventListener('ended', handleEnded);
		video.addEventListener('canplay', handleCanPlay);
		video.addEventListener('waiting', handleWaiting);
		video.addEventListener('playing', handlePlaying);

		return () => {
			video.removeEventListener('timeupdate', updateTime);
			video.removeEventListener('loadedmetadata', updateDuration);
			video.removeEventListener('ended', handleEnded);
			video.removeEventListener('canplay', handleCanPlay);
			video.removeEventListener('waiting', handleWaiting);
			video.removeEventListener('playing', handlePlaying);
		};
	}, [nextVideo, setCurrentTime, setDuration]);

	useEffect(() => {
		if (videoRef.current) {
			videoRef.current.volume = volume;
		}
	}, [volume]);

	useEffect(() => {
		if (videoRef.current) {
			videoRef.current.playbackRate = playbackRate;
		}
	}, [playbackRate]);

	const getVideoTitle = (): string => {
		if (!currentVideo) return 'No video playing';
		return currentVideo.name;
	};

	const getVideoMetadata = (): string => {
		if (!currentVideo?.metadata) return '';
		const metadata = currentVideo.metadata;
		const parts = [];

		if (metadata.width && metadata.height) {
			parts.push(`${metadata.width}x${metadata.height}`);
		}
		if (metadata.duration) {
			parts.push(metadata.duration);
		}
		if (metadata.codec_name) {
			parts.push(metadata.codec_name.toUpperCase());
		}

		return parts.join(' • ');
	};

	return (
		<div className='video-player'>
			<div className='video-container'>
				<video ref={videoRef} className='video-element' preload='metadata' playsInline />

				<div className='video-overlay'>
					<div className='video-info'>
						<h3 className='video-title'>{getVideoTitle()}</h3>
						{getVideoMetadata() && <p className='video-metadata'>{getVideoMetadata()}</p>}
					</div>
				</div>
			</div>
		</div>
	);
};

export default VideoPlayer;

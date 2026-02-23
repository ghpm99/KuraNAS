import useVideoPlayer from '@/components/hooks/useVideoPlayer/useVideoPlayer';
import VideoControls from '@/components/videos/videoControls/videoControls';
import VideoPlayer from '@/components/videos/videoPlayer/videoPlayer';
import VideoProgressBar from '@/components/videos/videoProgressBar/videoProgressBar';
import VideoSettings from '@/components/videos/videoSettings/videoSettings';
import { useEffect } from 'react';
import { useParams } from 'react-router-dom';

const VideoPlayerPage = () => {
	const { id } = useParams();

	if (!id || typeof id !== 'string') {
		return <div>Invalid video ID</div>;
	}

	const {
		videoRef,
		playVideo,
		pause,
		resume,
		seekTo,
		setVolume,
		setPlaybackRate,
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
	} = useVideoPlayer({ videoId: id });

	useEffect(() => {
		playVideo();
	}, []);

	return (
		<>
			<VideoSettings
				anchorEl={null}
				onClose={() => {}}
				playbackRate={playbackRate}
				quality={quality}
				setPlaybackRate={setPlaybackRate}
				setQuality={setQuality}
			/>
			<VideoControls
				currentTime={currentTime}
				duration={duration}
				isFullscreen={isFullscreen}
				isPlaying={status === 'playing'}
				volume={volume}
				playbackRate={playbackRate}
				pause={pause}
				resume={resume}
				seekTo={seekTo}
				setVolume={setVolume}
				setPlaybackRate={setPlaybackRate}
				toggleFullscreen={toggleFullscreen}
				togglePlayPause={togglePlayPause}
				nextVideo={() => {}}
				previousVideo={() => {}}
			/>
			<VideoPlayer
				currentVideo={null}
				volume={volume}
				playbackRate={playbackRate}
				videoRef={videoRef}
				setCurrentTime={setCurrentTime}
				setDuration={setDuration}
				nextVideo={() => {}}
			/>
			<VideoProgressBar currentTime={currentTime} duration={duration} seekTo={seekTo} />
		</>
	);
};

export default VideoPlayerPage;

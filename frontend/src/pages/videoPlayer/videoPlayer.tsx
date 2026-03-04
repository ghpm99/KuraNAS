import useVideoPlayer from '@/components/hooks/useVideoPlayer/useVideoPlayer';
import VideoControls from '@/components/videos/videoControls/videoControls';
import VideoPlayer from '@/components/videos/videoPlayer/videoPlayer';
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
		seekTo,
		setVolume,
		setPlaybackRate,
		toggleFullscreen,
		togglePlayPause,
		nextVideo,
		previousVideo,
		status,
		currentTime,
		duration,
		volume,
		playbackRate,
		isFullscreen,
		setCurrentTime,
		setDuration,
		currentVideo,
	} = useVideoPlayer({ videoId: id });

	useEffect(() => {
		playVideo();
	}, [playVideo]);

	return (
		<>
			{/* <VideoSettings
				anchorEl={null}
				onClose={() => {}}
				playbackRate={playbackRate}
				quality={quality}
				setPlaybackRate={setPlaybackRate}
				setQuality={setQuality}
			/> */}
			<VideoControls
				currentTime={currentTime}
				duration={duration}
				isFullscreen={isFullscreen}
				isPlaying={status === 'playing'}
				volume={volume}
				playbackRate={playbackRate}
				seekTo={seekTo}
				setVolume={setVolume}
				setPlaybackRate={setPlaybackRate}
				toggleFullscreen={toggleFullscreen}
				togglePlayPause={togglePlayPause}
				nextVideo={nextVideo}
				previousVideo={previousVideo}
			/>
			<VideoPlayer
				currentVideo={currentVideo}
				volume={volume}
				playbackRate={playbackRate}
				videoRef={videoRef}
				setCurrentTime={setCurrentTime}
				setDuration={setDuration}
				nextVideo={nextVideo}
			/>
			{/* <VideoProgressBar currentTime={currentTime} duration={duration} seekTo={seekTo} /> */}
		</>
	);
};

export default VideoPlayerPage;

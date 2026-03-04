import useVideoPlayer from '@/components/hooks/useVideoPlayer/useVideoPlayer';
import VideoControls from '@/components/videos/videoControls/videoControls';
import VideoPlayer from '@/components/videos/videoPlayer/videoPlayer';
import { useEffect } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

const VideoPlayerPage = () => {
	const { id } = useParams();
	const navigate = useNavigate();
	const location = useLocation();
	const playlistId = (location.state as { playlistId?: number } | null)?.playlistId;
	const videoId = typeof id === 'string' ? id : '';

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
		} = useVideoPlayer({ videoId, playlistId: playlistId ?? null });

	useEffect(() => {
		if (!videoId) return;
		playVideo();
	}, [playVideo, videoId]);

	if (!videoId) {
		return <div>Invalid video ID</div>;
	}

	const handleBack = () => {
		const fromState = (location.state as { from?: string } | null)?.from;
		if (fromState) {
			navigate(fromState);
			return;
		}
		if (window.history.length > 1) {
			navigate(-1);
			return;
		}
		navigate('/videos');
	};

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
				onBack={handleBack}
			/>
			{/* <VideoProgressBar currentTime={currentTime} duration={duration} seekTo={seekTo} /> */}
		</>
	);
};

export default VideoPlayerPage;

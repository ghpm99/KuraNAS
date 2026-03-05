import {
	nextVideoPlayback,
	previousVideoPlayback,
	startVideoPlayback,
	updateVideoPlaybackState,
	VideoPlaybackSessionDto,
} from '@/service/videoPlayback';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

type Status = 'waiting' | 'playing' | 'paused' | 'stopped';

const useVideoPlayer = ({ videoId, playlistId }: { videoId: string; playlistId?: number | null }) => {
	const [status, setStatus] = useState<Status>('waiting');
	const [currentTime, setCurrentTime] = useState(0);
	const [duration, setDuration] = useState(0);
	const [volume, setVolume] = useState(1);
	const [playbackRate, setPlaybackRate] = useState(1);
	const [quality, setQuality] = useState('auto');
	const [isFullscreen, setIsFullscreen] = useState(false);
	const [session, setSession] = useState<VideoPlaybackSessionDto | null>(null);
	const videoRef = useRef<HTMLVideoElement>(null);
	const syncTimerRef = useRef<number | null>(null);
	const latestPlaybackRef = useRef({
		currentTime: 0,
		duration: 0,
		status: 'waiting' as Status,
		session: null as VideoPlaybackSessionDto | null,
	});

	useEffect(() => {
		latestPlaybackRef.current = {
			currentTime,
			duration,
			status,
			session,
		};
	}, [currentTime, duration, status, session]);

	const currentVideo = useMemo(() => {
		if (!session?.playback_state.video_id) return null;
		return session.playlist.items.find((item) => item.video.id === session.playback_state.video_id)?.video ?? null;
	}, [session]);

	const attachVideoSource = useCallback(
		(videoToPlayId: number, seekSeconds?: number) => {
			if (!videoRef.current) return;
			videoRef.current.src = `${getApiV1BaseUrl()}/files/video-stream/${videoToPlayId}`;
			if (typeof seekSeconds === 'number' && seekSeconds > 0) {
				videoRef.current.currentTime = seekSeconds;
			}
			videoRef.current
				.play()
				.then(() => setStatus('playing'))
				.catch(() => setStatus('paused'));
		},
		[],
	);

	const syncState = useCallback(
		async (payload?: Partial<{ currentTime: number; duration: number; isPaused: boolean; completed: boolean }>) => {
			const latest = latestPlaybackRef.current;
			if (!latest.session?.playback_state.playlist_id || !latest.session.playback_state.video_id) return;
			try {
				await updateVideoPlaybackState({
					playlist_id: latest.session.playback_state.playlist_id,
					video_id: latest.session.playback_state.video_id,
					current_time: payload?.currentTime ?? latest.currentTime,
					duration: payload?.duration ?? latest.duration,
					is_paused: payload?.isPaused ?? latest.status !== 'playing',
					completed: payload?.completed ?? false,
				});
			} catch {
				// best effort sync
			}
		},
		[],
	);

	const playVideo = useCallback(async () => {
		const response = await startVideoPlayback(Number(videoId), playlistId ?? null);
		setSession(response);
		setCurrentTime(response.playback_state.current_time || 0);
		setDuration(response.playback_state.duration || 0);
		if (response.playback_state.video_id) {
			attachVideoSource(response.playback_state.video_id, response.playback_state.current_time || 0);
		}
	}, [attachVideoSource, playlistId, videoId]);

	const pause = useCallback(() => {
		if (videoRef.current) {
			videoRef.current.pause();
			setStatus('paused');
			syncState({ isPaused: true });
		}
	}, [syncState]);

	const resume = useCallback(() => {
		if (videoRef.current) {
			videoRef.current.play();
			setStatus('playing');
			syncState({ isPaused: false });
		}
	}, [syncState]);

	const seekTo = useCallback((time: number) => {
		if (videoRef.current) {
			videoRef.current.currentTime = time;
			setCurrentTime(time);
		}
	}, []);

	const setVolumeHandler = useCallback((newVolume: number) => {
		if (videoRef.current) {
			videoRef.current.volume = Math.max(0, Math.min(1, newVolume));
			setVolume(Math.max(0, Math.min(1, newVolume)));
		}
	}, []);

	const setPlaybackRateHandler = useCallback((rate: number) => {
		if (videoRef.current) {
			videoRef.current.playbackRate = rate;
			setPlaybackRate(rate);
		}
	}, []);

	const toggleFullscreen = useCallback(() => {
		if (!document.fullscreenElement) {
			videoRef.current?.requestFullscreen();
			setIsFullscreen(true);
		} else {
			document.exitFullscreen();
			setIsFullscreen(false);
		}
	}, []);

	const nextVideo = useCallback(async () => {
		const response = await nextVideoPlayback();
		setSession(response);
		setCurrentTime(0);
		setDuration(0);
		if (response.playback_state.video_id) {
			attachVideoSource(response.playback_state.video_id, 0);
		}
	}, [attachVideoSource]);

	const previousVideo = useCallback(async () => {
		const response = await previousVideoPlayback();
		setSession(response);
		setCurrentTime(0);
		setDuration(0);
		if (response.playback_state.video_id) {
			attachVideoSource(response.playback_state.video_id, 0);
		}
	}, [attachVideoSource]);

	const togglePlayPause = useCallback(() => {
		if (status === 'playing') {
			pause();
		} else {
			resume();
		}
	}, [pause, resume, status]);

	const setCurrentTimeHandler = useCallback((time: number) => {
		setCurrentTime(time);
	}, []);

	const setDurationHandler = useCallback((newDuration: number) => {
		setDuration(newDuration);
	}, []);

	const onVideoEnded = useCallback(() => {
		syncState({ completed: true, isPaused: true, currentTime: duration });
		nextVideo();
	}, [duration, nextVideo, syncState]);

	useEffect(() => {
		syncTimerRef.current = window.setInterval(() => {
			syncState();
		}, 5000);

		return () => {
			if (syncTimerRef.current != null) {
				window.clearInterval(syncTimerRef.current);
				syncTimerRef.current = null;
			}
		};
	}, [syncState]);

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
		nextVideo,
		previousVideo,
		status,
		currentTime,
		duration,
		volume,
		playbackRate,
		isFullscreen,
		setCurrentTime: setCurrentTimeHandler,
		setDuration: setDurationHandler,
		quality,
		setQuality,
		playlistItems: session?.playlist.items ?? [],
		currentVideo,
		onVideoEnded,
	};
};

export default useVideoPlayer;

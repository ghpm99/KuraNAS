import { RefObject, useCallback, useEffect, useRef, useState } from 'react';
import { useMusic } from '../musicProvider/musicProvider';

type Status = 'waiting' | 'playing' | 'paused' | 'stopped';

export interface IMusicPlayerContext {
	status: Status;
	isPlaying: boolean;
	currentTime: number;
	duration: number;
	volume: number;
	playTrack: (trackIndex: number) => void;
	pause: () => void;
	resume: () => void;
	next: () => void;
	previous: () => void;
	seek: (time: number) => void;
	setVolume: (volume: number) => void;
	togglePlayPause: () => void;
	audioRef: RefObject<HTMLAudioElement | null>;
	setCurrentTime: (time: number) => void;
	setDuration: (duration: number) => void;
}

const useMusicPlayer = (): IMusicPlayerContext => {
	const { currentTrack, playlist, setCurrentTrack } = useMusic();

	const [status, setStatus] = useState<Status>('waiting');

	const [currentTime, setCurrentTime] = useState(0);
	const [duration, setDuration] = useState(0);
	const [volume, setVolume] = useState(1);

	const audioRef = useRef<HTMLAudioElement>(null);

	const playTrack = useCallback(
		(trackIndex: number) => {
			if (audioRef.current && playlist.length > trackIndex && trackIndex >= 0) {
				setCurrentTrack(trackIndex);
				const track = playlist[trackIndex];
				if (!track) return;
				audioRef.current.src = `${import.meta.env.VITE_API_URL}/api/v1/files/stream/${track.id}`;
				audioRef.current.play();
				setStatus('playing');
			}
		},
		[playlist, setCurrentTrack],
	);

	useEffect(() => {
		if (currentTrack !== undefined && status === 'waiting') {
			playTrack(currentTrack);
		}
	}, [currentTrack, playTrack, status]);

	const pause = () => {
		if (audioRef.current) {
			audioRef.current.pause();
			setStatus('paused');
		}
	};

	const resume = () => {
		if (audioRef.current) {
			audioRef.current.play();
			setStatus('playing');
		}
	};

	const togglePlayPause = () => {
		if (status === 'playing') {
			pause();
		} else {
			resume();
		}
	};

	const next = () => {
		if (playlist.length === 0 || currentTrack === undefined) return;

		const nextIndex = (currentTrack + 1) % playlist.length;
		playTrack(nextIndex);
	};

	const previous = () => {
		if (playlist.length === 0 || currentTrack === undefined) return;

		const prevIndex = currentTrack === 0 ? playlist.length - 1 : currentTrack - 1;
		playTrack(prevIndex);
	};

	const seek = (time: number) => {
		if (audioRef.current) {
			audioRef.current.currentTime = time;
		}
	};

	const setVolumeHandler = (newVolume: number) => {
		if (audioRef.current) {
			audioRef.current.volume = Math.max(0, Math.min(1, newVolume));
			setVolume(Math.max(0, Math.min(1, newVolume)));
		}
	};

	return {
		status,
		currentTime,
		duration,
		volume,
		playTrack,
		pause,
		resume,
		next,
		previous,
		seek,
		setVolume: setVolumeHandler,
		togglePlayPause,
		audioRef,
		setCurrentTime,
		setDuration,
	};
};

export default useMusicPlayer;

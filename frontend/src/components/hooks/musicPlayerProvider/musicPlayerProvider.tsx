import React, { createContext, useContext, useState, ReactNode, useRef, RefObject } from 'react';

export interface IMusicData {
	id: number;
	name: string;
	path: string;
	format: string;
	size: number;
	metadata?: {
		id: number;
		title: string;
		artist: string;
		album: string;
		duration: number;
	};
}

export interface IMusicPlayerContext {
	currentTrack: IMusicData | null;
	isPlaying: boolean;
	currentTime: number;
	duration: number;
	volume: number;
	playlist: IMusicData[];
	playTrack: (track: IMusicData) => void;
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

const MusicPlayerContext = createContext<IMusicPlayerContext | undefined>(undefined);

export const MusicPlayerProvider = ({ children }: { children: ReactNode }) => {
	const [currentTrack, setCurrentTrack] = useState<IMusicData | null>(null);
	const [isPlaying, setIsPlaying] = useState(false);
	const [currentTime, setCurrentTime] = useState(0);
	const [duration, setDuration] = useState(0);
	const [volume, setVolume] = useState(1);
	const [playlist, setPlaylist] = useState<IMusicData[]>([]);

	const audioRef = useRef<HTMLAudioElement>(null);

	const playTrack = (track: IMusicData) => {
		if (audioRef.current) {
			setCurrentTrack(track);
			audioRef.current.src = `${import.meta.env.VITE_API_URL}/api/v1/files/stream/${track.id}`;
			audioRef.current.play();
			setIsPlaying(true);
		}
	};

	const pause = () => {
		if (audioRef.current) {
			audioRef.current.pause();
			setIsPlaying(false);
		}
	};

	const resume = () => {
		if (audioRef.current) {
			audioRef.current.play();
			setIsPlaying(true);
		}
	};

	const togglePlayPause = () => {
		if (isPlaying) {
			pause();
		} else {
			resume();
		}
	};

	const next = () => {
		if (playlist.length === 0) return;

		const currentIndex = currentTrack ? playlist.findIndex((track) => track.id === currentTrack.id) : -1;
		const nextIndex = (currentIndex + 1) % playlist.length;
		playTrack(playlist[nextIndex]);
	};

	const previous = () => {
		if (playlist.length === 0) return;

		const currentIndex = currentTrack ? playlist.findIndex((track) => track.id === currentTrack.id) : -1;
		const prevIndex = currentIndex === 0 ? playlist.length - 1 : currentIndex - 1;
		playTrack(playlist[prevIndex]);
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

	const value: IMusicPlayerContext = {
		currentTrack,
		isPlaying,
		currentTime,
		duration,
		volume,
		playlist,
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

	return <MusicPlayerContext.Provider value={value}>{children}</MusicPlayerContext.Provider>;
};

export const useMusicPlayer = () => {
	const context = useContext(MusicPlayerContext);
	if (!context) {
		throw new Error('useMusicPlayer must be used within a MusicPlayerProvider');
	}
	return context;
};

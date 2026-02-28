import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import { IMusicData, IMusicMetadata } from '../hooks/musicProvider/musicProvider';
import { updatePlayerState } from '@/service/playerState';
import { formatSize } from '@/utils';

type RepeatMode = 'none' | 'all' | 'one';

export interface IGlobalMusicContext {
	// Queue
	queue: IMusicData[];
	currentIndex: number | undefined;
	addToQueue: (track: IMusicData) => void;
	playTrackFromQueue: (index: number) => void;
	clearQueue: () => void;

	// Playback
	isPlaying: boolean;
	currentTime: number;
	duration: number;
	volume: number;
	shuffle: boolean;
	repeatMode: RepeatMode;

	// Controls
	togglePlayPause: () => void;
	next: () => void;
	previous: () => void;
	seek: (time: number) => void;
	setVolume: (volume: number) => void;
	toggleShuffle: () => void;
	setRepeatMode: (mode: RepeatMode) => void;

	// Helpers
	currentTrack: IMusicData | undefined;
	hasQueue: boolean;
	getMusicTitle: (music: IMusicData) => string;
	getMusicArtist: (music: IMusicData) => string;
	musicMetadata: (music: { format: string; size: number; metadata?: IMusicMetadata }) => string;
	formatDuration: (seconds: number) => string;
}

const GlobalMusicContext = createContext<IGlobalMusicContext | undefined>(undefined);

export const GlobalMusicProvider = ({ children }: { children: React.ReactNode }) => {
	const [queue, setQueue] = useState<IMusicData[]>([]);
	const [currentIndex, setCurrentIndex] = useState<number | undefined>(undefined);
	const [isPlaying, setIsPlaying] = useState(false);
	const [currentTime, setCurrentTime] = useState(0);
	const [duration, setDuration] = useState(0);
	const [volume, setVolumeState] = useState(1);
	const [shuffle, setShuffle] = useState(false);
	const [repeatMode, setRepeatMode] = useState<RepeatMode>('none');

	const audioRef = useRef<HTMLAudioElement | null>(null);
	const syncTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	// Initialize audio element once
	useEffect(() => {
		const audio = new Audio();
		audio.volume = volume;
		audioRef.current = audio;

		const onTimeUpdate = () => setCurrentTime(audio.currentTime);
		const onLoadedMetadata = () => setDuration(audio.duration);
		const onEnded = () => handleTrackEnded();
		const onPause = () => setIsPlaying(false);
		const onPlay = () => setIsPlaying(true);

		audio.addEventListener('timeupdate', onTimeUpdate);
		audio.addEventListener('loadedmetadata', onLoadedMetadata);
		audio.addEventListener('ended', onEnded);
		audio.addEventListener('pause', onPause);
		audio.addEventListener('play', onPlay);

		return () => {
			audio.removeEventListener('timeupdate', onTimeUpdate);
			audio.removeEventListener('loadedmetadata', onLoadedMetadata);
			audio.removeEventListener('ended', onEnded);
			audio.removeEventListener('pause', onPause);
			audio.removeEventListener('play', onPlay);
			audio.pause();
			audio.src = '';
		};
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	// Debounced sync to backend
	const syncState = useCallback(
		(overrides?: { fileId?: number | null; position?: number; vol?: number }) => {
			if (syncTimeoutRef.current) {
				clearTimeout(syncTimeoutRef.current);
			}
			syncTimeoutRef.current = setTimeout(() => {
				const track = currentIndex !== undefined ? queue[currentIndex] : undefined;
				updatePlayerState({
					current_file_id: overrides?.fileId !== undefined ? overrides.fileId : (track?.id ?? null),
					current_position: overrides?.position !== undefined ? overrides.position : (audioRef.current?.currentTime ?? 0),
					volume: overrides?.vol !== undefined ? overrides.vol : volume,
					shuffle,
					repeat_mode: repeatMode,
				}).catch(() => {
					// Silently ignore sync errors for personal NAS
				});
			}, 2000);
		},
		[currentIndex, queue, volume, shuffle, repeatMode],
	);

	const loadAndPlay = useCallback(
		(index: number) => {
			const audio = audioRef.current;
			if (!audio || index < 0 || index >= queue.length) return;

			const track = queue[index];
			setCurrentIndex(index);
			audio.src = `${import.meta.env.VITE_API_URL}/api/v1/files/stream/${track.id}`;
			audio.play().catch(() => {});
			syncState({ fileId: track.id, position: 0 });
		},
		[queue, syncState],
	);

	const handleTrackEnded = useCallback(() => {
		if (repeatMode === 'one') {
			const audio = audioRef.current;
			if (audio) {
				audio.currentTime = 0;
				audio.play().catch(() => {});
			}
			return;
		}

		if (currentIndex === undefined) return;

		if (shuffle) {
			const nextIndex = Math.floor(Math.random() * queue.length);
			loadAndPlay(nextIndex);
			return;
		}

		const nextIndex = currentIndex + 1;
		if (nextIndex < queue.length) {
			loadAndPlay(nextIndex);
		} else if (repeatMode === 'all') {
			loadAndPlay(0);
		} else {
			setIsPlaying(false);
		}
	}, [currentIndex, queue.length, repeatMode, shuffle, loadAndPlay]);

	// Update the ended handler when dependencies change
	useEffect(() => {
		const audio = audioRef.current;
		if (!audio) return;

		const onEnded = () => handleTrackEnded();
		audio.addEventListener('ended', onEnded);
		return () => audio.removeEventListener('ended', onEnded);
	}, [handleTrackEnded]);

	const addToQueue = useCallback(
		(track: IMusicData) => {
			setQueue((prev) => {
				if (prev.some((t) => t.id === track.id)) return prev;
				const newQueue = [...prev, track];
				// If nothing is playing, start playing the added track
				if (currentIndex === undefined) {
					setTimeout(() => loadAndPlay(newQueue.length - 1), 0);
				}
				return newQueue;
			});
			if (currentIndex === undefined) {
				setCurrentIndex(0);
			}
		},
		[currentIndex, loadAndPlay],
	);

	const playTrackFromQueue = useCallback(
		(index: number) => {
			loadAndPlay(index);
		},
		[loadAndPlay],
	);

	const clearQueue = useCallback(() => {
		const audio = audioRef.current;
		if (audio) {
			audio.pause();
			audio.src = '';
		}
		setQueue([]);
		setCurrentIndex(undefined);
		setIsPlaying(false);
		setCurrentTime(0);
		setDuration(0);
	}, []);

	const togglePlayPause = useCallback(() => {
		const audio = audioRef.current;
		if (!audio) return;

		if (isPlaying) {
			audio.pause();
		} else {
			audio.play().catch(() => {});
		}
	}, [isPlaying]);

	const next = useCallback(() => {
		if (queue.length === 0 || currentIndex === undefined) return;

		if (shuffle) {
			const nextIndex = Math.floor(Math.random() * queue.length);
			loadAndPlay(nextIndex);
			return;
		}

		const nextIndex = (currentIndex + 1) % queue.length;
		loadAndPlay(nextIndex);
	}, [currentIndex, queue.length, shuffle, loadAndPlay]);

	const previous = useCallback(() => {
		if (queue.length === 0 || currentIndex === undefined) return;

		// If more than 3 seconds in, restart current track
		if (audioRef.current && audioRef.current.currentTime > 3) {
			audioRef.current.currentTime = 0;
			return;
		}

		const prevIndex = currentIndex === 0 ? queue.length - 1 : currentIndex - 1;
		loadAndPlay(prevIndex);
	}, [currentIndex, queue.length, loadAndPlay]);

	const seek = useCallback(
		(time: number) => {
			if (audioRef.current) {
				audioRef.current.currentTime = time;
				syncState({ position: time });
			}
		},
		[syncState],
	);

	const setVolume = useCallback(
		(newVolume: number) => {
			const clamped = Math.max(0, Math.min(1, newVolume));
			setVolumeState(clamped);
			if (audioRef.current) {
				audioRef.current.volume = clamped;
			}
			syncState({ vol: clamped });
		},
		[syncState],
	);

	const toggleShuffle = useCallback(() => {
		setShuffle((prev) => !prev);
	}, []);

	const currentTrack = currentIndex !== undefined ? queue[currentIndex] : undefined;
	const hasQueue = queue.length > 0;

	const getMusicTitle = (music: IMusicData): string => {
		return music.metadata?.title || music.name;
	};

	const getMusicArtist = (music: IMusicData): string => {
		return music.metadata?.artist || 'Unknown Artist';
	};

	const formatDuration = (seconds: number): string => {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	};

	const musicMetadata = (music: { format: string; size: number; metadata?: IMusicMetadata }): string => {
		const format = music.format ? `${music.format} - ` : '';
		const fileSize = formatSize(music.size);
		const dur = music.metadata?.duration ? formatDuration(music.metadata.duration) : '';
		return `${format}${fileSize}${dur ? ` - ${dur}` : ''}`;
	};

	// Sync state when shuffle/repeat changes
	useEffect(() => {
		syncState();
	}, [shuffle, repeatMode, syncState]);

	return (
		<GlobalMusicContext.Provider
			value={{
				queue,
				currentIndex,
				addToQueue,
				playTrackFromQueue,
				clearQueue,
				isPlaying,
				currentTime,
				duration,
				volume,
				shuffle,
				repeatMode,
				togglePlayPause,
				next,
				previous,
				seek,
				setVolume,
				toggleShuffle,
				setRepeatMode,
				currentTrack,
				hasQueue,
				getMusicTitle,
				getMusicArtist,
				musicMetadata,
				formatDuration,
			}}
		>
			{children}
		</GlobalMusicContext.Provider>
	);
};

export const useGlobalMusic = () => {
	const context = useContext(GlobalMusicContext);
	if (!context) {
		throw new Error('useGlobalMusic must be used within a GlobalMusicProvider');
	}
	return context;
};

import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import type { IMusicData } from './musicProvider/musicProvider';
import { getPlayerState, updatePlayerState } from '@/service/playerState';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { createPlaylistPlaybackContext, type MusicPlaybackContext } from '@/components/music/playbackContext';
import { useSettings } from './settingsProvider/settingsContext';
import { getNowPlayingPlaylist, getPlaylistTracks } from '@/service/playlist';
import { getMusicTitle, getMusicArtist, formatMusicDuration, musicMetadata } from '@/utils/music';

type RepeatMode = 'none' | 'all' | 'one';

const SYNC_DEBOUNCE_MS = 2000;
const RESTART_THRESHOLD_SECONDS = 3;

export interface IGlobalMusicContext {
	queue: IMusicData[];
	currentIndex: number | undefined;
	addToQueue: (track: IMusicData, playbackContext?: MusicPlaybackContext) => void;
	replaceQueue: (tracks: IMusicData[], startIndex?: number, playbackContext?: MusicPlaybackContext) => void;
	playTrackFromQueue: (index: number) => void;
	clearQueue: () => void;
	removeFromQueue: (index: number) => void;
	queueOpen: boolean;
	setQueueOpen: (open: boolean) => void;
	toggleQueue: () => void;
	playbackContext?: MusicPlaybackContext;
	isPlaying: boolean;
	currentTime: number;
	duration: number;
	volume: number;
	shuffle: boolean;
	repeatMode: RepeatMode;
	togglePlayPause: () => void;
	next: () => void;
	previous: () => void;
	seek: (time: number) => void;
	setVolume: (volume: number) => void;
	toggleShuffle: () => void;
	setRepeatMode: (mode: RepeatMode) => void;
	currentTrack: IMusicData | undefined;
	hasQueue: boolean;
	getMusicTitle: (music: IMusicData) => string;
	getMusicArtist: (music: IMusicData) => string;
	musicMetadata: typeof musicMetadata;
	formatDuration: (seconds: number) => string;
}

const GlobalMusicContext = createContext<IGlobalMusicContext | undefined>(undefined);

const getShuffledIndex = (queueLength: number, currentIndex: number | undefined): number => {
	if (queueLength <= 1) return 0;
	const candidates = Array.from({ length: queueLength }, (_, i) => i).filter((i) => i !== currentIndex);
	return candidates[Math.floor(Math.random() * candidates.length)]!;
};

// eslint-disable-next-line react-refresh/only-export-components
export { getMusicTitle, getMusicArtist, formatMusicDuration, musicMetadata };

export const GlobalMusicProvider = ({ children }: { children: React.ReactNode }) => {
	const { settings, isLoading: isLoadingSettings } = useSettings();
	const [queue, setQueue] = useState<IMusicData[]>([]);
	const [currentIndex, setCurrentIndex] = useState<number | undefined>(undefined);
	const [isPlaying, setIsPlaying] = useState(false);
	const [currentTime, setCurrentTime] = useState(0);
	const [duration, setDuration] = useState(0);
	const [volume, setVolumeState] = useState(1);
	const [shuffle, setShuffle] = useState(false);
	const [repeatMode, setRepeatMode] = useState<RepeatMode>('none');
	const [queueOpen, setQueueOpen] = useState(false);
	const [playbackContext, setPlaybackContext] = useState<MusicPlaybackContext | undefined>(undefined);

	const audioRef = useRef<HTMLAudioElement | null>(null);
	const syncTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const hasHydratedQueueRef = useRef(false);
	const handleTrackEndedRef = useRef<() => void>(() => {});

	// Initialize audio element once
	useEffect(() => {
		const audio = new Audio();
		audio.volume = volume;
		audioRef.current = audio;

		const onTimeUpdate = () => setCurrentTime(audio.currentTime);
		const onLoadedMetadata = () => setDuration(audio.duration);
		const onEnded = () => handleTrackEndedRef.current();
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
		// volume is only needed at init time — audio.volume is updated directly in setVolume
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	useEffect(() => {
		if (isLoadingSettings || hasHydratedQueueRef.current) {
			return;
		}

		hasHydratedQueueRef.current = true;
		if (!settings.players.remember_music_queue) {
			return;
		}

		let cancelled = false;

		const hydrateQueue = async () => {
			try {
				const [playerState, nowPlayingPlaylist] = await Promise.all([
					getPlayerState(),
					getNowPlayingPlaylist(),
				]);

				if (!nowPlayingPlaylist.id || !playerState.current_file_id) {
					return;
				}

				const playlistTracks = await getPlaylistTracks(nowPlayingPlaylist.id, 1, Math.max(nowPlayingPlaylist.track_count, 1));
				if (cancelled) {
					return;
				}

				const hydratedQueue = playlistTracks.items.map((item) => item.file);
				const startIndex = hydratedQueue.findIndex((track) => track.id === playerState.current_file_id);
				if (hydratedQueue.length === 0 || startIndex < 0) {
					return;
				}

				setQueue(hydratedQueue);
				setCurrentIndex(startIndex);
				setPlaybackContext(createPlaylistPlaybackContext(nowPlayingPlaylist));
			} catch {
				// best effort hydration
			}
		};

		void hydrateQueue();

		return () => {
			cancelled = true;
		};
	}, [isLoadingSettings, settings.players.remember_music_queue]);

	const syncState = useCallback(
		(overrides?: { fileId?: number | null; position?: number; vol?: number; playlistId?: number | null }) => {
			if (syncTimeoutRef.current) {
				clearTimeout(syncTimeoutRef.current);
			}
			syncTimeoutRef.current = setTimeout(() => {
				const track = currentIndex !== undefined ? queue[currentIndex] : undefined;
				updatePlayerState({
					playlist_id: overrides?.playlistId !== undefined ? overrides.playlistId : (playbackContext?.playlistId ?? null),
					current_file_id: overrides?.fileId !== undefined ? overrides.fileId : (track?.id ?? null),
					current_position: overrides?.position !== undefined ? overrides.position : (audioRef.current?.currentTime ?? 0),
					volume: overrides?.vol !== undefined ? overrides.vol : volume,
					shuffle,
					repeat_mode: repeatMode,
				}).catch(() => {});
			}, SYNC_DEBOUNCE_MS);
		},
		[currentIndex, queue, volume, shuffle, repeatMode, playbackContext],
	);

	const loadAndPlay = useCallback(
		(index: number) => {
			const audio = audioRef.current;
			if (!audio || index < 0 || index >= queue.length) return;

			const track = queue[index];
			if (!track) return;
			setCurrentIndex(index);
			audio.src = `${getApiV1BaseUrl()}/files/stream/${track.id}`;
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
			loadAndPlay(getShuffledIndex(queue.length, currentIndex));
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

	// Keep the ref in sync so the audio 'ended' listener always calls the latest version
	useEffect(() => {
		handleTrackEndedRef.current = handleTrackEnded;
	}, [handleTrackEnded]);

	const addToQueue = useCallback(
		(track: IMusicData, nextPlaybackContext?: MusicPlaybackContext) => {
			if (currentIndex === undefined && nextPlaybackContext) {
				setPlaybackContext(nextPlaybackContext);
			}
			setQueue((prev) => {
				if (prev.some((t) => t.id === track.id)) return prev;
				const newQueue = [...prev, track];
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
		setPlaybackContext(undefined);
	}, []);

	const replaceQueue = useCallback(
		(tracks: IMusicData[], startIndex = 0, nextPlaybackContext?: MusicPlaybackContext) => {
			if (tracks.length === 0) return;
			setQueue(tracks);
			setCurrentIndex(startIndex);
			setPlaybackContext(nextPlaybackContext);
			const audio = audioRef.current;
			if (audio) {
				const track = tracks[startIndex];
				if (track) {
					audio.src = `${getApiV1BaseUrl()}/files/stream/${track.id}`;
					audio.play().catch(() => {});
					syncState({
						fileId: track.id,
						position: 0,
						playlistId: nextPlaybackContext?.playlistId ?? null,
					});
				}
			}
		},
		[syncState],
	);

	const removeFromQueue = useCallback(
		(index: number) => {
			setQueue((prev) => {
				const newQueue = prev.filter((_, i) => i !== index);
				if (newQueue.length === 0) {
					const audio = audioRef.current;
					if (audio) {
						audio.pause();
						audio.src = '';
					}
					setCurrentIndex(undefined);
					setIsPlaying(false);
					setCurrentTime(0);
					setDuration(0);
					setPlaybackContext(undefined);
				} else if (currentIndex !== undefined) {
					if (index < currentIndex) {
						setCurrentIndex(currentIndex - 1);
					} else if (index === currentIndex) {
						if (currentIndex >= newQueue.length) {
							setCurrentIndex(newQueue.length - 1);
						}
						const nextTrack = newQueue[Math.min(currentIndex, newQueue.length - 1)];
						if (nextTrack) {
							const audio = audioRef.current;
							if (audio) {
								audio.src = `${getApiV1BaseUrl()}/files/stream/${nextTrack.id}`;
								audio.play().catch(() => {});
							}
						}
					}
				}
				return newQueue;
			});
		},
		[currentIndex],
	);

	const toggleQueue = useCallback(() => {
		setQueueOpen((prev) => !prev);
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
			loadAndPlay(getShuffledIndex(queue.length, currentIndex));
			return;
		}

		const nextIndex = (currentIndex + 1) % queue.length;
		loadAndPlay(nextIndex);
	}, [currentIndex, queue.length, shuffle, loadAndPlay]);

	const previous = useCallback(() => {
		if (queue.length === 0 || currentIndex === undefined) return;

		if (audioRef.current && audioRef.current.currentTime > RESTART_THRESHOLD_SECONDS) {
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

	useEffect(() => {
		syncState();
	}, [shuffle, repeatMode, playbackContext, syncState]);

	return (
		<GlobalMusicContext.Provider
			value={{
				queue,
				currentIndex,
				addToQueue,
				replaceQueue,
				playTrackFromQueue,
				clearQueue,
				removeFromQueue,
				queueOpen,
				setQueueOpen,
				toggleQueue,
				playbackContext,
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
				formatDuration: formatMusicDuration,
			}}
		>
			{children}
		</GlobalMusicContext.Provider>
	);
};

// eslint-disable-next-line react-refresh/only-export-components
export const useGlobalMusic = () => {
	const context = useContext(GlobalMusicContext);
	if (!context) {
		throw new Error('useGlobalMusic must be used within a GlobalMusicProvider');
	}
	return context;
};

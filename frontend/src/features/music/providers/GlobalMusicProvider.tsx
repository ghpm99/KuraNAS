import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { IMusicData } from './musicProvider/musicProvider';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import type { MusicPlaybackContext } from '@/components/music/playbackContext';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import useAudioEngine from './globalMusic/useAudioEngine';
import useMediaSession from './globalMusic/useMediaSession';
import useMusicStateSync from './globalMusic/useMusicStateSync';
import useMusicQueueHydration from './globalMusic/useMusicQueueHydration';

type RepeatMode = 'none' | 'all' | 'one';

const RESTART_THRESHOLD_SECONDS = 3;

export interface IGlobalMusicContext {
    queue: IMusicData[];
    currentIndex: number | undefined;
    addToQueue: (track: IMusicData, playbackContext?: MusicPlaybackContext) => void;
    replaceQueue: (
        tracks: IMusicData[],
        startIndex?: number,
        playbackContext?: MusicPlaybackContext
    ) => void;
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
}

const GlobalMusicContext = createContext<IGlobalMusicContext | undefined>(undefined);

const getShuffledIndex = (queueLength: number, currentIndex: number | undefined): number => {
    if (queueLength <= 1) return 0;
    const candidates = Array.from({ length: queueLength }, (_, i) => i).filter(
        (i) => i !== currentIndex
    );
    return candidates[Math.floor(Math.random() * candidates.length)]!;
};

const buildStreamUrl = (trackId: number) => `${getApiV1BaseUrl()}/files/stream/${trackId}`;

export const GlobalMusicProvider = ({ children }: { children: React.ReactNode }) => {
    const { settings, isLoading: isLoadingSettings } = useSettings();
    const [queue, setQueue] = useState<IMusicData[]>([]);
    const [currentIndex, setCurrentIndex] = useState<number | undefined>(undefined);
    const [shuffle, setShuffle] = useState(false);
    const [repeatMode, setRepeatMode] = useState<RepeatMode>('none');
    const [queueOpen, setQueueOpen] = useState(false);
    const [playbackContext, setPlaybackContext] = useState<MusicPlaybackContext | undefined>(
        undefined
    );

    const currentTrack = currentIndex !== undefined ? queue[currentIndex] : undefined;

    // --- Audio engine ---
    const handleTrackEnded = useCallback(() => {
        if (repeatMode === 'one') {
            if (engine.audioRef.current) {
                engine.audioRef.current.currentTime = 0;
            }
            engine.audioRef.current?.play().catch(() => {});
            return;
        }
        if (currentIndex === undefined) return;
        if (shuffle) {
            const idx = getShuffledIndex(queue.length, currentIndex);
            const track = queue[idx];
            if (track) {
                setCurrentIndex(idx);
                engine.loadAndPlayUrl(buildStreamUrl(track.id));
                syncState({ fileId: track.id, position: 0 });
            }
            return;
        }
        const nextIndex = currentIndex + 1;
        if (nextIndex < queue.length) {
            const track = queue[nextIndex];
            if (track) {
                setCurrentIndex(nextIndex);
                engine.loadAndPlayUrl(buildStreamUrl(track.id));
                syncState({ fileId: track.id, position: 0 });
            }
        } else if (repeatMode === 'all') {
            const track = queue[0];
            if (track) {
                setCurrentIndex(0);
                engine.loadAndPlayUrl(buildStreamUrl(track.id));
                syncState({ fileId: track.id, position: 0 });
            }
        } else {
            engine.stop();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [currentIndex, queue, repeatMode, shuffle]);

    const engine = useAudioEngine(handleTrackEnded);

    // --- Backend sync ---
    const { syncState } = useMusicStateSync({
        getCurrentTrackId: () => currentTrack?.id,
        getCurrentTime: () => engine.audioRef.current?.currentTime ?? 0,
        volume: engine.volume,
        shuffle,
        repeatMode,
        playbackContext,
    });

    // --- Queue hydration ---
    const hydrationCallbacks = useMemo(
        () => ({ setQueue, setCurrentIndex, setPlaybackContext }),
        []
    );

    useMusicQueueHydration(
        !isLoadingSettings && settings.players.remember_music_queue,
        hydrationCallbacks
    );

    // --- Queue operations ---
    const loadAndPlay = useCallback(
        (index: number) => {
            if (index < 0 || index >= queue.length) return;
            const track = queue[index];
            if (!track) return;
            setCurrentIndex(index);
            engine.loadAndPlayUrl(buildStreamUrl(track.id));
            syncState({ fileId: track.id, position: 0 });
        },
        [queue, engine, syncState]
    );

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
        [currentIndex, loadAndPlay]
    );

    const replaceQueue = useCallback(
        (tracks: IMusicData[], startIndex = 0, nextPlaybackContext?: MusicPlaybackContext) => {
            if (tracks.length === 0) return;
            setQueue(tracks);
            setCurrentIndex(startIndex);
            setPlaybackContext(nextPlaybackContext);
            const track = tracks[startIndex];
            if (track) {
                engine.loadAndPlayUrl(buildStreamUrl(track.id));
                syncState({
                    fileId: track.id,
                    position: 0,
                    playlistId: nextPlaybackContext?.playlistId ?? null,
                });
            }
        },
        [engine, syncState]
    );

    const clearQueue = useCallback(() => {
        engine.stop();
        setQueue([]);
        setCurrentIndex(undefined);
        setPlaybackContext(undefined);
    }, [engine]);

    const removeFromQueue = useCallback(
        (index: number) => {
            setQueue((prev) => {
                const newQueue = prev.filter((_, i) => i !== index);
                if (newQueue.length === 0) {
                    engine.stop();
                    setCurrentIndex(undefined);
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
                            engine.loadAndPlayUrl(buildStreamUrl(nextTrack.id));
                        }
                    }
                }
                return newQueue;
            });
        },
        [currentIndex, engine]
    );

    const next = useCallback(() => {
        if (queue.length === 0 || currentIndex === undefined) return;
        if (shuffle) {
            loadAndPlay(getShuffledIndex(queue.length, currentIndex));
            return;
        }
        loadAndPlay((currentIndex + 1) % queue.length);
    }, [currentIndex, queue.length, shuffle, loadAndPlay]);

    const previous = useCallback(() => {
        if (queue.length === 0 || currentIndex === undefined) return;
        if (
            engine.audioRef.current &&
            engine.audioRef.current.currentTime > RESTART_THRESHOLD_SECONDS
        ) {
            engine.audioRef.current.currentTime = 0;
            return;
        }
        const prevIndex = currentIndex === 0 ? queue.length - 1 : currentIndex - 1;
        loadAndPlay(prevIndex);
    }, [currentIndex, queue.length, loadAndPlay, engine]);

    const seek = useCallback(
        (time: number) => {
            engine.seek(time);
            syncState({ position: time });
        },
        [engine, syncState]
    );

    const setVolume = useCallback(
        (newVolume: number) => {
            engine.setVolume(newVolume);
            syncState({ vol: newVolume });
        },
        [engine, syncState]
    );

    const toggleShuffle = useCallback(() => {
        setShuffle((prev) => !prev);
    }, []);

    const toggleQueue = useCallback(() => {
        setQueueOpen((prev) => !prev);
    }, []);

    // --- Preload next track for gapless background playback ---
    useEffect(() => {
        if (currentIndex === undefined || queue.length === 0 || shuffle) return;
        let nextIndex: number;
        if (repeatMode === 'one') {
            nextIndex = currentIndex;
        } else if (currentIndex + 1 < queue.length) {
            nextIndex = currentIndex + 1;
        } else if (repeatMode === 'all') {
            nextIndex = 0;
        } else {
            return;
        }
        const nextTrack = queue[nextIndex];
        if (nextTrack) {
            engine.preloadUrl(buildStreamUrl(nextTrack.id));
        }
    }, [currentIndex, queue, shuffle, repeatMode, engine]);

    // --- Media session & wake lock ---
    useMediaSession({
        currentTrack,
        isPlaying: engine.isPlaying,
        onPlay: engine.togglePlayPause,
        onPause: engine.togglePlayPause,
        onNext: next,
        onPrevious: previous,
        onSeekTo: seek,
        currentTime: engine.currentTime,
        duration: engine.duration,
    });

    // Sync settings changes to backend
    useEffect(() => {
        syncState();
    }, [shuffle, repeatMode, playbackContext, syncState]);

    const contextValue: IGlobalMusicContext = useMemo(
        () => ({
            queue,
            currentIndex,
            addToQueue,
            replaceQueue,
            playTrackFromQueue: loadAndPlay,
            clearQueue,
            removeFromQueue,
            queueOpen,
            setQueueOpen,
            toggleQueue,
            playbackContext,
            isPlaying: engine.isPlaying,
            currentTime: engine.currentTime,
            duration: engine.duration,
            volume: engine.volume,
            shuffle,
            repeatMode,
            togglePlayPause: engine.togglePlayPause,
            next,
            previous,
            seek,
            setVolume,
            toggleShuffle,
            setRepeatMode,
            currentTrack,
            hasQueue: queue.length > 0,
        }),
        [
            queue,
            currentIndex,
            addToQueue,
            replaceQueue,
            loadAndPlay,
            clearQueue,
            removeFromQueue,
            queueOpen,
            toggleQueue,
            playbackContext,
            engine.isPlaying,
            engine.currentTime,
            engine.duration,
            engine.volume,
            engine.togglePlayPause,
            shuffle,
            repeatMode,
            next,
            previous,
            seek,
            setVolume,
            toggleShuffle,
            currentTrack,
        ]
    );

    return (
        <GlobalMusicContext.Provider value={contextValue}>{children}</GlobalMusicContext.Provider>
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

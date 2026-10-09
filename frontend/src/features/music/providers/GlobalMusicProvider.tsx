import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react';
import type { IMusicData } from './musicProvider/musicProvider';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import type { MusicPlaybackContext } from '@/features/music/components/playbackContext';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import useAudioEngine from './globalMusic/useAudioEngine';
import usePlayReporting from './globalMusic/usePlayReporting';
import { getTrackDurationSeconds } from '@/utils/music';
import useMediaSession from './globalMusic/useMediaSession';
import useMusicStateSync from './globalMusic/useMusicStateSync';
import useMusicQueueHydration from './globalMusic/useMusicQueueHydration';
import useMusicQueuePersistence from './globalMusic/useMusicQueuePersistence';
import type { RepeatMode } from './globalMusic/repeatMode';
import {
    createQueueEntries,
    insertAfterIndex,
    moveQueueEntry,
    type QueueTrack,
} from './globalMusic/queueEntries';

const RESTART_THRESHOLD_SECONDS = 3;

export interface IGlobalMusicContext {
    queue: QueueTrack[];
    currentIndex: number | undefined;
    addToQueue: (tracks: IMusicData[], playbackContext?: MusicPlaybackContext) => void;
    playNext: (tracks: IMusicData[], playbackContext?: MusicPlaybackContext) => void;
    moveQueueItem: (fromIndex: number, toIndex: number) => void;
    replaceQueue: (
        tracks: IMusicData[],
        startIndex?: number,
        playbackContext?: MusicPlaybackContext
    ) => void;
    playTrackFromQueue: (index: number) => void;
    clearQueue: () => void;
    removeFromQueue: (queueEntryId: string) => void;
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
    const [queue, setQueue] = useState<QueueTrack[]>([]);
    const [currentIndex, setCurrentIndex] = useState<number | undefined>(undefined);
    const queueRef = useRef<QueueTrack[]>([]);
    const currentIndexRef = useRef<number | undefined>(undefined);

    const commitQueue = useCallback((nextQueue: QueueTrack[]) => {
        queueRef.current = nextQueue;
        setQueue(nextQueue);
    }, []);

    const commitCurrentIndex = useCallback((nextIndex: number | undefined) => {
        currentIndexRef.current = nextIndex;
        setCurrentIndex(nextIndex);
    }, []);
    const [shuffle, setShuffle] = useState(false);
    const [repeatMode, setRepeatMode] = useState<RepeatMode>('none');
    const [queueOpen, setQueueOpen] = useState(false);
    const [playbackContext, setPlaybackContext] = useState<MusicPlaybackContext | undefined>(
        undefined
    );

    const currentTrack = currentIndex !== undefined ? queue[currentIndex] : undefined;

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
                commitCurrentIndex(idx);
                engine.loadAndPlayUrl(buildStreamUrl(track.id));
                syncState({ fileId: track.id, position: 0 });
            }
            return;
        }
        const nextIndex = currentIndex + 1;
        if (nextIndex < queue.length) {
            const track = queue[nextIndex];
            if (track) {
                commitCurrentIndex(nextIndex);
                engine.loadAndPlayUrl(buildStreamUrl(track.id));
                syncState({ fileId: track.id, position: 0 });
            }
        } else if (repeatMode === 'all') {
            const track = queue[0];
            if (track) {
                commitCurrentIndex(0);
                engine.loadAndPlayUrl(buildStreamUrl(track.id));
                syncState({ fileId: track.id, position: 0 });
            }
        } else {
            engine.stop();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [currentIndex, queue, repeatMode, shuffle]);

    const engine = useAudioEngine(handleTrackEnded);

    usePlayReporting({
        queueEntryId: currentTrack?.queueEntryId,
        trackId: currentTrack?.id,
        isPlaying: engine.isPlaying,
        currentTime: engine.currentTime,
        trackDurationSeconds: getTrackDurationSeconds(currentTrack?.metadata) || engine.duration,
    });

    const { syncState } = useMusicStateSync({
        getCurrentTrackId: () => currentTrack?.id,
        getCurrentTime: () => engine.audioRef.current?.currentTime ?? 0,
        volume: engine.volume,
        shuffle,
        repeatMode,
        playbackContext,
    });

    const { loadUrlPaused, setVolume: setEngineVolume } = engine;
    const hydrationCallbacks = useMemo(
        () => ({
            setQueue: (tracks: IMusicData[]) => commitQueue(createQueueEntries(tracks)),
            setCurrentIndex: commitCurrentIndex,
            setShuffle,
            setRepeatMode,
            setVolume: setEngineVolume,
            loadPausedTrack: (trackId: number, startPositionSeconds: number) =>
                loadUrlPaused(buildStreamUrl(trackId), startPositionSeconds),
        }),
        [loadUrlPaused, setEngineVolume, commitQueue, commitCurrentIndex]
    );

    const isRememberQueueEnabled = !isLoadingSettings && settings.players.remember_music_queue;
    const { hasSettled: hasHydrationSettled } = useMusicQueueHydration(
        isRememberQueueEnabled,
        hydrationCallbacks
    );

    useMusicQueuePersistence({
        isEnabled: isRememberQueueEnabled && hasHydrationSettled,
        queue,
        currentIndex,
    });

    const startTrackAt = useCallback(
        (index: number, trackQueue: QueueTrack[]) => {
            const track = trackQueue[index];
            if (!track) return;
            commitCurrentIndex(index);
            engine.loadAndPlayUrl(buildStreamUrl(track.id));
            syncState({ fileId: track.id, position: 0 });
        },
        [commitCurrentIndex, engine, syncState]
    );

    const loadAndPlay = useCallback(
        (index: number) => startTrackAt(index, queueRef.current),
        [startTrackAt]
    );

    const startQueueWith = useCallback(
        (tracks: IMusicData[], nextPlaybackContext?: MusicPlaybackContext) => {
            const entries = createQueueEntries(tracks);
            commitQueue(entries);
            setPlaybackContext(nextPlaybackContext);
            startTrackAt(0, entries);
        },
        [commitQueue, startTrackAt]
    );

    const addToQueue = useCallback(
        (tracks: IMusicData[], nextPlaybackContext?: MusicPlaybackContext) => {
            if (tracks.length === 0) return;
            if (currentIndexRef.current === undefined) {
                startQueueWith(tracks, nextPlaybackContext);
                return;
            }
            commitQueue([...queueRef.current, ...createQueueEntries(tracks)]);
        },
        [commitQueue, startQueueWith]
    );

    const playNext = useCallback(
        (tracks: IMusicData[], nextPlaybackContext?: MusicPlaybackContext) => {
            if (tracks.length === 0) return;
            if (currentIndexRef.current === undefined) {
                startQueueWith(tracks, nextPlaybackContext);
                return;
            }
            commitQueue(
                insertAfterIndex(
                    queueRef.current,
                    currentIndexRef.current,
                    createQueueEntries(tracks)
                )
            );
        },
        [commitQueue, startQueueWith]
    );

    const moveQueueItem = useCallback(
        (fromIndex: number, toIndex: number) => {
            const reorderedQueue = moveQueueEntry(queueRef.current, fromIndex, toIndex);
            if (reorderedQueue === queueRef.current) return;
            const currentEntryId =
                currentIndexRef.current !== undefined
                    ? queueRef.current[currentIndexRef.current]?.queueEntryId
                    : undefined;
            commitQueue(reorderedQueue);
            if (currentEntryId !== undefined) {
                commitCurrentIndex(
                    reorderedQueue.findIndex((entry) => entry.queueEntryId === currentEntryId)
                );
            }
        },
        [commitQueue, commitCurrentIndex]
    );

    const replaceQueue = useCallback(
        (tracks: IMusicData[], startIndex = 0, nextPlaybackContext?: MusicPlaybackContext) => {
            if (tracks.length === 0) return;
            const entries = createQueueEntries(tracks);
            commitQueue(entries);
            setPlaybackContext(nextPlaybackContext);
            commitCurrentIndex(startIndex);
            const track = entries[startIndex];
            if (!track) return;
            engine.loadAndPlayUrl(buildStreamUrl(track.id));
            syncState({
                fileId: track.id,
                position: 0,
                playlistId: nextPlaybackContext?.playlistId ?? null,
            });
        },
        [commitQueue, commitCurrentIndex, engine, syncState]
    );

    const clearQueue = useCallback(() => {
        engine.stop();
        commitQueue([]);
        commitCurrentIndex(undefined);
        setPlaybackContext(undefined);
    }, [engine, commitQueue, commitCurrentIndex]);

    const removeFromQueue = useCallback(
        (queueEntryId: string) => {
            const removedIndex = queueRef.current.findIndex(
                (entry) => entry.queueEntryId === queueEntryId
            );
            if (removedIndex === -1) return;
            const remainingQueue = queueRef.current.filter(
                (entry) => entry.queueEntryId !== queueEntryId
            );
            const playingIndex = currentIndexRef.current;
            commitQueue(remainingQueue);

            if (remainingQueue.length === 0) {
                engine.stop();
                commitCurrentIndex(undefined);
                setPlaybackContext(undefined);
                return;
            }
            if (playingIndex === undefined || removedIndex > playingIndex) return;
            if (removedIndex < playingIndex) {
                commitCurrentIndex(playingIndex - 1);
                return;
            }
            const replacementIndex = Math.min(playingIndex, remainingQueue.length - 1);
            commitCurrentIndex(replacementIndex);
            engine.loadAndPlayUrl(buildStreamUrl(remainingQueue[replacementIndex]!.id));
        },
        [engine, commitQueue, commitCurrentIndex]
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

    useEffect(() => {
        syncState();
    }, [shuffle, repeatMode, playbackContext, syncState]);

    const contextValue: IGlobalMusicContext = useMemo(
        () => ({
            queue,
            currentIndex,
            addToQueue,
            playNext,
            moveQueueItem,
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
            playNext,
            moveQueueItem,
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

// eslint-disable-next-line react-refresh/only-export-components
export const useOptionalGlobalMusic = () => useContext(GlobalMusicContext);

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
import type { MusicPlaybackContext } from '@/features/music/components/playbackContext';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import useAudioEngine from './globalMusic/useAudioEngine';
import usePlayReporting from './globalMusic/usePlayReporting';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicTitle, getTrackDurationSeconds } from '@/utils/music';
import {
    buildDirectStreamUrl,
    resolveTrackStreamSource,
    type TrackStreamSource,
} from './globalMusic/trackStreamSource';
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
import {
    appendEntries,
    findEntryAfter,
    findEntryBefore,
    insertAfterEntry,
    reconcileShuffleOrder,
    removeEntry,
    reshuffleOrder,
    type ShuffleOrder,
} from './globalMusic/shufflePlan';

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

export const GlobalMusicProvider = ({ children }: { children: React.ReactNode }) => {
    const { settings, isLoading: isLoadingSettings } = useSettings();
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const [queue, setQueue] = useState<QueueTrack[]>([]);
    const [currentIndex, setCurrentIndex] = useState<number | undefined>(undefined);
    const queueRef = useRef<QueueTrack[]>([]);
    const currentIndexRef = useRef<number | undefined>(undefined);
    const shuffleOrderRef = useRef<ShuffleOrder>([]);

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

    const resolveShuffleOrder = useCallback(
        (trackQueue: QueueTrack[], currentEntryId: string | undefined): ShuffleOrder => {
            const order = reconcileShuffleOrder(
                shuffleOrderRef.current,
                trackQueue,
                currentEntryId
            );
            shuffleOrderRef.current = order;
            return order;
        },
        []
    );

    const findShuffleSuccessorIndex = useCallback(
        (trackQueue: QueueTrack[], playingIndex: number): number | undefined => {
            const currentEntryId = trackQueue[playingIndex]?.queueEntryId;
            if (currentEntryId === undefined) return undefined;
            const order = resolveShuffleOrder(trackQueue, currentEntryId);
            const successorEntryId = findEntryAfter(order, currentEntryId);
            if (successorEntryId === undefined) return undefined;
            return trackQueue.findIndex((entry) => entry.queueEntryId === successorEntryId);
        },
        [resolveShuffleOrder]
    );

    const findReshuffledStartIndex = useCallback(
        (trackQueue: QueueTrack[], playingIndex: number): number => {
            const order = reshuffleOrder(trackQueue, trackQueue[playingIndex]?.queueEntryId);
            shuffleOrderRef.current = order;
            return trackQueue.findIndex((entry) => entry.queueEntryId === order[0]);
        },
        []
    );

    const handleTrackEnded = useCallback(() => {
        if (repeatMode === 'one') {
            engine.seek(0);
            engine.audioRef.current?.play().catch(() => {});
            return;
        }
        if (currentIndex === undefined) return;
        if (shuffle) {
            const successorIndex = findShuffleSuccessorIndex(queue, currentIndex);
            const isEndOfOrder = successorIndex === undefined;
            if (isEndOfOrder && repeatMode !== 'all') {
                engine.stop();
                return;
            }
            const nextShuffledIndex = isEndOfOrder
                ? findReshuffledStartIndex(queue, currentIndex)
                : successorIndex;
            const track = queue[nextShuffledIndex];
            if (track) {
                commitCurrentIndex(nextShuffledIndex);
                playTrack(track);
                syncState({ fileId: track.id, position: 0 });
            }
            return;
        }
        const nextIndex = currentIndex + 1;
        if (nextIndex < queue.length) {
            const track = queue[nextIndex];
            if (track) {
                commitCurrentIndex(nextIndex);
                playTrack(track);
                syncState({ fileId: track.id, position: 0 });
            }
        } else if (repeatMode === 'all') {
            const track = queue[0];
            if (track) {
                commitCurrentIndex(0);
                playTrack(track);
                syncState({ fileId: track.id, position: 0 });
            }
        } else {
            engine.stop();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [currentIndex, queue, repeatMode, shuffle]);

    const notifyPlaybackFailure = useCallback(() => {
        const failedTrack = currentIndex !== undefined ? queue[currentIndex] : undefined;
        const title = failedTrack ? getMusicTitle(failedTrack) : '';
        enqueueSnackbar(t('MUSIC_PLAYBACK_FAILED', { title }), { variant: 'error' });
    }, [currentIndex, queue, enqueueSnackbar, t]);

    const engine = useAudioEngine(handleTrackEnded, notifyPlaybackFailure);

    const { canPlayType, loadAndPlayUrl } = engine;

    const resolveStreamSource = useCallback(
        (track: IMusicData): TrackStreamSource => resolveTrackStreamSource(track, canPlayType),
        [canPlayType]
    );

    const playTrack = useCallback(
        (track: IMusicData) => {
            const streamSource = resolveStreamSource(track);
            loadAndPlayUrl(streamSource.url, streamSource.transcodedStream);
        },
        [resolveStreamSource, loadAndPlayUrl]
    );

    usePlayReporting({
        queueEntryId: currentTrack?.queueEntryId,
        trackId: currentTrack?.id,
        isPlaying: engine.isPlaying,
        currentTime: engine.currentTime,
        trackDurationSeconds: getTrackDurationSeconds(currentTrack?.metadata) || engine.duration,
    });

    const { syncState } = useMusicStateSync({
        getCurrentTrackId: () => currentTrack?.id,
        getCurrentTime: () => engine.getPositionSeconds(),
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
            loadPausedTrack: (trackId: number, startPositionSeconds: number) => {
                const pausedTrack = queueRef.current.find((entry) => entry.id === trackId);
                const streamSource = pausedTrack
                    ? resolveTrackStreamSource(pausedTrack, canPlayType)
                    : { url: buildDirectStreamUrl(trackId) };
                loadUrlPaused(
                    streamSource.url,
                    startPositionSeconds,
                    streamSource.transcodedStream
                );
            },
        }),
        [loadUrlPaused, canPlayType, setEngineVolume, commitQueue, commitCurrentIndex]
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
            playTrack(track);
            syncState({ fileId: track.id, position: 0 });
        },
        [commitCurrentIndex, playTrack, syncState]
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
            const appendedEntries = createQueueEntries(tracks);
            if (shuffle) {
                const order = resolveShuffleOrder(
                    queueRef.current,
                    queueRef.current[currentIndexRef.current]?.queueEntryId
                );
                shuffleOrderRef.current = appendEntries(
                    order,
                    appendedEntries.map((entry) => entry.queueEntryId)
                );
            }
            commitQueue([...queueRef.current, ...appendedEntries]);
        },
        [commitQueue, startQueueWith, shuffle, resolveShuffleOrder]
    );

    const playNext = useCallback(
        (tracks: IMusicData[], nextPlaybackContext?: MusicPlaybackContext) => {
            if (tracks.length === 0) return;
            if (currentIndexRef.current === undefined) {
                startQueueWith(tracks, nextPlaybackContext);
                return;
            }
            const insertedEntries = createQueueEntries(tracks);
            if (shuffle) {
                const currentEntryId = queueRef.current[currentIndexRef.current]?.queueEntryId;
                const order = resolveShuffleOrder(queueRef.current, currentEntryId);
                shuffleOrderRef.current = insertAfterEntry(
                    order,
                    currentEntryId,
                    insertedEntries.map((entry) => entry.queueEntryId)
                );
            }
            commitQueue(
                insertAfterIndex(queueRef.current, currentIndexRef.current, insertedEntries)
            );
        },
        [commitQueue, startQueueWith, shuffle, resolveShuffleOrder]
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
            playTrack(track);
            syncState({
                fileId: track.id,
                position: 0,
                playlistId: nextPlaybackContext?.playlistId ?? null,
            });
        },
        [commitQueue, commitCurrentIndex, playTrack, syncState]
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
            const isRemovingCurrentEntry = removedIndex === playingIndex;
            shuffleOrderRef.current = isRemovingCurrentEntry
                ? []
                : removeEntry(shuffleOrderRef.current, queueEntryId);
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
            playTrack(remainingQueue[replacementIndex]!);
        },
        [engine, playTrack, commitQueue, commitCurrentIndex]
    );

    const next = useCallback(() => {
        if (queue.length === 0 || currentIndex === undefined) return;
        if (shuffle) {
            const successorIndex = findShuffleSuccessorIndex(queue, currentIndex);
            loadAndPlay(successorIndex ?? findReshuffledStartIndex(queue, currentIndex));
            return;
        }
        loadAndPlay((currentIndex + 1) % queue.length);
    }, [
        currentIndex,
        queue,
        shuffle,
        loadAndPlay,
        findShuffleSuccessorIndex,
        findReshuffledStartIndex,
    ]);

    const previous = useCallback(() => {
        if (queue.length === 0 || currentIndex === undefined) return;
        if (engine.getPositionSeconds() > RESTART_THRESHOLD_SECONDS) {
            engine.seek(0);
            return;
        }
        if (shuffle) {
            const currentEntryId = queue[currentIndex]?.queueEntryId;
            if (currentEntryId === undefined) return;
            const order = resolveShuffleOrder(queue, currentEntryId);
            const previousEntryId = findEntryBefore(order, currentEntryId);
            loadAndPlay(queue.findIndex((entry) => entry.queueEntryId === previousEntryId));
            return;
        }
        const prevIndex = currentIndex === 0 ? queue.length - 1 : currentIndex - 1;
        loadAndPlay(prevIndex);
    }, [currentIndex, queue, shuffle, loadAndPlay, engine, resolveShuffleOrder]);

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
        shuffleOrderRef.current = [];
        setShuffle((isShuffleOn) => !isShuffleOn);
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
        if (!nextTrack) return;
        const nextStreamSource = resolveStreamSource(nextTrack);
        if (nextStreamSource.transcodedStream) return;
        engine.preloadUrl(nextStreamSource.url);
    }, [currentIndex, queue, shuffle, repeatMode, engine, resolveStreamSource]);

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

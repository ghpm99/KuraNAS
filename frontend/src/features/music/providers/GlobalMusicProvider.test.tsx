import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createPlaylistPlaybackContext } from '@/features/music/components/playbackContext';
import type { IMusicData } from '@/types/music';
import { GlobalMusicProvider, useGlobalMusic } from './GlobalMusicProvider';

const fakeSettings = {
    players: {
        remember_music_queue: true,
    },
};

const createEngineMock = () => ({
    audioRef: {
        current: {
            currentTime: 0,
            play: jest.fn().mockResolvedValue(undefined),
        } as { currentTime: number; play: jest.Mock } | null,
    },
    loadAndPlayUrl: jest.fn(),
    loadUrlPaused: jest.fn(),
    preloadUrl: jest.fn(),
    canPlayType: jest.fn().mockReturnValue('maybe'),
    getPositionSeconds: jest.fn().mockReturnValue(0),
    togglePlayPause: jest.fn(),
    seek: jest.fn(),
    setVolume: jest.fn(),
    stop: jest.fn(),
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 1,
});

const linkEnginePositionToAudio = (engine: ReturnType<typeof createEngineMock>) => {
    engine.getPositionSeconds.mockImplementation(() => engine.audioRef.current?.currentTime ?? 0);
    engine.seek.mockImplementation((seconds: number) => {
        if (engine.audioRef.current) engine.audioRef.current.currentTime = seconds;
    });
    return engine;
};

let engineMock = linkEnginePositionToAudio(createEngineMock());
const mockSyncState = jest.fn();
const mockQueueHydration = jest.fn();
const mockQueuePersistence = jest.fn();
let mockHasHydrationSettled = false;
let capturedOnTrackEnded: (() => void) | undefined;
let capturedOnPlaybackFailure: (() => void) | undefined;
const mockEnqueueSnackbar = jest.fn();

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('./globalMusic/useAudioEngine', () => ({
    __esModule: true,
    default: (onTrackEnded: () => void, onPlaybackFailure: () => void) => {
        capturedOnTrackEnded = onTrackEnded;
        capturedOnPlaybackFailure = onPlaybackFailure;
        return engineMock;
    },
}));

type SyncDeps = {
    getCurrentTrackId: () => number | undefined;
    getCurrentTime: () => number;
};
let capturedSyncDeps: SyncDeps | undefined;

jest.mock('./globalMusic/useMusicStateSync', () => ({
    __esModule: true,
    default: (deps: SyncDeps) => {
        capturedSyncDeps = deps;
        return { syncState: mockSyncState };
    },
}));

jest.mock('./globalMusic/useMusicQueueHydration', () => ({
    __esModule: true,
    default: (enabled: boolean, callbacks: unknown) => {
        mockQueueHydration(enabled, callbacks);
        return { hasSettled: mockHasHydrationSettled };
    },
}));

jest.mock('./globalMusic/useMusicQueuePersistence', () => ({
    __esModule: true,
    default: (params: unknown) => mockQueuePersistence(params),
}));

jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
    __esModule: true,
    useSettings: () => ({
        settings: fakeSettings,
        isLoading: false,
        isSaving: false,
        hasError: false,
        refresh: jest.fn(),
        saveSettings: jest.fn(),
    }),
}));

const wrapper = ({ children }: { children?: ReactNode }) => (
    <GlobalMusicProvider>{children}</GlobalMusicProvider>
);

const createTrack = (id: number): IMusicData => ({
    id,
    name: `track-${id}`,
    path: `/tracks/${id}.mp3`,
    type: id,
    format: 'mp3',
    size: 1024,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
});

describe('GlobalMusicProvider', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        engineMock = linkEnginePositionToAudio(createEngineMock());
        mockSyncState.mockReset();
        mockEnqueueSnackbar.mockReset();
        mockQueueHydration.mockReset();
        mockQueuePersistence.mockReset();
        mockHasHydrationSettled = false;
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('restores the saved player through the hydration callbacks without autoplay', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const [isHydrationEnabled, hydrationCallbacks] = mockQueueHydration.mock.calls[0];

        act(() => {
            hydrationCallbacks.setQueue([createTrack(4), createTrack(5)]);
            hydrationCallbacks.setCurrentIndex(1);
            hydrationCallbacks.setShuffle(true);
            hydrationCallbacks.setRepeatMode('all');
            hydrationCallbacks.setVolume(0.3);
            hydrationCallbacks.loadPausedTrack(5, 12);
        });

        expect(isHydrationEnabled).toBe(true);
        expect(result.current.queue).toHaveLength(2);
        expect(result.current.currentIndex).toBe(1);
        expect(result.current.shuffle).toBe(true);
        expect(result.current.repeatMode).toBe('all');
        expect(engineMock.setVolume).toHaveBeenCalledWith(0.3);
        expect(engineMock.loadUrlPaused).toHaveBeenCalledWith(
            expect.stringContaining('/files/stream/5'),
            12,
            undefined
        );
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
    });

    it('persists the queue only after hydration settled', () => {
        const { rerender } = renderHook(() => useGlobalMusic(), { wrapper });
        expect(mockQueuePersistence).toHaveBeenLastCalledWith(
            expect.objectContaining({ isEnabled: false })
        );

        mockHasHydrationSettled = true;
        rerender();

        expect(mockQueuePersistence).toHaveBeenLastCalledWith(
            expect.objectContaining({ isEnabled: true, queue: [], currentIndex: undefined })
        );
    });

    it('manages queue operations and shuffle/previous flows', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const trackA = createTrack(1);
        const trackB = createTrack(2);
        const trackC = createTrack(3);

        act(() => {
            result.current.addToQueue(
                [trackA],
                createPlaylistPlaybackContext({ id: 10, name: 'Queue 10' })
            );
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        expect(result.current.queue).toEqual([expect.objectContaining({ id: trackA.id })]);
        expect(result.current.currentIndex).toBe(0);
        expect(result.current.playbackContext).toEqual(
            createPlaylistPlaybackContext({ id: 10, name: 'Queue 10' })
        );

        act(() => {
            result.current.addToQueue([trackA]);
            result.current.addToQueue([trackB]);
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        expect(result.current.queue).toHaveLength(3);

        act(() => {
            result.current.replaceQueue(
                [trackB, trackC],
                1,
                createPlaylistPlaybackContext({ id: 22, name: 'Queue 22' })
            );
        });
        expect(result.current.queue[1]).toEqual(expect.objectContaining({ id: trackC.id }));
        expect(result.current.currentIndex).toBe(1);
        expect(engineMock.loadAndPlayUrl).toHaveBeenLastCalledWith(
            expect.stringContaining('/files/stream/3'),
            undefined
        );
        expect(mockSyncState).toHaveBeenCalledWith(
            expect.objectContaining({ fileId: 3, position: 0, playlistId: 22 })
        );

        act(() => {
            result.current.playTrackFromQueue(0);
        });
        expect(result.current.currentIndex).toBe(0);

        act(() => {
            result.current.toggleShuffle();
        });
        const mathSpy = jest.spyOn(Math, 'random').mockReturnValue(0);
        act(() => {
            result.current.next();
        });
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalled();
        mathSpy.mockRestore();

        engineMock.audioRef.current!.currentTime = 5;
        const previousCalls = engineMock.loadAndPlayUrl.mock.calls.length;
        act(() => {
            result.current.previous();
        });
        expect(engineMock.audioRef.current!.currentTime).toBe(0);
        expect(engineMock.loadAndPlayUrl.mock.calls.length).toBe(previousCalls);

        engineMock.audioRef.current!.currentTime = 1;
        act(() => {
            result.current.previous();
        });
        expect(engineMock.loadAndPlayUrl.mock.calls.length).toBe(previousCalls + 1);
    });

    it('exposes helpers and resets queue correctly', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const track = createTrack(5);

        act(() => {
            result.current.replaceQueue([], 0);
        });
        expect(result.current.queue).toEqual([]);

        act(() => {
            result.current.addToQueue([track]);
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        expect(result.current.hasQueue).toBe(true);

        act(() => {
            result.current.togglePlayPause();
            result.current.seek(12);
            result.current.setVolume(0.25);
            result.current.toggleQueue();
        });
        expect(engineMock.togglePlayPause).toHaveBeenCalled();
        expect(engineMock.seek).toHaveBeenCalledWith(12);
        expect(engineMock.setVolume).toHaveBeenCalledWith(0.25);
        expect(result.current.queueOpen).toBe(true);
        expect(mockSyncState).toHaveBeenCalledWith(expect.objectContaining({ position: 12 }));

        act(() => {
            result.current.toggleQueue();
            result.current.clearQueue();
        });
        expect(result.current.queue).toEqual([]);
        expect(result.current.currentIndex).toBeUndefined();
        expect(result.current.hasQueue).toBe(false);
        expect(engineMock.stop).toHaveBeenCalled();
    });

    it('starts a track whose format the browser cannot play through the transcode URL', () => {
        engineMock.canPlayType.mockReturnValue('');
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const wmaTrack = { ...createTrack(8), format: '.wma', metadata: { length: 180 } as never };

        act(() => {
            result.current.replaceQueue([wmaTrack], 0);
        });

        expect(engineMock.canPlayType).toHaveBeenCalledWith('audio/x-ms-wma');
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
            expect.stringContaining('/music/tracks/8/stream?format=mp3'),
            expect.objectContaining({ durationSeconds: 180 })
        );
    });

    it('shows an error snackbar when playback keeps failing', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue(
                [{ ...createTrack(5), metadata: { title: 'Broken' } as never }],
                0
            );
        });
        act(() => {
            capturedOnPlaybackFailure?.();
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYBACK_FAILED', {
            variant: 'error',
        });
    });

    it('does not preload the next track when it needs transcoding', () => {
        engineMock.canPlayType.mockImplementation((mimeType: string) =>
            mimeType === 'audio/x-ms-wma' ? '' : 'maybe'
        );
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), { ...createTrack(2), format: 'wma' }, createTrack(3)];

        act(() => {
            result.current.replaceQueue(tracks, 0);
        });

        expect(engineMock.preloadUrl).not.toHaveBeenCalled();
    });

    it('removeFromQueue adjusts currentIndex when removing before current', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), createTrack(2), createTrack(3)];

        act(() => {
            result.current.replaceQueue(tracks, 2);
        });
        expect(result.current.currentIndex).toBe(2);

        // Remove track before current index: index should shift down by 1
        act(() => {
            result.current.removeFromQueue(result.current.queue[0]!.queueEntryId);
        });
        expect(result.current.currentIndex).toBe(1);
        expect(result.current.queue).toHaveLength(2);
    });

    it('removeFromQueue plays next track when removing the current track', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), createTrack(2), createTrack(3)];

        act(() => {
            result.current.replaceQueue(tracks, 1);
        });
        engineMock.loadAndPlayUrl.mockClear();

        // Remove the currently playing track (index 1)
        act(() => {
            result.current.removeFromQueue(result.current.queue[1]!.queueEntryId);
        });
        // Should load the next track (track 3 is now at index 1)
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
            expect.stringContaining('/files/stream/3'),
            undefined
        );
        expect(result.current.queue).toHaveLength(2);
    });

    it('removeFromQueue wraps currentIndex when removing current track at end', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), createTrack(2)];

        act(() => {
            result.current.replaceQueue(tracks, 1);
        });
        engineMock.loadAndPlayUrl.mockClear();

        // Remove the last track which is also the current track
        act(() => {
            result.current.removeFromQueue(result.current.queue[1]!.queueEntryId);
        });
        // currentIndex should clamp to newQueue.length - 1 = 0
        expect(result.current.currentIndex).toBe(0);
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
            expect.stringContaining('/files/stream/1'),
            undefined
        );
    });

    it('removeFromQueue does not change currentIndex when removing after current', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), createTrack(2), createTrack(3)];

        act(() => {
            result.current.replaceQueue(tracks, 0);
        });

        act(() => {
            result.current.removeFromQueue(result.current.queue[2]!.queueEntryId);
        });
        expect(result.current.currentIndex).toBe(0);
        expect(result.current.queue).toHaveLength(2);
    });

    it('removeFromQueue clears everything when removing the last remaining track', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1)], 0);
        });

        act(() => {
            result.current.removeFromQueue(result.current.queue[0]!.queueEntryId);
        });
        expect(result.current.queue).toHaveLength(0);
        expect(result.current.currentIndex).toBeUndefined();
        expect(engineMock.stop).toHaveBeenCalled();
    });

    it('next wraps around to index 0 when at end of queue without shuffle', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), createTrack(2)];

        act(() => {
            result.current.replaceQueue(tracks, 1);
        });
        engineMock.loadAndPlayUrl.mockClear();

        act(() => {
            result.current.next();
        });
        // (1 + 1) % 2 = 0
        expect(result.current.currentIndex).toBe(0);
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
            expect.stringContaining('/files/stream/1'),
            undefined
        );
    });

    it('next does nothing when queue is empty', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.next();
        });
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
    });

    it('previous does nothing when queue is empty', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.previous();
        });
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
    });

    it('previous wraps to last track when at index 0 and currentTime <= 3s', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), createTrack(2), createTrack(3)];

        act(() => {
            result.current.replaceQueue(tracks, 0);
        });
        engineMock.loadAndPlayUrl.mockClear();
        engineMock.audioRef.current!.currentTime = 1;

        act(() => {
            result.current.previous();
        });
        // Should wrap to last track
        expect(result.current.currentIndex).toBe(2);
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
            expect.stringContaining('/files/stream/3'),
            undefined
        );
    });

    it('setRepeatMode changes the repeat mode', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        expect(result.current.repeatMode).toBe('none');
        act(() => {
            result.current.setRepeatMode('all');
        });
        expect(result.current.repeatMode).toBe('all');
        act(() => {
            result.current.setRepeatMode('one');
        });
        expect(result.current.repeatMode).toBe('one');
    });

    it('toggleShuffle toggles shuffle state', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        expect(result.current.shuffle).toBe(false);
        act(() => {
            result.current.toggleShuffle();
        });
        expect(result.current.shuffle).toBe(true);
        act(() => {
            result.current.toggleShuffle();
        });
        expect(result.current.shuffle).toBe(false);
    });

    it('setQueueOpen controls queue visibility', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        expect(result.current.queueOpen).toBe(false);
        act(() => {
            result.current.setQueueOpen(true);
        });
        expect(result.current.queueOpen).toBe(true);
        act(() => {
            result.current.setQueueOpen(false);
        });
        expect(result.current.queueOpen).toBe(false);
    });

    it('addToQueue allows the same track to be queued twice with distinct entry ids', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const track = createTrack(1);

        act(() => {
            result.current.addToQueue([track]);
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        act(() => {
            result.current.addToQueue([track]);
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        expect(result.current.queue).toHaveLength(2);
        expect(result.current.queue[0]!.queueEntryId).not.toBe(
            result.current.queue[1]!.queueEntryId
        );
    });

    it('addToQueue does not overwrite playbackContext when queue is already playing', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.addToQueue(
                [createTrack(1)],
                createPlaylistPlaybackContext({ id: 10, name: 'Queue 10' })
            );
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        expect(result.current.playbackContext).toEqual(
            createPlaylistPlaybackContext({ id: 10, name: 'Queue 10' })
        );

        // Adding another track with a different context should NOT change playbackContext
        act(() => {
            result.current.addToQueue(
                [createTrack(2)],
                createPlaylistPlaybackContext({ id: 20, name: 'Queue 20' })
            );
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        expect(result.current.playbackContext).toEqual(
            createPlaylistPlaybackContext({ id: 10, name: 'Queue 10' })
        );
    });

    it('playNext inserts right after the current track and keeps the current index', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1), createTrack(2), createTrack(3)], 1);
        });
        engineMock.loadAndPlayUrl.mockClear();
        act(() => {
            result.current.playNext([createTrack(8), createTrack(9)]);
        });

        expect(result.current.queue.map((entry) => entry.id)).toEqual([1, 2, 8, 9, 3]);
        expect(result.current.currentIndex).toBe(1);
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
    });

    it('playNext starts playback when the queue is empty', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.playNext([createTrack(4)]);
        });

        expect(result.current.currentIndex).toBe(0);
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledTimes(1);
    });

    it('addToQueue appends to the end without touching playback', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1), createTrack(2)], 0);
        });
        engineMock.loadAndPlayUrl.mockClear();
        act(() => {
            result.current.addToQueue([createTrack(1), createTrack(7)]);
        });

        expect(result.current.queue.map((entry) => entry.id)).toEqual([1, 2, 1, 7]);
        expect(result.current.currentIndex).toBe(0);
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
    });

    it('moveQueueItem reorders and keeps following the current track', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1), createTrack(2), createTrack(3)], 0);
        });
        act(() => {
            result.current.moveQueueItem(2, 0);
        });
        expect(result.current.queue.map((entry) => entry.id)).toEqual([3, 1, 2]);
        expect(result.current.currentIndex).toBe(1);
        expect(result.current.currentTrack?.id).toBe(1);

        act(() => {
            result.current.moveQueueItem(1, 2);
        });
        expect(result.current.queue.map((entry) => entry.id)).toEqual([3, 2, 1]);
        expect(result.current.currentIndex).toBe(2);

        act(() => {
            result.current.moveQueueItem(0, 9);
        });
        expect(result.current.queue.map((entry) => entry.id)).toEqual([3, 2, 1]);
    });

    it('removeFromQueue removes only the chosen duplicate entry and ignores unknown ids', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1), createTrack(2), createTrack(1)], 0);
        });
        const lastEntryId = result.current.queue[2]!.queueEntryId;
        act(() => {
            result.current.removeFromQueue('missing');
            result.current.removeFromQueue(lastEntryId);
        });

        expect(result.current.queue.map((entry) => entry.id)).toEqual([1, 2]);
    });

    it('runs playback side effects once per operation', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1), createTrack(2)], 0);
        });
        engineMock.loadAndPlayUrl.mockClear();
        act(() => {
            result.current.removeFromQueue(result.current.queue[0]!.queueEntryId);
        });
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledTimes(1);

        engineMock.loadAndPlayUrl.mockClear();
        act(() => {
            result.current.clearQueue();
            result.current.addToQueue([createTrack(5)]);
        });
        act(() => {
            jest.runOnlyPendingTimers();
        });
        expect(engineMock.loadAndPlayUrl).toHaveBeenCalledTimes(1);
    });

    it('replaceQueue with empty array is a no-op', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1)], 0);
        });
        engineMock.loadAndPlayUrl.mockClear();

        act(() => {
            result.current.replaceQueue([]);
        });
        // Queue should remain unchanged
        expect(result.current.queue).toHaveLength(1);
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
    });

    it('seek syncs position to backend', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.seek(30);
        });
        expect(engineMock.seek).toHaveBeenCalledWith(30);
        expect(mockSyncState).toHaveBeenCalledWith({ position: 30 });
    });

    it('setVolume syncs volume to backend', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.setVolume(0.5);
        });
        expect(engineMock.setVolume).toHaveBeenCalledWith(0.5);
        expect(mockSyncState).toHaveBeenCalledWith({ vol: 0.5 });
    });

    it('playTrackFromQueue ignores out-of-bounds index', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        act(() => {
            result.current.replaceQueue([createTrack(1)], 0);
        });
        engineMock.loadAndPlayUrl.mockClear();

        act(() => {
            result.current.playTrackFromQueue(5);
        });
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();

        act(() => {
            result.current.playTrackFromQueue(-1);
        });
        expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
    });

    it('currentTrack returns the track at currentIndex', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });
        const tracks = [createTrack(1), createTrack(2)];

        expect(result.current.currentTrack).toBeUndefined();

        act(() => {
            result.current.replaceQueue(tracks, 1);
        });
        expect(result.current.currentTrack).toEqual(expect.objectContaining({ id: 2 }));
    });

    it('throws when useGlobalMusic is used outside provider', () => {
        // Suppress console.error for this test
        const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
        expect(() => {
            renderHook(() => useGlobalMusic());
        }).toThrow('useGlobalMusic must be used within a GlobalMusicProvider');
        consoleSpy.mockRestore();
    });

    it('sync deps callbacks return correct values', () => {
        const { result } = renderHook(() => useGlobalMusic(), { wrapper });

        // With no current track, getCurrentTrackId returns undefined
        expect(capturedSyncDeps!.getCurrentTrackId()).toBeUndefined();
        // getCurrentTime returns 0 when audio has currentTime 0
        expect(capturedSyncDeps!.getCurrentTime()).toBe(0);

        // Set up a track
        const tracks = [createTrack(1), createTrack(2)];
        act(() => {
            result.current.replaceQueue(tracks, 0);
        });

        expect(capturedSyncDeps!.getCurrentTrackId()).toBe(1);

        // Simulate audioRef with currentTime
        engineMock.audioRef.current!.currentTime = 42;
        expect(capturedSyncDeps!.getCurrentTime()).toBe(42);

        // When audioRef.current is null, getCurrentTime returns 0
        engineMock.audioRef.current = null;
        expect(capturedSyncDeps!.getCurrentTime()).toBe(0);
    });

    describe('handleTrackEnded', () => {
        it('repeat one: restarts the current track', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });
            const tracks = [createTrack(1), createTrack(2)];

            act(() => {
                result.current.replaceQueue(tracks, 0);
            });
            act(() => {
                result.current.setRepeatMode('one');
            });

            engineMock.audioRef.current!.currentTime = 50;
            engineMock.loadAndPlayUrl.mockClear();

            act(() => {
                capturedOnTrackEnded!();
            });

            // Should restart current track, not load a new one
            expect(engineMock.audioRef.current!.currentTime).toBe(0);
        });

        it('repeat one with null audioRef does not throw', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });

            act(() => {
                result.current.replaceQueue([createTrack(1)], 0);
            });
            act(() => {
                result.current.setRepeatMode('one');
            });

            engineMock.audioRef.current = null;

            // Should not throw
            act(() => {
                capturedOnTrackEnded!();
            });
        });

        it('advances to next track when not at end of queue (repeat none)', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });
            const tracks = [createTrack(1), createTrack(2), createTrack(3)];

            act(() => {
                result.current.replaceQueue(tracks, 0);
            });
            engineMock.loadAndPlayUrl.mockClear();
            mockSyncState.mockClear();

            act(() => {
                capturedOnTrackEnded!();
            });

            expect(result.current.currentIndex).toBe(1);
            expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
                expect.stringContaining('/files/stream/2'),
                undefined
            );
            expect(mockSyncState).toHaveBeenCalledWith(
                expect.objectContaining({ fileId: 2, position: 0 })
            );
        });

        it('stops at end of queue when repeat is none', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });
            const tracks = [createTrack(1), createTrack(2)];

            act(() => {
                result.current.replaceQueue(tracks, 1);
            });
            engineMock.loadAndPlayUrl.mockClear();
            engineMock.stop.mockClear();

            act(() => {
                capturedOnTrackEnded!();
            });

            expect(engineMock.stop).toHaveBeenCalled();
        });

        it('wraps to first track when repeat is all and at end of queue', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });
            const tracks = [createTrack(1), createTrack(2)];

            act(() => {
                result.current.replaceQueue(tracks, 1);
            });
            act(() => {
                result.current.setRepeatMode('all');
            });
            engineMock.loadAndPlayUrl.mockClear();
            mockSyncState.mockClear();

            act(() => {
                capturedOnTrackEnded!();
            });

            expect(result.current.currentIndex).toBe(0);
            expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
                expect.stringContaining('/files/stream/1'),
                undefined
            );
            expect(mockSyncState).toHaveBeenCalledWith(
                expect.objectContaining({ fileId: 1, position: 0 })
            );
        });

        it('picks a shuffled track when shuffle is on', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });
            const tracks = [createTrack(1), createTrack(2), createTrack(3)];

            act(() => {
                result.current.replaceQueue(tracks, 0);
            });
            act(() => {
                result.current.toggleShuffle();
            });
            engineMock.loadAndPlayUrl.mockClear();
            mockSyncState.mockClear();

            const mathSpy = jest.spyOn(Math, 'random').mockReturnValue(0.99);
            act(() => {
                capturedOnTrackEnded!();
            });
            mathSpy.mockRestore();

            // Should have picked a random index != 0
            expect(engineMock.loadAndPlayUrl).toHaveBeenCalled();
            expect(mockSyncState).toHaveBeenCalledWith(expect.objectContaining({ position: 0 }));
        });

        it('does nothing when currentIndex is undefined', () => {
            renderHook(() => useGlobalMusic(), { wrapper });
            engineMock.loadAndPlayUrl.mockClear();
            engineMock.stop.mockClear();

            act(() => {
                capturedOnTrackEnded!();
            });

            expect(engineMock.loadAndPlayUrl).not.toHaveBeenCalled();
            expect(engineMock.stop).not.toHaveBeenCalled();
        });

        it('shuffle with single-track queue replays it when repeat is all', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });

            act(() => {
                result.current.replaceQueue([createTrack(1)], 0);
            });
            act(() => {
                result.current.toggleShuffle();
                result.current.setRepeatMode('all');
            });
            engineMock.loadAndPlayUrl.mockClear();

            act(() => {
                capturedOnTrackEnded!();
            });

            expect(result.current.currentIndex).toBe(0);
            expect(engineMock.loadAndPlayUrl).toHaveBeenCalledWith(
                expect.stringContaining('/files/stream/1'),
                undefined
            );
        });

        it('next with shuffle on single-track queue picks index 0', () => {
            const { result } = renderHook(() => useGlobalMusic(), { wrapper });

            act(() => {
                result.current.replaceQueue([createTrack(1)], 0);
            });
            act(() => {
                result.current.toggleShuffle();
            });
            engineMock.loadAndPlayUrl.mockClear();

            act(() => {
                result.current.next();
            });

            expect(result.current.currentIndex).toBe(0);
        });
    });

    describe('shuffle as a permutation', () => {
        const tracks = [1, 2, 3, 4, 5].map(createTrack);

        const startShuffledQueue = () => {
            const rendered = renderHook(() => useGlobalMusic(), { wrapper });
            act(() => {
                rendered.result.current.replaceQueue(tracks, 0);
            });
            act(() => {
                rendered.result.current.toggleShuffle();
            });
            return rendered.result;
        };

        const currentTrackId = (result: ReturnType<typeof startShuffledQueue>) =>
            result.current.currentTrack?.id;

        const advanceThroughTrackEnds = (
            result: ReturnType<typeof startShuffledQueue>,
            count: number
        ) => {
            const playedTrackIds = [currentTrackId(result)];
            for (let step = 0; step < count; step += 1) {
                act(() => {
                    capturedOnTrackEnded!();
                });
                playedTrackIds.push(currentTrackId(result));
            }
            return playedTrackIds;
        };

        it('plays every track once before stopping at the end of the order', () => {
            const result = startShuffledQueue();
            const playedTrackIds = advanceThroughTrackEnds(result, 4);

            expect([...playedTrackIds].sort()).toEqual([1, 2, 3, 4, 5]);

            engineMock.stop.mockClear();
            act(() => {
                capturedOnTrackEnded!();
            });
            expect(engineMock.stop).toHaveBeenCalled();
        });

        it('reshuffles at the end of the order when repeat is all', () => {
            const result = startShuffledQueue();
            act(() => {
                result.current.setRepeatMode('all');
            });
            const lastTrackId = advanceThroughTrackEnds(result, 4).pop();
            engineMock.stop.mockClear();

            const secondRound = advanceThroughTrackEnds(result, 5);

            expect(engineMock.stop).not.toHaveBeenCalled();
            expect(secondRound[1]).not.toBe(lastTrackId);
            expect([...secondRound.slice(1)].sort()).toEqual([1, 2, 3, 4, 5]);
        });

        it('previous returns to the previously played track', () => {
            const result = startShuffledQueue();
            const playedTrackIds = advanceThroughTrackEnds(result, 2);

            act(() => {
                result.current.previous();
            });
            expect(currentTrackId(result)).toBe(playedTrackIds[1]);

            act(() => {
                result.current.previous();
            });
            expect(currentTrackId(result)).toBe(playedTrackIds[0]);
        });

        it('next walks the same order as track end', () => {
            const result = startShuffledQueue();
            const firstNextTrackIds: (number | undefined)[] = [];
            act(() => {
                result.current.next();
            });
            firstNextTrackIds.push(currentTrackId(result));
            act(() => {
                result.current.next();
            });
            firstNextTrackIds.push(currentTrackId(result));

            expect(new Set([0, ...firstNextTrackIds]).size).toBe(3);
            expect(firstNextTrackIds).not.toContain(1);
        });

        it('turning shuffle off continues in natural order after the current track', () => {
            const result = startShuffledQueue();
            act(() => {
                result.current.playTrackFromQueue(2);
            });
            act(() => {
                result.current.next();
            });
            const currentNaturalIndex = 2;
            act(() => {
                result.current.playTrackFromQueue(currentNaturalIndex);
            });
            act(() => {
                result.current.toggleShuffle();
            });
            act(() => {
                capturedOnTrackEnded!();
            });

            expect(result.current.currentIndex).toBe((currentNaturalIndex + 1) % tracks.length);
        });

        it('turning shuffle on again builds a fresh order starting from the current track', () => {
            const result = startShuffledQueue();
            act(() => {
                result.current.toggleShuffle();
            });
            act(() => {
                result.current.playTrackFromQueue(2);
            });
            act(() => {
                result.current.toggleShuffle();
            });
            const playedTrackIds = advanceThroughTrackEnds(result, 4);

            expect(playedTrackIds[0]).toBe(3);
            expect([...playedTrackIds].sort()).toEqual([1, 2, 3, 4, 5]);
        });

        it('playNext plays the inserted tracks right after the current one', () => {
            const result = startShuffledQueue();
            act(() => {
                result.current.playNext([createTrack(90), createTrack(91)]);
            });

            const playedTrackIds = advanceThroughTrackEnds(result, 2);

            expect(playedTrackIds.slice(1)).toEqual([90, 91]);
        });

        it('addToQueue plays the appended tracks at the end of the order', () => {
            const result = startShuffledQueue();
            act(() => {
                result.current.addToQueue([createTrack(90)]);
            });

            const playedTrackIds = advanceThroughTrackEnds(result, 5);

            expect(playedTrackIds[5]).toBe(90);
        });

        it('moving a queue item keeps the order intact', () => {
            const result = startShuffledQueue();
            act(() => {
                result.current.moveQueueItem(4, 1);
            });

            const playedTrackIds = advanceThroughTrackEnds(result, 4);

            expect([...playedTrackIds].sort()).toEqual([1, 2, 3, 4, 5]);
        });

        it('removed tracks are never played', () => {
            const result = startShuffledQueue();
            const removedEntryId = result.current.queue[3]!.queueEntryId;
            act(() => {
                result.current.removeFromQueue(removedEntryId);
            });

            const playedTrackIds = advanceThroughTrackEnds(result, 3);

            expect([...playedTrackIds].sort()).toEqual([1, 2, 3, 5]);
        });

        it('removing the current track rebuilds the order from the replacement track', () => {
            const result = startShuffledQueue();
            const currentEntryId = result.current.queue[result.current.currentIndex!]!.queueEntryId;
            act(() => {
                result.current.removeFromQueue(currentEntryId);
            });

            const playedTrackIds = advanceThroughTrackEnds(result, 3);

            expect(new Set(playedTrackIds).size).toBe(4);
            expect(playedTrackIds).not.toContain(1);
        });

        it('replacing the queue while shuffling starts the new order at the chosen track', () => {
            const result = startShuffledQueue();
            act(() => {
                result.current.replaceQueue([createTrack(7), createTrack(8), createTrack(9)], 1);
            });

            const playedTrackIds = advanceThroughTrackEnds(result, 2);

            expect(playedTrackIds[0]).toBe(8);
            expect([...playedTrackIds].sort()).toEqual([7, 8, 9]);
        });
    });
});

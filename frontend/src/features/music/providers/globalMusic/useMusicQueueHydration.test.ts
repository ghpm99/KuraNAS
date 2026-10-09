import { act, renderHook } from '@testing-library/react';
import useMusicQueueHydration from './useMusicQueueHydration';
import { getPlayerQueue, getPlayerState } from '@/service/playerState';

jest.mock('@/service/playerState', () => ({
    getPlayerState: jest.fn(),
    getPlayerQueue: jest.fn(),
}));

const mockedGetPlayerState = getPlayerState as jest.Mock;
const mockedGetPlayerQueue = getPlayerQueue as jest.Mock;

const queueEntry = (fileId: number) => ({
    file_id: fileId,
    name: `${fileId}.mp3`,
    path: `/music/${fileId}.mp3`,
    format: '.mp3',
    title: `Title ${fileId}`,
    artist: 'Artist',
    album: 'Album',
    length: 200,
});

const settle = () =>
    act(async () => {
        await Promise.resolve();
        await Promise.resolve();
        await Promise.resolve();
    });

describe('useMusicQueueHydration', () => {
    const buildCallbacks = () => ({
        setQueue: jest.fn(),
        setCurrentIndex: jest.fn(),
        setShuffle: jest.fn(),
        setRepeatMode: jest.fn(),
        setVolume: jest.fn(),
        loadPausedTrack: jest.fn(),
    });

    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetPlayerState.mockRejectedValue(new Error('backend absent'));
        mockedGetPlayerQueue.mockRejectedValue(new Error('backend absent'));
    });

    it('mounts and settles without touching the player when the backend is absent', async () => {
        const callbacks = buildCallbacks();

        const { result } = renderHook(() => useMusicQueueHydration(true, callbacks));
        await settle();

        expect(result.current.hasSettled).toBe(true);
        Object.values(callbacks).forEach((callback) => expect(callback).not.toHaveBeenCalled());
    });

    it('restores queue, index, position, volume, shuffle and repeat without autoplay', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({
            items: [queueEntry(4), queueEntry(9), queueEntry(12)],
            current_index: 1,
        });
        mockedGetPlayerState.mockResolvedValue({
            current_file_id: 9,
            current_position: 87.5,
            volume: 0.35,
            shuffle: true,
            repeat_mode: 'all',
        });

        const { result } = renderHook(() => useMusicQueueHydration(true, callbacks));
        await settle();

        const restoredQueue = callbacks.setQueue.mock.calls[0][0];
        expect(restoredQueue.map((track: { id: number }) => track.id)).toEqual([4, 9, 12]);
        expect(restoredQueue[1].metadata).toMatchObject({ title: 'Title 9', length: 200 });
        expect(callbacks.setCurrentIndex).toHaveBeenCalledWith(1);
        expect(callbacks.loadPausedTrack).toHaveBeenCalledWith(9, 87.5);
        expect(callbacks.setShuffle).toHaveBeenCalledWith(true);
        expect(callbacks.setRepeatMode).toHaveBeenCalledWith('all');
        expect(callbacks.setVolume).toHaveBeenCalledWith(0.35);
        expect(result.current.hasSettled).toBe(true);
    });

    it('starts from zero when the saved state belongs to another track', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({
            items: [queueEntry(4), queueEntry(9)],
            current_index: 0,
        });
        mockedGetPlayerState.mockResolvedValue({
            current_file_id: 9,
            current_position: 50,
            volume: 2,
            shuffle: false,
            repeat_mode: 'bogus',
        });

        renderHook(() => useMusicQueueHydration(true, callbacks));
        await settle();

        expect(callbacks.loadPausedTrack).toHaveBeenCalledWith(4, 0);
        expect(callbacks.setRepeatMode).toHaveBeenCalledWith('none');
        expect(callbacks.setVolume).not.toHaveBeenCalled();
    });

    it('restores the queue even when the player state is missing', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({ items: [queueEntry(4)], current_index: 0 });

        renderHook(() => useMusicQueueHydration(true, callbacks));
        await settle();

        expect(callbacks.setQueue).toHaveBeenCalledTimes(1);
        expect(callbacks.loadPausedTrack).toHaveBeenCalledWith(4, 0);
        expect(callbacks.setShuffle).not.toHaveBeenCalled();
        expect(callbacks.setRepeatMode).not.toHaveBeenCalled();
    });

    it('clamps an out of range saved index and a missing position', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({
            items: [queueEntry(4), queueEntry(9)],
            current_index: 7,
        });
        mockedGetPlayerState.mockResolvedValue({
            current_file_id: 9,
            shuffle: false,
            repeat_mode: 'none',
        });

        renderHook(() => useMusicQueueHydration(true, callbacks));
        await settle();

        expect(callbacks.setCurrentIndex).toHaveBeenCalledWith(1);
        expect(callbacks.loadPausedTrack).toHaveBeenCalledWith(9, 0);
    });

    it('treats a non integer index as the first track', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({
            items: [queueEntry(4), queueEntry(9)],
            current_index: 0.5,
        });

        renderHook(() => useMusicQueueHydration(true, callbacks));
        await settle();

        expect(callbacks.setCurrentIndex).toHaveBeenCalledWith(0);
    });

    it('does nothing when the saved queue is empty but still settles', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({ items: [], current_index: 0 });

        const { result } = renderHook(() => useMusicQueueHydration(true, callbacks));
        await settle();

        expect(callbacks.setQueue).not.toHaveBeenCalled();
        expect(result.current.hasSettled).toBe(true);
    });

    it('does nothing when disabled', async () => {
        const callbacks = buildCallbacks();

        const { result } = renderHook(() => useMusicQueueHydration(false, callbacks));
        await settle();

        expect(mockedGetPlayerQueue).not.toHaveBeenCalled();
        expect(result.current.hasSettled).toBe(false);
    });

    it('hydrates only once even when re-enabled', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({ items: [queueEntry(4)], current_index: 0 });

        const { rerender } = renderHook(
            ({ enabled }) => useMusicQueueHydration(enabled, callbacks),
            {
                initialProps: { enabled: true },
            }
        );
        await settle();
        rerender({ enabled: false });
        rerender({ enabled: true });
        await settle();

        expect(mockedGetPlayerQueue).toHaveBeenCalledTimes(1);
    });

    it('ignores a response that arrives after unmount', async () => {
        const callbacks = buildCallbacks();
        mockedGetPlayerQueue.mockResolvedValue({ items: [queueEntry(4)], current_index: 0 });

        const { unmount } = renderHook(() => useMusicQueueHydration(true, callbacks));
        unmount();
        await settle();

        expect(callbacks.setQueue).not.toHaveBeenCalled();
    });
});

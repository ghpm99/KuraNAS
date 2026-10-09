import { act, renderHook } from '@testing-library/react';
import useMusicQueuePersistence, {
    MAX_PERSISTED_QUEUE_ENTRIES,
    QUEUE_PERSIST_DEBOUNCE_MS,
} from './useMusicQueuePersistence';
import { replacePlayerQueue } from '@/service/playerState';
import { flushPlayerQueue } from '@/service/playerStateFlush';
import type { IMusicData } from '@/types/music';

jest.mock('@/service/playerState', () => ({
    replacePlayerQueue: jest.fn(),
}));

jest.mock('@/service/playerStateFlush', () => ({
    flushPlayerQueue: jest.fn(),
}));

const track = (id: number) => ({ id }) as IMusicData;

type HookProps = { isEnabled: boolean; queue: IMusicData[]; currentIndex: number | undefined };

const renderPersistence = (initialProps: HookProps) =>
    renderHook((props: HookProps) => useMusicQueuePersistence(props), { initialProps });

describe('useMusicQueuePersistence', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        jest.clearAllMocks();
        (replacePlayerQueue as jest.Mock).mockResolvedValue(undefined);
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('mounts with an empty queue and an absent backend without writing anything', () => {
        (replacePlayerQueue as jest.Mock).mockRejectedValue(new Error('backend absent'));

        renderPersistence({ isEnabled: true, queue: [], currentIndex: undefined });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS * 2);
        });

        expect(replacePlayerQueue).not.toHaveBeenCalled();
    });

    it('does not write the baseline queue it was enabled with', () => {
        renderPersistence({ isEnabled: true, queue: [track(1), track(2)], currentIndex: 1 });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS * 2);
        });

        expect(replacePlayerQueue).not.toHaveBeenCalled();
    });

    it('debounces changes and sends only the latest queue', () => {
        const { rerender } = renderPersistence({
            isEnabled: true,
            queue: [],
            currentIndex: undefined,
        });

        rerender({ isEnabled: true, queue: [track(1)], currentIndex: 0 });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS - 1);
        });
        rerender({ isEnabled: true, queue: [track(1), track(2)], currentIndex: 1 });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS - 1);
        });
        expect(replacePlayerQueue).not.toHaveBeenCalled();

        act(() => {
            jest.advanceTimersByTime(1);
        });

        expect(replacePlayerQueue).toHaveBeenCalledTimes(1);
        expect(replacePlayerQueue).toHaveBeenCalledWith({ file_ids: [1, 2], current_index: 1 });
    });

    it('persists an emptied queue with index zero', () => {
        const { rerender } = renderPersistence({
            isEnabled: true,
            queue: [track(1)],
            currentIndex: 0,
        });

        rerender({ isEnabled: true, queue: [], currentIndex: undefined });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS);
        });

        expect(replacePlayerQueue).toHaveBeenCalledWith({ file_ids: [], current_index: 0 });
    });

    it('skips a re-render that did not change the queue', () => {
        const queue = [track(1)];
        const { rerender } = renderPersistence({ isEnabled: true, queue, currentIndex: 0 });

        rerender({ isEnabled: true, queue: [track(1)], currentIndex: 0 });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS);
        });

        expect(replacePlayerQueue).not.toHaveBeenCalled();
    });

    it('caps the persisted queue and clamps the index', () => {
        const hugeQueue = Array.from({ length: MAX_PERSISTED_QUEUE_ENTRIES + 5 }, (_, index) =>
            track(index + 1)
        );
        const { rerender } = renderPersistence({
            isEnabled: true,
            queue: [],
            currentIndex: undefined,
        });

        rerender({
            isEnabled: true,
            queue: hugeQueue,
            currentIndex: MAX_PERSISTED_QUEUE_ENTRIES + 2,
        });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS);
        });

        const request = (replacePlayerQueue as jest.Mock).mock.calls[0][0];
        expect(request.file_ids).toHaveLength(MAX_PERSISTED_QUEUE_ENTRIES);
        expect(request.current_index).toBe(MAX_PERSISTED_QUEUE_ENTRIES - 1);
    });

    it('does not persist while disabled and re-baselines when enabled again', () => {
        const { rerender } = renderPersistence({
            isEnabled: false,
            queue: [],
            currentIndex: undefined,
        });

        rerender({ isEnabled: false, queue: [track(1)], currentIndex: 0 });
        rerender({ isEnabled: true, queue: [track(1)], currentIndex: 0 });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS);
        });

        expect(replacePlayerQueue).not.toHaveBeenCalled();
    });

    it('flushes the pending queue on pagehide instead of waiting for the debounce', () => {
        const { rerender } = renderPersistence({
            isEnabled: true,
            queue: [],
            currentIndex: undefined,
        });
        rerender({ isEnabled: true, queue: [track(7)], currentIndex: 0 });

        act(() => {
            window.dispatchEvent(new Event('pagehide'));
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS);
        });

        expect(flushPlayerQueue).toHaveBeenCalledWith({ file_ids: [7], current_index: 0 });
        expect(replacePlayerQueue).not.toHaveBeenCalled();
    });

    it('does not flush on pagehide when nothing is pending', () => {
        renderPersistence({ isEnabled: true, queue: [track(1)], currentIndex: 0 });

        act(() => {
            window.dispatchEvent(new Event('pagehide'));
        });

        expect(flushPlayerQueue).not.toHaveBeenCalled();
    });

    it('does not flush a request that was already sent', () => {
        const { rerender } = renderPersistence({
            isEnabled: true,
            queue: [],
            currentIndex: undefined,
        });
        rerender({ isEnabled: true, queue: [track(7)], currentIndex: 0 });
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS);
        });

        act(() => {
            window.dispatchEvent(new Event('pagehide'));
        });

        expect(flushPlayerQueue).not.toHaveBeenCalled();
    });

    it('drops the pending write on unmount', () => {
        const { rerender, unmount } = renderPersistence({
            isEnabled: true,
            queue: [],
            currentIndex: undefined,
        });
        rerender({ isEnabled: true, queue: [track(7)], currentIndex: 0 });

        unmount();
        act(() => {
            jest.advanceTimersByTime(QUEUE_PERSIST_DEBOUNCE_MS);
        });

        expect(replacePlayerQueue).not.toHaveBeenCalled();
    });
});

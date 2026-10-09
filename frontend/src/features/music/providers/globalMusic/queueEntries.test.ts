import type { IMusicData } from '@/types/music';
import { createQueueEntries, insertAfterIndex, moveQueueEntry } from './queueEntries';

const createTrack = (id: number) => ({ id, name: `track-${id}` }) as IMusicData;

describe('queueEntries', () => {
    it('gives every entry a unique id even for the same track', () => {
        const entries = createQueueEntries([createTrack(1), createTrack(1)]);

        expect(entries[0]!.id).toBe(1);
        expect(entries[0]!.queueEntryId).not.toBe(entries[1]!.queueEntryId);
    });

    it('inserts entries right after the given index', () => {
        const queue = createQueueEntries([createTrack(1), createTrack(2)]);
        const inserted = createQueueEntries([createTrack(9)]);

        expect(insertAfterIndex(queue, 0, inserted).map((entry) => entry.id)).toEqual([1, 9, 2]);
    });

    it('moves an entry and ignores invalid moves', () => {
        const queue = createQueueEntries([createTrack(1), createTrack(2), createTrack(3)]);

        expect(moveQueueEntry(queue, 0, 2).map((entry) => entry.id)).toEqual([2, 3, 1]);
        expect(moveQueueEntry(queue, 0, 0)).toBe(queue);
        expect(moveQueueEntry(queue, -1, 1)).toBe(queue);
        expect(moveQueueEntry(queue, 0, 3)).toBe(queue);
    });
});

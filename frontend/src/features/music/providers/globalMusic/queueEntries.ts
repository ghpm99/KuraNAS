import type { IMusicData } from '@/types/music';

export type QueueTrack = IMusicData & { queueEntryId: string };

let lastQueueEntrySequence = 0;

export const createQueueEntries = (tracks: IMusicData[]): QueueTrack[] =>
    tracks.map((track) => {
        lastQueueEntrySequence += 1;
        return { ...track, queueEntryId: `queue-entry-${lastQueueEntrySequence}` };
    });

export const insertAfterIndex = (
    queue: QueueTrack[],
    index: number,
    entries: QueueTrack[]
): QueueTrack[] => [...queue.slice(0, index + 1), ...entries, ...queue.slice(index + 1)];

export const moveQueueEntry = (
    queue: QueueTrack[],
    fromIndex: number,
    toIndex: number
): QueueTrack[] => {
    const isOutOfRange = (index: number) =>
        !Number.isInteger(index) || index < 0 || index >= queue.length;
    if (isOutOfRange(fromIndex) || isOutOfRange(toIndex) || fromIndex === toIndex) return queue;
    const reorderedQueue = [...queue];
    const [movedEntry] = reorderedQueue.splice(fromIndex, 1);
    reorderedQueue.splice(toIndex, 0, movedEntry!);
    return reorderedQueue;
};

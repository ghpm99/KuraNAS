import { shuffleItems } from '@/utils/shuffleItems';
import type { QueueTrack } from './queueEntries';

export type ShuffleOrder = string[];

type RandomSource = () => number;

export const buildShuffleOrder = (
    queue: QueueTrack[],
    currentEntryId: string | undefined,
    random?: RandomSource
): ShuffleOrder => {
    const otherEntryIds = queue
        .map((entry) => entry.queueEntryId)
        .filter((entryId) => entryId !== currentEntryId);
    const shuffledOthers = shuffleItems(otherEntryIds, random);
    return currentEntryId === undefined ? shuffledOthers : [currentEntryId, ...shuffledOthers];
};

export const reshuffleOrder = (
    queue: QueueTrack[],
    lastPlayedEntryId: string | undefined,
    random?: RandomSource
): ShuffleOrder => {
    const reshuffledOrder = shuffleItems(
        queue.map((entry) => entry.queueEntryId),
        random
    );
    const startsWithLastPlayed = reshuffledOrder[0] === lastPlayedEntryId;
    if (!startsWithLastPlayed || reshuffledOrder.length < 2) return reshuffledOrder;
    return [...reshuffledOrder.slice(1), reshuffledOrder[0]!];
};

export const isOrderConsistentWithQueue = (order: ShuffleOrder, queue: QueueTrack[]): boolean => {
    if (order.length !== queue.length) return false;
    const orderedEntryIds = new Set(order);
    return queue.every((entry) => orderedEntryIds.has(entry.queueEntryId));
};

export const reconcileShuffleOrder = (
    order: ShuffleOrder,
    queue: QueueTrack[],
    currentEntryId: string | undefined,
    random?: RandomSource
): ShuffleOrder =>
    isOrderConsistentWithQueue(order, queue)
        ? order
        : buildShuffleOrder(queue, currentEntryId, random);

export const insertAfterEntry = (
    order: ShuffleOrder,
    anchorEntryId: string | undefined,
    entryIds: string[]
): ShuffleOrder => {
    const anchorPosition = anchorEntryId === undefined ? -1 : order.indexOf(anchorEntryId);
    return [...order.slice(0, anchorPosition + 1), ...entryIds, ...order.slice(anchorPosition + 1)];
};

export const appendEntries = (order: ShuffleOrder, entryIds: string[]): ShuffleOrder => [
    ...order,
    ...entryIds,
];

export const removeEntry = (order: ShuffleOrder, entryId: string): ShuffleOrder =>
    order.filter((orderedEntryId) => orderedEntryId !== entryId);

export const findEntryAfter = (order: ShuffleOrder, entryId: string): string | undefined =>
    order[order.indexOf(entryId) + 1];

export const findEntryBefore = (order: ShuffleOrder, entryId: string): string | undefined => {
    const position = order.indexOf(entryId);
    if (position === -1) return undefined;
    return position === 0 ? order[order.length - 1] : order[position - 1];
};

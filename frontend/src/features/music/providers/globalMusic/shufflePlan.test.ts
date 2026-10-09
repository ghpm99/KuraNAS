import {
    appendEntries,
    buildShuffleOrder,
    findEntryAfter,
    findEntryBefore,
    insertAfterEntry,
    isOrderConsistentWithQueue,
    reconcileShuffleOrder,
    removeEntry,
    reshuffleOrder,
} from './shufflePlan';
import type { QueueTrack } from './queueEntries';

const createQueue = (count: number): QueueTrack[] =>
    Array.from({ length: count }, (_, index) => ({
        id: index + 1,
        queueEntryId: `e${index + 1}`,
    })) as unknown as QueueTrack[];

const alwaysZero = () => 0;

describe('shufflePlan', () => {
    it('builds an order with the current entry first and every other entry exactly once', () => {
        const order = buildShuffleOrder(createQueue(6), 'e3');
        expect(order[0]).toBe('e3');
        expect([...order].sort()).toEqual(['e1', 'e2', 'e3', 'e4', 'e5', 'e6']);
    });

    it('builds an order without a current entry', () => {
        expect([...buildShuffleOrder(createQueue(3), undefined)].sort()).toEqual([
            'e1',
            'e2',
            'e3',
        ]);
    });

    it('reshuffles into a permutation that does not start with the last played entry', () => {
        const order = reshuffleOrder(createQueue(4), 'e1', alwaysZero);
        expect(order).toHaveLength(4);
        expect(order[0]).not.toBe('e1');
        expect([...order].sort()).toEqual(['e1', 'e2', 'e3', 'e4']);
    });

    it('reshuffles a single entry queue back to itself', () => {
        expect(reshuffleOrder(createQueue(1), 'e1')).toEqual(['e1']);
    });

    it('keeps the reshuffled order when it starts with another entry', () => {
        const order = reshuffleOrder(createQueue(3), 'unrelated', alwaysZero);
        expect(order).toHaveLength(3);
    });

    it('detects whether an order matches the queue', () => {
        const queue = createQueue(3);
        expect(isOrderConsistentWithQueue(['e3', 'e1', 'e2'], queue)).toBe(true);
        expect(isOrderConsistentWithQueue(['e1', 'e2'], queue)).toBe(false);
        expect(isOrderConsistentWithQueue(['e1', 'e2', 'x'], queue)).toBe(false);
    });

    it('keeps a consistent order and rebuilds a stale one', () => {
        const queue = createQueue(3);
        const consistentOrder = ['e2', 'e3', 'e1'];
        expect(reconcileShuffleOrder(consistentOrder, queue, 'e2')).toBe(consistentOrder);
        const rebuiltOrder = reconcileShuffleOrder([], queue, 'e2');
        expect(rebuiltOrder[0]).toBe('e2');
        expect(rebuiltOrder).toHaveLength(3);
    });

    it('inserts right after the anchor entry, or at the start without an anchor', () => {
        expect(insertAfterEntry(['a', 'b', 'c'], 'b', ['x'])).toEqual(['a', 'b', 'x', 'c']);
        expect(insertAfterEntry(['a', 'b'], undefined, ['x'])).toEqual(['x', 'a', 'b']);
    });

    it('appends and removes entries', () => {
        expect(appendEntries(['a'], ['b', 'c'])).toEqual(['a', 'b', 'c']);
        expect(removeEntry(['a', 'b', 'c'], 'b')).toEqual(['a', 'c']);
    });

    it('walks forward and backward through the order', () => {
        const order = ['a', 'b', 'c'];
        expect(findEntryAfter(order, 'a')).toBe('b');
        expect(findEntryAfter(order, 'c')).toBeUndefined();
        expect(findEntryBefore(order, 'c')).toBe('b');
        expect(findEntryBefore(order, 'a')).toBe('c');
        expect(findEntryBefore(order, 'missing')).toBeUndefined();
    });
});

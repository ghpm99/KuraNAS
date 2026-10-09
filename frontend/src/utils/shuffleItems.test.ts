import { shuffleItems } from './shuffleItems';

describe('shuffleItems', () => {
    it('returns an empty list for empty input', () => {
        expect(shuffleItems([])).toEqual([]);
    });

    it('does not mutate the input list', () => {
        const items = [1, 2, 3, 4];
        shuffleItems(items);
        expect(items).toEqual([1, 2, 3, 4]);
    });

    it('is deterministic with an injected random source', () => {
        const alwaysZero = () => 0;
        expect(shuffleItems([1, 2, 3, 4], alwaysZero)).toEqual([2, 3, 4, 1]);

        const almostOne = () => 0.999999;
        expect(shuffleItems([1, 2, 3, 4], almostOne)).toEqual([1, 2, 3, 4]);
    });

    it('keeps every item exactly once with the default random source', () => {
        const items = Array.from({ length: 200 }, (_, index) => index);
        const shuffledItems = shuffleItems(items);
        expect([...shuffledItems].sort((left, right) => left - right)).toEqual(items);
    });

    it('can produce every permutation of three items', () => {
        const permutations = new Set<string>();
        const randomSequences = [0, 0.34, 0.67, 0.99];
        for (const firstDraw of randomSequences) {
            for (const secondDraw of randomSequences) {
                const draws = [firstDraw, secondDraw];
                permutations.add(shuffleItems([1, 2, 3], () => draws.shift() ?? 0).join(','));
            }
        }
        expect(permutations.size).toBe(6);
    });
});

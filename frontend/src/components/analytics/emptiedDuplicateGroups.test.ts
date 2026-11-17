import { countGroupsLeftWithoutCopies } from './emptiedDuplicateGroups';
import type { DuplicateCopy, DuplicateGroupDetail } from '@/types/analytics';

const copy = (id: number): DuplicateCopy => ({
    id,
    path: `/f${id}.jpg`,
    updated_at: '2026-01-01T00:00:00Z',
    tier: 'hot',
});

const group = (signature: string, copyIds: number[]): DuplicateGroupDetail => ({
    signature,
    copies: copyIds.length,
    size_bytes: 1,
    reclaimable_size: 1,
    paths: [],
    files: copyIds.map(copy),
});

describe('countGroupsLeftWithoutCopies', () => {
    const groups = [group('a', [1, 2]), group('b', [3, 4, 5])];

    it('counts only groups whose every copy is selected', () => {
        expect(countGroupsLeftWithoutCopies(groups, [copy(1), copy(2), copy(3)])).toBe(1);
        expect(
            countGroupsLeftWithoutCopies(groups, [copy(1), copy(2), copy(3), copy(4), copy(5)])
        ).toBe(2);
    });

    it('is zero when every group keeps at least one copy', () => {
        expect(countGroupsLeftWithoutCopies(groups, [copy(1), copy(3)])).toBe(0);
        expect(countGroupsLeftWithoutCopies(groups, [])).toBe(0);
    });
});

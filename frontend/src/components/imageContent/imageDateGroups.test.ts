import { buildImageLibraryItem } from './imageLibraryTestFixtures';
import {
    buildMonthKey,
    buildTimelineCountsByMonth,
    groupImagesByMonth,
    groupImagesWithoutHeader,
    parseTakenAt,
} from './imageDateGroups';

const formatMonth = (date: Date) => `${date.getUTCFullYear()}/${date.getUTCMonth() + 1}`;

describe('imageDateGroups', () => {
    it('parses ISO dates and rejects missing or invalid values', () => {
        expect(parseTakenAt('2026-03-10T10:00:00Z')?.toISOString()).toBe(
            '2026-03-10T10:00:00.000Z'
        );
        expect(parseTakenAt(null)).toBeNull();
        expect(parseTakenAt('not a date')).toBeNull();
    });

    it('groups by UTC month keeping server order and sending undated images last', () => {
        const items = [
            buildImageLibraryItem({ file_id: 1, taken_at: null }),
            buildImageLibraryItem({ file_id: 2, taken_at: '2026-03-31T23:30:00Z' }),
            buildImageLibraryItem({ file_id: 3, taken_at: '2026-03-01T00:30:00Z' }),
            buildImageLibraryItem({ file_id: 4, taken_at: '2026-02-10T10:00:00Z' }),
            buildImageLibraryItem({ file_id: 5, taken_at: 'garbage' }),
        ];

        const groups = groupImagesByMonth({
            items,
            formatMonth,
            undatedLabel: 'Sem data',
            countsByMonth: buildTimelineCountsByMonth([{ year: 2026, month: 3, count: 40 }]),
        });

        expect(groups.map((group) => group.label)).toEqual(['2026/3', '2026/2', 'Sem data']);
        expect(groups[0]!.items.map((item) => item.file_id)).toEqual([2, 3]);
        expect(groups[2]!.items.map((item) => item.file_id)).toEqual([1, 5]);
        expect(groups[0]!.totalCount).toBe(40);
        expect(groups[1]!.totalCount).toBeNull();
        expect(groups[2]!.totalCount).toBeNull();
    });

    it('returns no groups for no items and a headerless group otherwise', () => {
        expect(
            groupImagesByMonth({
                items: [],
                formatMonth,
                undatedLabel: 'x',
                countsByMonth: new Map(),
            })
        ).toEqual([]);
        expect(groupImagesWithoutHeader([])).toEqual([]);
        const [group] = groupImagesWithoutHeader([buildImageLibraryItem()]);
        expect(group!.label).toBe('');
        expect(buildMonthKey(2026, 3)).toBe('2026-3');
    });
});

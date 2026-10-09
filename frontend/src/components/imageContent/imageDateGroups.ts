import type { ImageLibraryItem, ImageTimelineBucket } from '@/types/imageLibrary';

export type ImageDateGroup = {
    key: string;
    label: string;
    items: ImageLibraryItem[];
    totalCount: number | null;
};

const undatedGroupKey = 'undated';

export const buildMonthKey = (year: number, month: number) => `${year}-${month}`;

export const parseTakenAt = (takenAt: string | null): Date | null => {
    if (!takenAt) {
        return null;
    }
    const parsedDate = new Date(takenAt);
    return Number.isNaN(parsedDate.getTime()) ? null : parsedDate;
};

export const buildTimelineCountsByMonth = (buckets: ImageTimelineBucket[]) =>
    new Map(buckets.map((bucket) => [buildMonthKey(bucket.year, bucket.month), bucket.count]));

type GroupByMonthParams = {
    items: ImageLibraryItem[];
    formatMonth: (date: Date) => string;
    undatedLabel: string;
    countsByMonth: Map<string, number>;
};

export const groupImagesByMonth = ({
    items,
    formatMonth,
    undatedLabel,
    countsByMonth,
}: GroupByMonthParams): ImageDateGroup[] => {
    const groupsByKey = new Map<string, ImageDateGroup>();

    for (const item of items) {
        const takenAt = parseTakenAt(item.taken_at);
        const key = takenAt
            ? buildMonthKey(takenAt.getUTCFullYear(), takenAt.getUTCMonth() + 1)
            : undatedGroupKey;
        const existingGroup = groupsByKey.get(key);
        if (existingGroup) {
            existingGroup.items.push(item);
            continue;
        }
        groupsByKey.set(key, {
            key,
            label: takenAt ? formatMonth(takenAt) : undatedLabel,
            items: [item],
            totalCount: takenAt ? (countsByMonth.get(key) ?? null) : null,
        });
    }

    const undatedGroup = groupsByKey.get(undatedGroupKey);
    groupsByKey.delete(undatedGroupKey);
    return undatedGroup ? [...groupsByKey.values(), undatedGroup] : [...groupsByKey.values()];
};

export const groupImagesWithoutHeader = (items: ImageLibraryItem[]): ImageDateGroup[] =>
    items.length === 0 ? [] : [{ key: 'all', label: '', items, totalCount: null }];

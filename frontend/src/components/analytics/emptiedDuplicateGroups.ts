import type { DuplicateCopy, DuplicateGroupDetail } from '@/types/analytics';

export const countGroupsLeftWithoutCopies = (
    groups: DuplicateGroupDetail[],
    copiesToDelete: DuplicateCopy[]
): number => {
    const idsToDelete = new Set(copiesToDelete.map((copy) => copy.id));
    return groups.filter(
        (group) => group.files.length > 0 && group.files.every((copy) => idsToDelete.has(copy.id))
    ).length;
};

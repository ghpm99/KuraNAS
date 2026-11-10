import type { TrashItem } from '@/types/trash';

const MILLISECONDS_PER_DAY = 24 * 60 * 60 * 1000;

const splitPathSegments = (originalPath: string): string[] =>
    originalPath.split(/[\\/]/).filter((segment) => segment !== '');

export const trashItemName = (item: TrashItem): string => {
    if (item.name) {
        return item.name;
    }
    const pathSegments = splitPathSegments(item.original_path ?? '');
    return pathSegments[pathSegments.length - 1] ?? '';
};

export const parentFolderOf = (path: string): string => {
    const lastSeparatorIndex = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
    if (lastSeparatorIndex <= 0) {
        return '/';
    }
    return path.slice(0, lastSeparatorIndex);
};

export const trashItemOriginFolder = (item: TrashItem): string =>
    parentFolderOf(item.original_path ?? '');

export const daysUntilExpiry = (
    expiresAt: string | undefined,
    currentTime: Date
): number | undefined => {
    if (!expiresAt) {
        return undefined;
    }
    const expiryTime = new Date(expiresAt).getTime();
    if (Number.isNaN(expiryTime)) {
        return undefined;
    }
    const remainingDays = Math.ceil((expiryTime - currentTime.getTime()) / MILLISECONDS_PER_DAY);
    return Math.max(0, remainingDays);
};

import type { FilesSort, FilesSortKey, FilesSortOrder } from './fileContext';

export const defaultFilesSort: FilesSort = { key: 'name', order: 'asc' };

export const filesSortStorageKey = 'kuranas.files.sort';

const validSortKeys: FilesSortKey[] = ['name', 'size', 'updated_at', 'created_at'];
const validSortOrders: FilesSortOrder[] = ['asc', 'desc'];

const isFilesSort = (candidate: unknown): candidate is FilesSort => {
    if (typeof candidate !== 'object' || candidate === null) {
        return false;
    }
    const { key, order } = candidate as Partial<FilesSort>;
    return (
        validSortKeys.includes(key as FilesSortKey) &&
        validSortOrders.includes(order as FilesSortOrder)
    );
};

export const loadFilesSort = (): FilesSort => {
    try {
        const storedSort = window.localStorage.getItem(filesSortStorageKey);
        if (!storedSort) {
            return defaultFilesSort;
        }
        const parsedSort: unknown = JSON.parse(storedSort);
        return isFilesSort(parsedSort) ? parsedSort : defaultFilesSort;
    } catch {
        return defaultFilesSort;
    }
};

export const saveFilesSort = (sort: FilesSort): void => {
    try {
        window.localStorage.setItem(filesSortStorageKey, JSON.stringify(sort));
    } catch {
        return;
    }
};

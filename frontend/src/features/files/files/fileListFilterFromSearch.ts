import type { FileListCategoryType } from '@/features/files/providers/fileProvider/fileContext';

const filtersSelectableByUrl: FileListCategoryType[] = ['recent', 'starred'];

export const fileListFilterSearchParam = 'filter';

export const parseFileListFilter = (search: string): FileListCategoryType => {
    const requestedFilter = new URLSearchParams(search).get(fileListFilterSearchParam);
    return filtersSelectableByUrl.find((filter) => filter === requestedFilter) ?? 'all';
};

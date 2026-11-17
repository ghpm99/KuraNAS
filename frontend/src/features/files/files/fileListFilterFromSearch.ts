import type { FileListCategoryType } from '@/features/files/providers/fileProvider/fileContext';

export const parseFileListFilter = (search: string): FileListCategoryType =>
    new URLSearchParams(search).get('filter') === 'recent' ? 'recent' : 'all';

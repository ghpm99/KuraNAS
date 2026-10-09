import { FileType } from '@/utils';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

export const resolveListedFiles = (
    openedItem: FileData | null,
    rootFiles: FileData[] | undefined
): FileData[] => {
    if (!openedItem) return rootFiles ?? [];
    if (openedItem.type !== FileType.Directory) return [];
    return openedItem.file_children ?? [];
};

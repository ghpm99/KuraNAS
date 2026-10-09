import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import type { ImageLibraryItem } from '@/types/imageLibrary';

export const mapFileDataToLibraryItem = (file: FileData): ImageLibraryItem => ({
    file_id: file.id,
    name: file.name,
    path: file.path,
    parent_path: file.parent_path,
    format: file.format,
    size: file.size,
    width: 0,
    height: 0,
    taken_at: null,
    category: 'other',
    starred: file.starred,
    tier: file.tier ?? '',
    updated_at: file.updated_at,
});

export const readImageLibraryItemId = (image: ImageLibraryItem) => image.file_id;

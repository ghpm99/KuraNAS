import type { ImageLibraryItem } from '@/types/imageLibrary';
import type { ImageCollectionCard } from './components/ImageCollectionsPanel';

const getFolderTitle = (folderPath: string) => {
    const segments = folderPath.replace(/\\/g, '/').split('/').filter(Boolean);
    return segments.length > 0 ? segments[segments.length - 1]! : folderPath;
};

export const buildFolderCards = (items: ImageLibraryItem[]): ImageCollectionCard[] => {
    const cardsByFolder = new Map<string, ImageCollectionCard>();

    for (const item of items) {
        if (cardsByFolder.has(item.parent_path)) {
            continue;
        }
        cardsByFolder.set(item.parent_path, {
            id: item.parent_path,
            title: getFolderTitle(item.parent_path),
            description: item.parent_path,
            coverImageId: item.file_id,
        });
    }

    return [...cardsByFolder.values()];
};

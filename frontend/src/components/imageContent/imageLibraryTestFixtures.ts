import type { ImageLibraryItem } from '@/types/imageLibrary';

export const buildImageLibraryItem = (
    overrides: Partial<ImageLibraryItem> = {}
): ImageLibraryItem => ({
    file_id: 7,
    name: 'Trip.jpg',
    path: '/photos/travel/Trip.jpg',
    parent_path: '/photos/travel',
    format: '.jpg',
    size: 2048,
    width: 1600,
    height: 900,
    taken_at: '2026-03-10T10:00:00Z',
    category: 'photo',
    starred: false,
    tier: 'hot',
    updated_at: '2026-03-10T10:00:00Z',
    ...overrides,
});

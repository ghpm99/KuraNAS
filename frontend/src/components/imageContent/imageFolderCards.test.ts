import { buildImageLibraryItem } from './imageLibraryTestFixtures';
import { buildFolderCards } from './imageFolderCards';
import { mapFileDataToLibraryItem, readImageLibraryItemId } from './imageLibraryItemMapping';
import { getImageCategoryLabelKey } from './imageCategoryLabels';

describe('imageFolderCards', () => {
    it('builds one card per folder using the first image as cover', () => {
        const cards = buildFolderCards([
            buildImageLibraryItem({ file_id: 1, parent_path: '/photos/trip' }),
            buildImageLibraryItem({ file_id: 2, parent_path: '/photos/trip' }),
            buildImageLibraryItem({ file_id: 3, parent_path: '/' }),
            buildImageLibraryItem({ file_id: 4, parent_path: 'C:\\Fotos\\Casa' }),
        ]);

        expect(cards).toEqual([
            { id: '/photos/trip', title: 'trip', description: '/photos/trip', coverImageId: 1 },
            { id: '/', title: '/', description: '/', coverImageId: 3 },
            {
                id: 'C:\\Fotos\\Casa',
                title: 'Casa',
                description: 'C:\\Fotos\\Casa',
                coverImageId: 4,
            },
        ]);
        expect(buildFolderCards([])).toEqual([]);
    });

    it('maps a file record to a library item for deep links', () => {
        const item = mapFileDataToLibraryItem({
            id: 9,
            name: 'a.jpg',
            path: '/p/a.jpg',
            parent_path: '/p',
            type: 2,
            format: '.jpg',
            size: 10,
            updated_at: 'u',
            created_at: 'c',
            deleted_at: '',
            last_interaction: '',
            last_backup: '',
            check_sum: '',
            directory_content_count: 0,
            starred: true,
        });

        expect(item).toEqual(
            expect.objectContaining({ file_id: 9, parent_path: '/p', starred: true, tier: '' })
        );
        expect(readImageLibraryItemId(item)).toBe(9);
    });

    it('resolves a label key for every known and unknown category', () => {
        expect(getImageCategoryLabelKey('screenshot_app')).toBe(
            'IMAGES_CLASSIFICATION_SCREENSHOT_APP'
        );
        expect(getImageCategoryLabelKey(undefined)).toBe('IMAGES_CLASSIFICATION_OTHER');
        expect(getImageCategoryLabelKey('weird')).toBe('IMAGES_CLASSIFICATION_OTHER');
    });
});

import { mapFileDataToLibraryItem, readImageLibraryItemId } from './imageLibraryItemMapping';

describe('imageLibraryItemMapping', () => {
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
});

import type { FileData } from './fileContext';
import { mergeChildrenIntoTree } from './fileProviderUtils';

const makeNode = (id: number, name: string, fileChildren?: FileData[]): FileData => ({
    id,
    name,
    path: `/${name}`,
    parent_path: '/',
    type: 1,
    format: '',
    size: 0,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    file_children: fileChildren,
});

describe('mergeChildrenIntoTree', () => {
    it('attaches children to the matching nested parent', () => {
        const tree = [makeNode(1, 'a', [makeNode(2, 'b')])];

        const merged = mergeChildrenIntoTree(tree, 2, [makeNode(3, 'c')]);

        expect(merged[0]?.file_children?.[0]?.file_children?.map((node) => node.id)).toEqual([3]);
    });

    it('keeps already loaded grandchildren of children that are still present', () => {
        const tree = [makeNode(1, 'a', [makeNode(2, 'b', [makeNode(3, 'c')])])];

        const merged = mergeChildrenIntoTree(tree, 1, [makeNode(2, 'b'), makeNode(4, 'd')]);

        expect(merged[0]?.file_children?.map((node) => node.id)).toEqual([2, 4]);
        expect(merged[0]?.file_children?.[0]?.file_children?.map((node) => node.id)).toEqual([3]);
    });

    it('returns the tree untouched when the parent is not loaded', () => {
        const tree = [makeNode(1, 'a'), makeNode(5, 'e', [makeNode(6, 'f')])];

        const merged = mergeChildrenIntoTree(tree, 99, [makeNode(7, 'g')]);

        expect(merged).toEqual(tree);
    });
});

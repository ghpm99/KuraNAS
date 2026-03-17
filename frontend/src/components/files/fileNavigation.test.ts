import type { FileData } from '@/components/providers/fileProvider/fileContext';
import { findTrailById } from './fileNavigation';

const makeFile = (overrides: Partial<FileData> & { id: number; name: string }): FileData => ({
    path: `/${overrides.name}`,
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
    ...overrides,
});

describe('findTrailById', () => {
    const childNode = makeFile({
        id: 3,
        name: 'child',
        path: '/root/parent/child',
        parent_path: '/root/parent',
    });
    const parentNode = makeFile({
        id: 2,
        name: 'parent',
        path: '/root/parent',
        parent_path: '/root',
        file_children: [childNode],
    });
    const rootNode = makeFile({
        id: 1,
        name: 'root',
        path: '/root',
        file_children: [parentNode],
    });
    const nodes: FileData[] = [rootNode];

    it('returns null when targetId is null', () => {
        expect(findTrailById(nodes, null)).toBeNull();
    });

    it('returns null when targetId is 0 (falsy)', () => {
        expect(findTrailById(nodes, 0)).toBeNull();
    });

    it('returns trail for a top-level node', () => {
        const trail = findTrailById(nodes, 1);
        expect(trail).toEqual([rootNode]);
    });

    it('returns trail for a nested node', () => {
        const trail = findTrailById(nodes, 3);
        expect(trail).toEqual([rootNode, parentNode, childNode]);
    });

    it('returns null when targetId is not found in the tree', () => {
        expect(findTrailById(nodes, 999)).toBeNull();
    });

    it('returns null for an empty nodes array', () => {
        expect(findTrailById([], 1)).toBeNull();
    });

    it('handles nodes with empty file_children array', () => {
        const leaf = makeFile({ id: 10, name: 'leaf', file_children: [] });
        expect(findTrailById([leaf], 99)).toBeNull();
    });
});

import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { findSiblingFiles, locateAmongSiblings } from './siblingFiles';

const createNode = (overrides: Partial<FileData>): FileData =>
    ({
        id: 1,
        name: 'n',
        path: '/n',
        parent_path: '/',
        type: 2,
        format: '.txt',
        size: 1,
        ...overrides,
    }) as FileData;

const photoA = createNode({ id: 10, name: 'a.jpg', path: '/pics/a.jpg', parent_path: '/pics' });
const photoB = createNode({ id: 11, name: 'b.jpg', path: '/pics/b.jpg', parent_path: '/pics' });
const subfolder = createNode({
    id: 12,
    name: 'sub',
    path: '/pics/sub',
    parent_path: '/pics',
    type: 1,
});
const photoC = createNode({ id: 13, name: 'c.jpg', path: '/pics/c.jpg', parent_path: '/pics' });
const picsFolder = createNode({
    id: 2,
    name: 'pics',
    path: '/pics',
    parent_path: '/',
    type: 1,
    file_children: [photoA, subfolder, photoB, photoC],
});
const rootFolder = createNode({
    id: 1,
    name: 'root',
    path: '/',
    parent_path: '',
    type: 1,
    file_children: [picsFolder],
});

describe('findSiblingFiles', () => {
    it('returns the files next to the given one, skipping folders, from the loaded tree', () => {
        const siblingFiles = findSiblingFiles([rootFolder], photoB);

        expect(siblingFiles.map((sibling) => sibling.id)).toEqual([10, 11, 13]);
    });

    it('uses the top-level listing when the file is a root item', () => {
        const rootFile = createNode({ id: 99, path: '/x.txt', parent_path: '/' });

        expect(findSiblingFiles([rootFile, picsFolder], rootFile)).toEqual([rootFile]);
    });

    it('returns nothing when the parent folder was never loaded', () => {
        expect(findSiblingFiles([], photoA)).toEqual([]);
        expect(findSiblingFiles([createNode({ id: 5, path: '/other', type: 1 })], photoA)).toEqual(
            []
        );
    });
});

describe('locateAmongSiblings', () => {
    const siblingFiles = [photoA, photoB, photoC];

    it('reports neighbours and position, disabled at both ends', () => {
        expect(locateAmongSiblings(siblingFiles, 10)).toEqual({
            previous: null,
            next: photoB,
            position: 1,
            total: 3,
        });
        expect(locateAmongSiblings(siblingFiles, 11)).toEqual({
            previous: photoA,
            next: photoC,
            position: 2,
            total: 3,
        });
        expect(locateAmongSiblings(siblingFiles, 13)).toEqual({
            previous: photoB,
            next: null,
            position: 3,
            total: 3,
        });
    });

    it('returns null when the file is not among the siblings', () => {
        expect(locateAmongSiblings(siblingFiles, 404)).toBeNull();
        expect(locateAmongSiblings([], 10)).toBeNull();
    });
});

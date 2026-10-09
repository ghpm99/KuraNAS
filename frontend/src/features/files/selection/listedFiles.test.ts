import { FileType } from '@/utils';
import { resolveListedFiles } from './listedFiles';
import { createTestFile } from './testFileFactory';

describe('resolveListedFiles', () => {
    it('lists root files when nothing is opened', () => {
        const rootFiles = [createTestFile(1)];
        expect(resolveListedFiles(null, rootFiles)).toBe(rootFiles);
        expect(resolveListedFiles(null, undefined)).toEqual([]);
    });

    it('lists the children of an opened directory', () => {
        const children = [createTestFile(2)];
        const directory = createTestFile(1, { type: FileType.Directory, file_children: children });
        expect(resolveListedFiles(directory, [])).toBe(children);
        expect(resolveListedFiles({ ...directory, file_children: undefined }, [])).toEqual([]);
    });

    it('lists nothing when a file is opened', () => {
        expect(resolveListedFiles(createTestFile(1, { type: FileType.File }), [createTestFile(2)])).toEqual([]);
    });
});

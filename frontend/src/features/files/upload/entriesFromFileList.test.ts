import { entriesFromFileList } from './entriesFromFileList';

describe('entriesFromFileList', () => {
    it('keeps plain files without a relative path', () => {
        const file = new File(['x'], 'a.txt');
        expect(entriesFromFileList([file])).toEqual([{ file, relativePath: undefined }]);
    });

    it('uses webkitRelativePath for folder picks', () => {
        const file = new File(['x'], 'a.mp3');
        Object.defineProperty(file, 'webkitRelativePath', { value: 'Album/Disc1/a.mp3' });
        expect(entriesFromFileList([file])).toEqual([{ file, relativePath: 'Album/Disc1/a.mp3' }]);
    });
});

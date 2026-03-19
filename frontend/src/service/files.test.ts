jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
        delete: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    getFilesTree,
    getRecentAccessByFileId,
    getFileByPath,
    toggleStarredFile,
    rescanFiles,
    uploadFiles,
    createFolder,
    moveFile,
    copyFile,
    renameFile,
    deleteFile,
    downloadFileBlob,
    getMusicFiles,
    getImageFiles,
} from './files';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
    delete: jest.Mock;
};

describe('service/files', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('gets files tree with pagination params', async () => {
        const payload = { items: [], total: 0 };
        mockedApi.get.mockResolvedValue({ data: payload });

        const result = await getFilesTree({
            page: 1,
            pageSize: 20,
            fileParent: 5,
            category: 'all',
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/files/tree', {
            params: {
                page: 1,
                page_size: 20,
                file_parent: 5,
                category: 'all',
            },
        });
        expect(result).toEqual(payload);
    });

    it('gets recent access by file id', async () => {
        const payload = [{ id: 1, accessedAt: '2024-01-01' }];
        mockedApi.get.mockResolvedValue({ data: payload });

        const result = await getRecentAccessByFileId(42);

        expect(mockedApi.get).toHaveBeenCalledWith('/files/recent/42');
        expect(result).toEqual(payload);
    });

    it('gets file by path when item exists', async () => {
        const file = { id: 1, name: 'test.txt' };
        mockedApi.get.mockResolvedValue({ data: { items: [file] } });

        const result = await getFileByPath('/docs/test.txt');

        expect(mockedApi.get).toHaveBeenCalledWith('/files/path', {
            params: { path: '/docs/test.txt' },
        });
        expect(result).toEqual(file);
    });

    it('returns null when getFileByPath has empty items', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [] } });

        const result = await getFileByPath('/nonexistent');

        expect(result).toBeNull();
    });

    it('toggles starred file', async () => {
        mockedApi.post.mockResolvedValue({});

        await toggleStarredFile(10);

        expect(mockedApi.post).toHaveBeenCalledWith('/files/starred/10');
    });

    it('rescans files', async () => {
        mockedApi.post.mockResolvedValue({});

        await rescanFiles();

        expect(mockedApi.post).toHaveBeenCalledWith('/files/update', expect.any(FormData), {
            headers: { 'Content-Type': 'multipart/form-data' },
        });
    });

    it('uploads files with targetFolderId', async () => {
        mockedApi.post.mockResolvedValue({});

        const file = new File(['content'], 'photo.jpg');
        const fileList = {
            [Symbol.iterator]: function* () {
                yield file;
            },
            length: 1,
            item: () => file,
            0: file,
        } as unknown as FileList;

        await uploadFiles(fileList, 5);

        expect(mockedApi.post).toHaveBeenCalledWith('/files/upload', expect.any(FormData), {
            headers: { 'Content-Type': 'multipart/form-data' },
        });

        const formData: FormData = mockedApi.post.mock.calls[0][1];
        expect(formData.get('files')).toBeTruthy();
        expect(formData.get('target_folder_id')).toBe('5');
    });

    it('uploads files without targetFolderId', async () => {
        mockedApi.post.mockResolvedValue({});

        const file = new File(['content'], 'photo.jpg');
        const fileList = {
            [Symbol.iterator]: function* () {
                yield file;
            },
            length: 1,
            item: () => file,
            0: file,
        } as unknown as FileList;

        await uploadFiles(fileList);

        const formData: FormData = mockedApi.post.mock.calls[0][1];
        expect(formData.get('target_folder_id')).toBeNull();
    });

    it('creates folder with parentId', async () => {
        mockedApi.post.mockResolvedValue({});

        await createFolder('new-folder', 10);

        expect(mockedApi.post).toHaveBeenCalledWith('/files/folder', {
            name: 'new-folder',
            parent_id: 10,
        });
    });

    it('creates folder without parentId', async () => {
        mockedApi.post.mockResolvedValue({});

        await createFolder('new-folder');

        expect(mockedApi.post).toHaveBeenCalledWith('/files/folder', {
            name: 'new-folder',
            parent_id: null,
        });
    });

    it('moves file by id', async () => {
        mockedApi.post.mockResolvedValue({});

        await moveFile(1, 2, '');

        expect(mockedApi.post).toHaveBeenCalledWith('/files/move', {
            source_id: 1,
            destination_folder_id: 2,
            destination_path: '',
        });
    });

    it('moves file by path', async () => {
        mockedApi.post.mockResolvedValue({});

        await moveFile(1, undefined, '/target/dir');

        expect(mockedApi.post).toHaveBeenCalledWith('/files/move', {
            source_id: 1,
            destination_folder_id: null,
            destination_path: '/target/dir',
        });
    });

    it('copies file by id', async () => {
        mockedApi.post.mockResolvedValue({});

        await copyFile(1, 2, '', 'copy.txt');

        expect(mockedApi.post).toHaveBeenCalledWith('/files/copy', {
            source_id: 1,
            destination_folder_id: 2,
            destination_path: '',
            new_name: 'copy.txt',
        });
    });

    it('copies file by path', async () => {
        mockedApi.post.mockResolvedValue({});

        await copyFile(1, undefined, '/target/dir');

        expect(mockedApi.post).toHaveBeenCalledWith('/files/copy', {
            source_id: 1,
            destination_folder_id: null,
            destination_path: '/target/dir',
            new_name: '',
        });
    });

    it('renames file by id', async () => {
        mockedApi.post.mockResolvedValue({});

        await renameFile(5, 'new.txt');

        expect(mockedApi.post).toHaveBeenCalledWith('/files/rename', {
            id: 5,
            new_name: 'new.txt',
        });
    });

    it('deletes file by id', async () => {
        mockedApi.delete.mockResolvedValue({});

        await deleteFile(42);

        expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', {
            data: { id: 42 },
        });
    });

    it('downloads file blob', async () => {
        const blob = new Blob(['data']);
        mockedApi.get.mockResolvedValue({ data: blob });

        const result = await downloadFileBlob(99);

        expect(mockedApi.get).toHaveBeenCalledWith('/files/blob/99', {
            responseType: 'blob',
        });
        expect(result).toEqual(blob);
    });

    it('gets music files', async () => {
        const payload = { items: [], total: 0 };
        mockedApi.get.mockResolvedValue({ data: payload });

        const result = await getMusicFiles(1, 50);

        expect(mockedApi.get).toHaveBeenCalledWith('/files/music', {
            params: { page: 1, page_size: 50 },
        });
        expect(result).toEqual(payload);
    });

    it('gets image files', async () => {
        const payload = { items: [], total: 0 };
        mockedApi.get.mockResolvedValue({ data: payload });

        const result = await getImageFiles(1, 30, 'date');

        expect(mockedApi.get).toHaveBeenCalledWith('/files/images', {
            params: { page: 1, page_size: 30, group_by: 'date' },
        });
        expect(result).toEqual(payload);
    });
});

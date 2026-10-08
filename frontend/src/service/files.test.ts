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
    searchFiles,
    getStarredFiles,
    getRecentlyAccessedFiles,
    getRecentAccessByFileId,
    getFileByPath,
    getFileLocation,
    getFileAncestors,
    getFileByDiskPath,
    toggleStarredFile,
    rescanFiles,
    uploadSingleFile,
    createFolder,
    moveFile,
    copyFile,
    renameFile,
    deleteFile,
    getFileDownloadUrl,
    getFilesZipDownloadUrl,
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

    it('gets the file location by id', async () => {
        const location = { file_id: 7, tier: 'cold', disk_path: '/cold/a' };
        mockedApi.get.mockResolvedValue({ data: location });

        await expect(getFileLocation(7)).resolves.toEqual(location);
        expect(mockedApi.get).toHaveBeenCalledWith('/files/location/7');
    });

    it('gets the ancestor folders of a file sending only the id in the url', async () => {
        const ancestors = [
            { id: 1, name: 'Midia', path: '/Midia', type: 1 },
            { id: 4, name: 'fotos', path: '/Midia/fotos', type: 1 },
        ];
        mockedApi.get.mockResolvedValue({ data: ancestors });

        await expect(getFileAncestors(9)).resolves.toEqual(ancestors);
        expect(mockedApi.get).toHaveBeenCalledWith('/files/ancestors/9');
    });

    it('gets a file by disk path sending only the path query param', async () => {
        const file = { id: 3, name: 'a.txt' };
        mockedApi.get.mockResolvedValue({ data: file });

        await expect(getFileByDiskPath('D:\\Cold\\a.txt')).resolves.toEqual(file);
        expect(mockedApi.get).toHaveBeenCalledWith('/files/by-disk-path', {
            params: { path: 'D:\\Cold\\a.txt' },
        });
    });

    it('searches files sending the snake_case params the backend decodes', async () => {
        const payload = { items: [], pagination: { page: 2 } };
        mockedApi.get.mockResolvedValue({ data: payload });

        await expect(
            searchFiles({ q: 'relatorio', parentId: 7, recursive: false, page: 2, pageSize: 100 })
        ).resolves.toEqual(payload);
        expect(mockedApi.get).toHaveBeenCalledWith('/files/search', {
            params: {
                q: 'relatorio',
                parent_id: 7,
                recursive: false,
                page: 2,
                page_size: 100,
            },
        });
    });

    it('searches globally leaving parent_id undefined', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [] } });

        await searchFiles({ q: 'foto', page: 1, pageSize: 50 });

        expect(mockedApi.get).toHaveBeenCalledWith('/files/search', {
            params: {
                q: 'foto',
                parent_id: undefined,
                recursive: undefined,
                page: 1,
                page_size: 50,
            },
        });
    });

    it('gets starred files from the global endpoint with pagination params', async () => {
        const payload = { items: [], pagination: { page: 2 } };
        mockedApi.get.mockResolvedValue({ data: payload });

        const result = await getStarredFiles({ page: 2, pageSize: 50 });

        expect(mockedApi.get).toHaveBeenCalledWith('/files/starred', {
            params: { page: 2, page_size: 50 },
        });
        expect(result).toEqual(payload);
    });

    it('gets recently accessed files from the global endpoint with pagination params', async () => {
        const payload = { items: [], pagination: { page: 1 } };
        mockedApi.get.mockResolvedValue({ data: payload });

        const result = await getRecentlyAccessedFiles({ page: 1, pageSize: 30 });

        expect(mockedApi.get).toHaveBeenCalledWith('/files/recent-files', {
            params: { page: 1, page_size: 30 },
        });
        expect(result).toEqual(payload);
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
                sort: undefined,
                order: undefined,
            },
        });
        expect(result).toEqual(payload);
    });

    it('gets files tree with optional sort params', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [] } });

        await getFilesTree({
            page: 1,
            pageSize: 20,
            category: 'all',
            sort: { key: 'updated_at', order: 'desc' },
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/files/tree', {
            params: {
                page: 1,
                page_size: 20,
                file_parent: undefined,
                category: 'all',
                sort: 'updated_at',
                order: 'desc',
            },
        });
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

    describe('uploadSingleFile', () => {
        const file = new File(['content'], 'photo.jpg');

        it('posts one file with conflict policy, folder and relative path fields', async () => {
            mockedApi.post.mockResolvedValue({});
            const controller = new AbortController();

            await uploadSingleFile({
                file,
                targetFolderId: 5,
                relativePath: 'Album/photo.jpg',
                onConflict: 'replace',
                signal: controller.signal,
            });

            expect(mockedApi.post).toHaveBeenCalledWith(
                '/files/upload',
                expect.any(FormData),
                expect.objectContaining({
                    headers: { 'Content-Type': 'multipart/form-data' },
                    signal: controller.signal,
                    onUploadProgress: expect.any(Function),
                })
            );
            const formData: FormData = mockedApi.post.mock.calls[0][1];
            expect(formData.getAll('files')).toHaveLength(1);
            expect((formData.get('files') as File).name).toBe('photo.jpg');
            expect(formData.get('on_conflict')).toBe('replace');
            expect(formData.get('target_folder_id')).toBe('5');
            expect(formData.get('relative_paths')).toBe('Album/photo.jpg');
        });

        it('omits optional fields when not provided', async () => {
            mockedApi.post.mockResolvedValue({});

            await uploadSingleFile({ file, onConflict: 'rename' });

            const formData: FormData = mockedApi.post.mock.calls[0][1];
            expect(formData.get('target_folder_id')).toBeNull();
            expect(formData.get('relative_paths')).toBeNull();
            expect(formData.get('on_conflict')).toBe('rename');
        });

        it('reports progress as a percentage', async () => {
            mockedApi.post.mockImplementation(async (_url, _body, config) => {
                config.onUploadProgress({ loaded: 25, total: 100 });
                config.onUploadProgress({ loaded: 10 });
                return {};
            });
            const onProgress = jest.fn();

            await uploadSingleFile({ file, onConflict: 'skip', onProgress });

            expect(onProgress).toHaveBeenNthCalledWith(1, 25);
            expect(onProgress).toHaveBeenNthCalledWith(2, 0);
        });

        it('returns the per-file outcome from the response', async () => {
            mockedApi.post.mockResolvedValue({
                data: { files: [{ name: 'photo (2).jpg', status: 'renamed' }] },
            });

            await expect(uploadSingleFile({ file, onConflict: 'rename' })).resolves.toEqual({
                status: 'renamed',
                name: 'photo (2).jpg',
                error: undefined,
            });
        });

        it('treats a response without per-file results as uploaded', async () => {
            mockedApi.post.mockResolvedValue({ data: {} });

            await expect(uploadSingleFile({ file, onConflict: 'rename' })).resolves.toEqual({
                status: 'uploaded',
            });
        });
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

    it('returns the new path from move and rename responses', async () => {
        mockedApi.post.mockResolvedValue({ data: { path: '/new/place' } });

        await expect(moveFile(1, 2, '')).resolves.toBe('/new/place');
        await expect(renameFile(5, 'x')).resolves.toBe('/new/place');
    });

    it('returns an empty path when move and rename responses omit it', async () => {
        mockedApi.post.mockResolvedValue({});

        await expect(moveFile(1, 2, '')).resolves.toBe('');
        await expect(renameFile(5, 'x')).resolves.toBe('');
    });

    it('deletes file by id', async () => {
        mockedApi.delete.mockResolvedValue({});

        await deleteFile(42);

        expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', {
            data: { id: 42 },
        });
    });

    it('builds the single download url without touching axios', () => {
        expect(getFileDownloadUrl(99)).toBe('/api/v1/files/download/99');
        expect(mockedApi.get).not.toHaveBeenCalled();
    });

    it('builds the zip download url from the selected ids', () => {
        expect(getFilesZipDownloadUrl([1, 2, 3])).toBe('/api/v1/files/download-zip?ids=1,2,3');
        expect(mockedApi.get).not.toHaveBeenCalled();
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

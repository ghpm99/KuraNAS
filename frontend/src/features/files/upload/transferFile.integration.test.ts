import { apiBase } from '@/service';
import type { FileTransferParams } from './fileTransferTypes';
import { chunkedUploadThresholdBytes, transferFile } from './transferFile';

jest.mock('@/service', () => ({
    apiBase: {
        post: jest.fn(),
        get: jest.fn(),
        delete: jest.fn(),
    },
}));

const mockedApi = apiBase as unknown as { post: jest.Mock; get: jest.Mock; delete: jest.Mock };

const largeFileSize = chunkedUploadThresholdBytes + 1;
const serverChunkSize = 40 * 1024 * 1024;
const knownChecksum = 'c'.repeat(64);
const modifiedAt = Date.UTC(2026, 4, 6, 7, 8, 9);

const buildLargeFile = (): File => {
    const file = new File(['content'], 'film.mkv', { lastModified: modifiedAt });
    Object.defineProperty(file, 'size', { value: largeFileSize });
    return file;
};

const buildParams = (file: File): FileTransferParams => ({
    file,
    targetFolderId: 12,
    relativePath: 'Filmes/film.mkv',
    onConflict: 'replace',
    signal: new AbortController().signal,
    knownChecksum,
    onPhaseChange: jest.fn(),
    onChecksumComputed: jest.fn(),
    onSentBytes: jest.fn(),
});

describe('transferFile (seam)', () => {
    beforeEach(() => jest.clearAllMocks());

    it('runs init, chunk and complete against the real routes with the exact payloads', async () => {
        mockedApi.post.mockImplementation((url: string) => {
            switch (url) {
                case '/files/upload/init':
                    return Promise.resolve({
                        data: {
                            upload_id: 'upload-77',
                            chunk_size: serverChunkSize,
                            received_size: 0,
                            resumed: false,
                            already_exists: false,
                        },
                    });
                case '/files/upload/chunk':
                    return Promise.resolve({
                        data: {
                            received_size:
                                mockedApi.post.mock.calls.filter(
                                    ([calledUrl]) => calledUrl === '/files/upload/chunk'
                                ).length === 1
                                    ? serverChunkSize
                                    : largeFileSize,
                        },
                    });
                case '/files/upload/complete':
                    return Promise.resolve({
                        data: {
                            file: {
                                name: 'film.mkv',
                                path: '/Filmes/film.mkv',
                                status: 'replaced',
                            },
                            job_id: 5,
                        },
                    });
                default:
                    return Promise.reject(new Error(`unexpected route ${url}`));
            }
        });
        const params = buildParams(buildLargeFile());

        const outcome = await transferFile(params);

        expect(outcome).toEqual({ status: 'replaced', name: 'film.mkv' });
        expect(mockedApi.post.mock.calls.map(([url]) => url)).toEqual([
            '/files/upload/init',
            '/files/upload/chunk',
            '/files/upload/chunk',
            '/files/upload/complete',
        ]);
        expect(mockedApi.post).toHaveBeenNthCalledWith(
            1,
            '/files/upload/init',
            {
                file_name: 'film.mkv',
                size: largeFileSize,
                on_conflict: 'replace',
                target_folder_id: 12,
                relative_path: 'Filmes/film.mkv',
                checksum: knownChecksum,
                modified_at: '2026-05-06T07:08:09.000Z',
            },
            { signal: params.signal }
        );
        const chunkForms = [2, 3].map(
            (callNumber) => mockedApi.post.mock.calls[callNumber - 1]![1] as FormData
        );
        expect(chunkForms.map((form) => form.get('upload_id'))).toEqual(['upload-77', 'upload-77']);
        expect(chunkForms.map((form) => form.get('offset'))).toEqual([
            '0',
            String(serverChunkSize),
        ]);
        expect(chunkForms[0]!.get('chunk')).toBeInstanceOf(Blob);
        expect(mockedApi.post.mock.calls[1]![2]).toEqual(
            expect.objectContaining({
                headers: { 'Content-Type': 'multipart/form-data' },
                signal: params.signal,
            })
        );
        expect(mockedApi.post).toHaveBeenNthCalledWith(
            4,
            '/files/upload/complete',
            { upload_id: 'upload-77' },
            { signal: params.signal }
        );
        expect(params.onSentBytes).toHaveBeenLastCalledWith(largeFileSize);
    });

    it('sends a file at the threshold in the single request route', async () => {
        mockedApi.post.mockResolvedValue({
            data: { files: [{ name: 'a.txt', status: 'uploaded' }] },
        });
        const file = new File(['content'], 'a.txt');
        Object.defineProperty(file, 'size', { value: chunkedUploadThresholdBytes });

        await expect(transferFile(buildParams(file))).resolves.toEqual({
            status: 'uploaded',
            name: 'a.txt',
            error: undefined,
        });

        expect(mockedApi.post.mock.calls.map(([url]) => url)).toEqual(['/files/upload']);
    });
});

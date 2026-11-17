jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
        delete: jest.fn(),
    },
}));

import { AxiosError, AxiosHeaders } from 'axios';
import { apiBase } from '@/service';
import {
    cancelFileUpload,
    completeFileUpload,
    getFileUploadSession,
    initFileUpload,
    readMisalignedReceivedSize,
    sendFileUploadChunk,
} from './fileUpload';

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; delete: jest.Mock };

const buildAxiosError = (status: number | undefined, body?: unknown): AxiosError => {
    const config = { headers: new AxiosHeaders() };
    const response =
        status === undefined
            ? undefined
            : { status, statusText: '', headers: {}, config, data: body };
    return new AxiosError('failed', undefined, config, undefined, response);
};

describe('service/fileUpload', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('initFileUpload', () => {
        const startResponse = {
            upload_id: 'u1',
            chunk_size: 1024,
            received_size: 10,
            resumed: true,
            already_exists: false,
        };

        it('sends the full payload and converts the response', async () => {
            mockedApi.post.mockResolvedValue({ data: startResponse });
            const signal = new AbortController().signal;

            const session = await initFileUpload({
                fileName: 'a.bin',
                size: 100,
                targetFolderId: 7,
                relativePath: 'dir/a.bin',
                checksum: 'abc',
                modifiedAt: '2026-01-02T03:04:05.000Z',
                onConflict: 'rename',
                signal,
            });

            expect(mockedApi.post).toHaveBeenCalledWith(
                '/files/upload/init',
                {
                    file_name: 'a.bin',
                    size: 100,
                    on_conflict: 'rename',
                    target_folder_id: 7,
                    relative_path: 'dir/a.bin',
                    checksum: 'abc',
                    modified_at: '2026-01-02T03:04:05.000Z',
                },
                { signal }
            );
            expect(session).toEqual({
                uploadId: 'u1',
                chunkSize: 1024,
                receivedSize: 10,
                resumed: true,
                alreadyExists: false,
            });
        });

        it('omits empty optionals and non-positive target folder', async () => {
            mockedApi.post.mockResolvedValue({ data: startResponse });

            await initFileUpload({
                fileName: 'a.bin',
                size: 0,
                targetFolderId: 0,
                relativePath: '',
                checksum: '',
                onConflict: 'skip',
            });

            expect(mockedApi.post).toHaveBeenCalledWith(
                '/files/upload/init',
                { file_name: 'a.bin', size: 0, on_conflict: 'skip' },
                { signal: undefined }
            );
        });

        it('returns no upload id when the file already exists', async () => {
            mockedApi.post.mockResolvedValue({
                data: { chunk_size: 1, received_size: 0, resumed: false, already_exists: true },
            });

            const session = await initFileUpload({ fileName: 'a', size: 1, onConflict: 'skip' });

            expect(session.uploadId).toBeUndefined();
            expect(session.alreadyExists).toBe(true);
        });
    });

    describe('sendFileUploadChunk', () => {
        it('posts multipart form and reports progress', async () => {
            mockedApi.post.mockResolvedValue({ data: { received_size: 30 } });
            const onProgress = jest.fn();
            const signal = new AbortController().signal;
            const chunk = new Blob(['abc']);

            const outcome = await sendFileUploadChunk({
                uploadId: 'u1',
                chunk,
                offset: 27,
                signal,
                onProgress,
            });

            const [url, formData, config] = mockedApi.post.mock.calls[0];
            expect(url).toBe('/files/upload/chunk');
            expect(formData).toBeInstanceOf(FormData);
            expect((formData as FormData).get('upload_id')).toBe('u1');
            expect((formData as FormData).get('offset')).toBe('27');
            expect((formData as FormData).get('chunk')).toBeInstanceOf(Blob);
            expect(config.headers).toEqual({ 'Content-Type': 'multipart/form-data' });
            expect(config.signal).toBe(signal);
            config.onUploadProgress({ loaded: 2 });
            expect(onProgress).toHaveBeenCalledWith(2);
            expect(outcome).toEqual({ receivedSize: 30 });
        });

        it('tolerates absence of onProgress', async () => {
            mockedApi.post.mockResolvedValue({ data: { received_size: 3 } });

            await sendFileUploadChunk({ uploadId: 'u1', chunk: new Blob(['abc']), offset: 0 });

            const config = mockedApi.post.mock.calls[0][2];
            expect(() => config.onUploadProgress({ loaded: 1 })).not.toThrow();
        });
    });

    it('gets the upload session', async () => {
        mockedApi.get.mockResolvedValue({
            data: { upload_id: 'u1', size: 100, received_size: 40 },
        });

        const session = await getFileUploadSession('u1');

        expect(mockedApi.get).toHaveBeenCalledWith('/files/upload/u1', { signal: undefined });
        expect(session).toEqual({ uploadId: 'u1', size: 100, receivedSize: 40 });
    });

    it('completes the upload', async () => {
        mockedApi.post.mockResolvedValue({
            data: { file: { name: 'a', path: '/a', status: 'renamed' }, job_id: 9 },
        });

        const completed = await completeFileUpload('u1');

        expect(mockedApi.post).toHaveBeenCalledWith(
            '/files/upload/complete',
            { upload_id: 'u1' },
            { signal: undefined }
        );
        expect(completed).toEqual({ file: { name: 'a', path: '/a', status: 'renamed' }, jobId: 9 });
    });

    it('cancels the upload', async () => {
        mockedApi.delete.mockResolvedValue({ status: 204 });

        await expect(cancelFileUpload('u1')).resolves.toBeUndefined();

        expect(mockedApi.delete).toHaveBeenCalledWith('/files/upload/u1');
    });

    describe('readMisalignedReceivedSize', () => {
        it('reads received_size from a 409', () => {
            expect(
                readMisalignedReceivedSize(buildAxiosError(409, { error: 'x', received_size: 12 }))
            ).toBe(12);
        });

        it('returns undefined for a 409 without received_size', () => {
            expect(
                readMisalignedReceivedSize(buildAxiosError(409, { error: 'x' }))
            ).toBeUndefined();
            expect(readMisalignedReceivedSize(buildAxiosError(409))).toBeUndefined();
        });

        it('returns undefined for other statuses', () => {
            expect(
                readMisalignedReceivedSize(buildAxiosError(500, { received_size: 5 }))
            ).toBeUndefined();
        });

        it('returns undefined for network errors and non-axios errors', () => {
            expect(readMisalignedReceivedSize(buildAxiosError(undefined))).toBeUndefined();
            expect(readMisalignedReceivedSize(new Error('boom'))).toBeUndefined();
            expect(readMisalignedReceivedSize(undefined)).toBeUndefined();
        });
    });
});

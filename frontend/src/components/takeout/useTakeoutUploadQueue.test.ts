import { act, renderHook } from '@testing-library/react';
import { apiBase } from '@/service';
import useTakeoutUploadQueue from './useTakeoutUploadQueue';

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { post: jest.Mock };

const zipFile = (name: string, content = 'hello') =>
    new File([content], name, { type: 'application/zip' });

const postedUrls = (): string[] => mockedApi.post.mock.calls.map(([url]) => url as string);

const respondToUploadEndpoints = (chunkSize = 10 * 1024 * 1024) => {
    let uploadCount = 0;
    mockedApi.post.mockImplementation((url: string) => {
        if (url === '/takeout/upload/init') {
            uploadCount += 1;
            return Promise.resolve({
                data: { upload_id: `u${uploadCount}`, chunk_size: chunkSize },
            });
        }
        if (url === '/takeout/upload/chunk') return Promise.resolve({ data: { received: true } });
        if (url === '/takeout/upload/complete') {
            return Promise.resolve({ data: { job_id: 42, import_id: 7, message: 'ok' } });
        }
        return Promise.reject(new Error(`unexpected POST ${url}`));
    });
};

describe('components/takeout/useTakeoutUploadQueue', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        respondToUploadEndpoints();
    });

    it('runs init -> chunk -> complete against the takeout endpoints and keeps the import id', async () => {
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(zipFile('takeout.zip')));

        await act(async () => {
            await result.current.startQueue();
        });

        expect(mockedApi.post).toHaveBeenCalledWith('/takeout/upload/init', {
            file_name: 'takeout.zip',
            size: 5,
        });
        expect(mockedApi.post).toHaveBeenCalledWith(
            '/takeout/upload/chunk',
            expect.any(FormData),
            expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } })
        );
        expect(mockedApi.post).toHaveBeenCalledWith('/takeout/upload/complete', {
            upload_id: 'u1',
        });
        expect(result.current.entries[0]).toMatchObject({
            state: 'done',
            importId: 7,
            uploadedBytes: 5,
        });
    });

    it('sends the file in several chunks', async () => {
        respondToUploadEndpoints(2);
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(zipFile('big.zip')));

        await act(async () => {
            await result.current.startQueue();
        });

        expect(postedUrls().filter((url) => url === '/takeout/upload/chunk')).toHaveLength(3);
        expect(result.current.entries[0]?.state).toBe('done');
    });

    it('sends two zips in sequence, starting the second only after the first completes', async () => {
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => {
            result.current.addFile(zipFile('first.zip'));
            result.current.addFile(zipFile('second.zip'));
        });

        await act(async () => {
            await result.current.startQueue();
        });

        expect(postedUrls()).toEqual([
            '/takeout/upload/init',
            '/takeout/upload/chunk',
            '/takeout/upload/complete',
            '/takeout/upload/init',
            '/takeout/upload/chunk',
            '/takeout/upload/complete',
        ]);
        expect(result.current.entries.map((entry) => entry.state)).toEqual(['done', 'done']);
    });

    it('processes a file added while the queue is running in the same run', async () => {
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(zipFile('first.zip')));

        await act(async () => {
            const running = result.current.startQueue();
            result.current.addFile(zipFile('second.zip'));
            await running;
        });

        expect(result.current.entries.map((entry) => entry.state)).toEqual(['done', 'done']);
    });

    it('rejects a non zip file without calling the API', async () => {
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(new File(['x'], 'backup.rar')));

        expect(result.current.entries[0]).toMatchObject({
            state: 'rejected',
            errorMessage: 'TAKEOUT_FILE_TYPE_ERROR',
        });
        expect(result.current.hasQueuedEntries).toBe(false);

        await act(async () => {
            await result.current.startQueue();
        });

        expect(mockedApi.post).not.toHaveBeenCalled();
    });

    it('marks only the failed entry with the server message and goes on to the next', async () => {
        let initCount = 0;
        mockedApi.post.mockImplementation((url: string) => {
            if (url === '/takeout/upload/init') {
                initCount += 1;
                if (initCount === 1) {
                    return Promise.reject({
                        response: { data: { error: 'O arquivo deve terminar em .zip' } },
                    });
                }
                return Promise.resolve({ data: { upload_id: 'u2', chunk_size: 1024 } });
            }
            if (url === '/takeout/upload/chunk')
                return Promise.resolve({ data: { received: true } });
            return Promise.resolve({ data: { job_id: 1, message: 'ok' } });
        });
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => {
            result.current.addFile(zipFile('first.zip'));
            result.current.addFile(zipFile('second.zip'));
        });

        await act(async () => {
            await result.current.startQueue();
        });

        expect(result.current.entries[0]).toMatchObject({
            state: 'error',
            errorMessage: 'O arquivo deve terminar em .zip',
        });
        expect(result.current.entries[1]?.state).toBe('done');
    });

    it('falls back to the upload error key when there is no server response', async () => {
        mockedApi.post.mockRejectedValue(new Error('network'));
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(zipFile('takeout.zip')));

        await act(async () => {
            await result.current.startQueue();
        });

        expect(result.current.entries[0]).toMatchObject({
            state: 'error',
            errorMessage: 'TAKEOUT_UPLOAD_ERROR',
        });
    });

    it('clears settled entries', async () => {
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => {
            result.current.addFile(zipFile('takeout.zip'));
            result.current.addFile(new File(['x'], 'a.rar'));
        });
        await act(async () => {
            await result.current.startQueue();
        });

        act(() => result.current.clearSettled());

        expect(result.current.entries).toEqual([]);
        expect(result.current.isBusy).toBe(false);
    });
});

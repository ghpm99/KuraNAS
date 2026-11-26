import { act, renderHook } from '@testing-library/react';
import { AxiosError } from 'axios';
import { apiBase } from '@/service';
import useTakeoutUploadQueue from './useTakeoutUploadQueue';

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn(), get: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { post: jest.Mock; get: jest.Mock; delete: jest.Mock };

const zipFile = (name: string) => new File(['0123456789'], name, { type: 'application/zip' });

const chunkPostCount = (): number =>
    mockedApi.post.mock.calls.filter(([url]) => url === '/takeout/upload/chunk').length;

const uploadedFileNames = (): string[] =>
    mockedApi.post.mock.calls
        .filter(([url]) => url === '/takeout/upload/init')
        .map(([, body]) => (body as { file_name: string }).file_name);

const respondToInitAndComplete = (url: string) => {
    if (url === '/takeout/upload/init') {
        return Promise.resolve({
            data: { upload_id: 'u1', chunk_size: 4, received_size: 0, resumed: false },
        });
    }
    return Promise.resolve({ data: { job_id: 1, import_id: 7, message: 'ok' } });
};

describe('components/takeout/useTakeoutUploadQueue cancel', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.delete.mockResolvedValue({});
    });

    it('cancels an uploading zip, stops sending parts and goes on to the next zip', async () => {
        let releaseFirstChunk: () => void = () => undefined;
        let chunkCalls = 0;
        mockedApi.post.mockImplementation((url: string, body: unknown) => {
            if (url !== '/takeout/upload/chunk') return respondToInitAndComplete(url);
            chunkCalls += 1;
            if (chunkCalls > 1) return Promise.resolve({ data: { received: true } });
            return new Promise((resolve, reject) => {
                releaseFirstChunk = () => reject(new AxiosError('aborted', 'ERR_CANCELED'));
                void body;
                void resolve;
            });
        });
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => {
            result.current.addFile(zipFile('first.zip'));
            result.current.addFile(zipFile('second.zip'));
        });

        let queueRun: Promise<void> = Promise.resolve();
        await act(async () => {
            queueRun = result.current.startQueue();
            await Promise.resolve();
        });
        const chunksBeforeCancel = chunkPostCount();

        await act(async () => {
            const canceling = result.current.cancelEntry('takeout-upload-1');
            releaseFirstChunk();
            await canceling;
            await queueRun;
        });

        expect(mockedApi.delete).toHaveBeenCalledWith('/takeout/upload/u1');
        expect(chunksBeforeCancel).toBe(1);
        expect(result.current.entries[0]?.state).toBe('canceled');
        expect(result.current.entries[1]?.state).toBe('done');
        expect(uploadedFileNames()).toEqual(['first.zip', 'second.zip']);
    });

    it('cancels a queued zip without calling the API and never sends it', async () => {
        mockedApi.post.mockImplementation((url: string) => respondToInitAndComplete(url));
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(zipFile('queued.zip')));

        await act(async () => {
            await result.current.cancelEntry('takeout-upload-1');
        });
        await act(async () => {
            await result.current.startQueue();
        });

        expect(result.current.entries[0]?.state).toBe('canceled');
        expect(mockedApi.post).not.toHaveBeenCalled();
        expect(mockedApi.delete).not.toHaveBeenCalled();
    });

    it('aborts the reconnection wait and deletes the upload', async () => {
        jest.useFakeTimers();
        try {
            mockedApi.post.mockImplementation((url: string) =>
                url === '/takeout/upload/chunk'
                    ? Promise.reject(new AxiosError('Network Error', 'ERR_NETWORK'))
                    : respondToInitAndComplete(url)
            );
            const { result } = renderHook(() => useTakeoutUploadQueue());
            act(() => result.current.addFile(zipFile('takeout.zip')));

            let queueRun: Promise<void> = Promise.resolve();
            await act(async () => {
                queueRun = result.current.startQueue();
                await jest.advanceTimersByTimeAsync(0);
            });
            expect(result.current.entries[0]?.state).toBe('waitingConnection');

            await act(async () => {
                await result.current.cancelEntry('takeout-upload-1');
                await queueRun;
            });

            expect(mockedApi.delete).toHaveBeenCalledWith('/takeout/upload/u1');
            expect(mockedApi.get).not.toHaveBeenCalled();
            expect(result.current.entries[0]?.state).toBe('canceled');
        } finally {
            jest.useRealTimers();
        }
    });

    it('keeps the entry canceled and shows the fallback message when the delete fails', async () => {
        mockedApi.delete.mockRejectedValue(new AxiosError('Network Error', 'ERR_NETWORK'));
        let releaseFirstChunk: () => void = () => undefined;
        mockedApi.post.mockImplementation((url: string) =>
            url === '/takeout/upload/chunk'
                ? new Promise((_resolve, reject) => {
                      releaseFirstChunk = () => reject(new AxiosError('aborted', 'ERR_CANCELED'));
                  })
                : respondToInitAndComplete(url)
        );
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(zipFile('takeout.zip')));
        let queueRun: Promise<void> = Promise.resolve();
        await act(async () => {
            queueRun = result.current.startQueue();
            await Promise.resolve();
        });

        await act(async () => {
            const canceling = result.current.cancelEntry('takeout-upload-1');
            releaseFirstChunk();
            await canceling;
            await queueRun;
        });

        expect(result.current.entries[0]?.state).toBe('canceled');
        expect(result.current.entries[0]?.errorMessage).toBe('TAKEOUT_CANCEL_ERROR');
    });
});

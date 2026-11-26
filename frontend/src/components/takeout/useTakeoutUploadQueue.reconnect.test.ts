import { act, renderHook } from '@testing-library/react';
import { AxiosError } from 'axios';
import { apiBase } from '@/service';
import useTakeoutUploadQueue from './useTakeoutUploadQueue';

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn(), get: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { post: jest.Mock; get: jest.Mock };

const networkError = () => new AxiosError('Network Error', 'ERR_NETWORK');

const zipFile = (name: string) => new File(['hello'], name, { type: 'application/zip' });

const answerInitAndComplete = (url: string) => {
    if (url === '/takeout/upload/init') {
        return Promise.resolve({
            data: { upload_id: 'u1', chunk_size: 100, received_size: 0, resumed: false },
        });
    }
    return Promise.resolve({ data: { job_id: 1, import_id: 7, message: 'ok' } });
};

describe('components/takeout/useTakeoutUploadQueue reconnection', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        jest.useFakeTimers();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('shows waitingConnection during the wait and ends done after reconnecting', async () => {
        let chunkAttempts = 0;
        mockedApi.post.mockImplementation((url: string) => {
            if (url !== '/takeout/upload/chunk') return answerInitAndComplete(url);
            chunkAttempts += 1;
            return chunkAttempts === 1
                ? Promise.reject(networkError())
                : Promise.resolve({ data: { received: true } });
        });
        mockedApi.get.mockResolvedValue({ data: { upload_id: 'u1', size: 5, received_size: 0 } });
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(zipFile('takeout.zip')));

        let queueRun: Promise<void> = Promise.resolve();
        await act(async () => {
            queueRun = result.current.startQueue();
            await jest.advanceTimersByTimeAsync(0);
        });
        expect(result.current.entries[0]?.state).toBe('waitingConnection');
        expect(result.current.isBusy).toBe(true);

        await act(async () => {
            await jest.advanceTimersByTimeAsync(1000);
            await queueRun;
        });

        expect(result.current.entries[0]?.state).toBe('done');
        expect(result.current.entries[0]?.importId).toBe(7);
    });

    it('reports the lost connection after the limit and goes on to the next zip', async () => {
        let firstZipInitialized = false;
        mockedApi.post.mockImplementation((url: string, body: unknown) => {
            if (url === '/takeout/upload/init') {
                const isFirstZip = (body as { file_name: string }).file_name === 'first.zip';
                firstZipInitialized ||= isFirstZip;
                return Promise.resolve({
                    data: {
                        upload_id: isFirstZip ? 'u1' : 'u2',
                        chunk_size: 100,
                        received_size: 0,
                        resumed: false,
                    },
                });
            }
            if (url === '/takeout/upload/chunk') {
                const isFirstZip = (body as FormData).get('upload_id') === 'u1';
                return isFirstZip
                    ? Promise.reject(networkError())
                    : Promise.resolve({ data: { received: true } });
            }
            return Promise.resolve({ data: { job_id: 2, import_id: 8, message: 'ok' } });
        });
        mockedApi.get.mockRejectedValue(networkError());
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => {
            result.current.addFile(zipFile('first.zip'));
            result.current.addFile(zipFile('second.zip'));
        });

        await act(async () => {
            const queueRun = result.current.startQueue();
            await jest.advanceTimersByTimeAsync(11 * 60 * 1000);
            await queueRun;
        });

        expect(firstZipInitialized).toBe(true);
        expect(result.current.entries[0]).toMatchObject({
            state: 'error',
            errorMessage: 'TAKEOUT_UPLOAD_CONNECTION_LOST',
        });
        expect(result.current.entries[1]?.state).toBe('done');
    });
});

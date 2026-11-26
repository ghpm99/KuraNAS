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

const tenByteZip = () => new File(['0123456789'], 'takeout.zip', { type: 'application/zip' });

const answerInit = () =>
    Promise.resolve({
        data: { upload_id: 'u1', chunk_size: 2, received_size: 0, resumed: false },
    });

const answerComplete = () => Promise.resolve({ data: { job_id: 1, import_id: 7, message: 'ok' } });

const oneSecondLater = () => jest.setSystemTime(Date.now() + 1000);

describe('components/takeout/useTakeoutUploadQueue transfer rate', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        jest.useFakeTimers();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('computes speed and remaining time from parts sent at distinct instants', async () => {
        mockedApi.post.mockImplementation((url: string) => {
            if (url === '/takeout/upload/init') return answerInit();
            if (url !== '/takeout/upload/chunk') return answerComplete();
            oneSecondLater();
            return Promise.resolve({ data: { received: true } });
        });
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(tenByteZip()));

        await act(async () => {
            await result.current.startQueue();
        });

        expect(result.current.entries[0]?.bytesPerSecond).toBeCloseTo(2);
        expect(result.current.entries[0]?.remainingSeconds).toBe(0);
    });

    it('clears speed and remaining time while waiting for the connection', async () => {
        let chunkAttempts = 0;
        mockedApi.post.mockImplementation((url: string) => {
            if (url === '/takeout/upload/init') return answerInit();
            if (url !== '/takeout/upload/chunk') return answerComplete();
            chunkAttempts += 1;
            if (chunkAttempts === 4) return Promise.reject(new AxiosError('x', 'ERR_NETWORK'));
            oneSecondLater();
            return Promise.resolve({ data: { received: true } });
        });
        mockedApi.get.mockResolvedValue({ data: { upload_id: 'u1', size: 10, received_size: 6 } });
        const { result } = renderHook(() => useTakeoutUploadQueue());
        act(() => result.current.addFile(tenByteZip()));

        await act(async () => {
            void result.current.startQueue();
            await jest.advanceTimersByTimeAsync(0);
        });

        expect(result.current.entries[0]?.state).toBe('waitingConnection');
        expect(result.current.entries[0]?.bytesPerSecond).toBeUndefined();
        expect(result.current.entries[0]?.remainingSeconds).toBeUndefined();

        await act(async () => {
            await jest.advanceTimersByTimeAsync(10000);
        });
    });
});

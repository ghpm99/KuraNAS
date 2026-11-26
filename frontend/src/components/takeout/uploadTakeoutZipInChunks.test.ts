import { AxiosError, AxiosHeaders } from 'axios';
import { apiBase } from '@/service';
import { TakeoutConnectionLostError } from './takeoutConnectionLostError';
import {
    uploadTakeoutZipInChunks,
    type TakeoutUploadCallbacks,
    type TakeoutUploadDependencies,
} from './uploadTakeoutZipInChunks';

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn(), get: jest.fn() },
}));

const mockedApi = apiBase as unknown as { post: jest.Mock; get: jest.Mock };

const networkError = () => new AxiosError('Network Error', 'ERR_NETWORK');

const serverError = (status: number) =>
    new AxiosError('failed', 'ERR_BAD_REQUEST', undefined, undefined, {
        status,
        statusText: '',
        data: { error: 'server says no' },
        headers: {},
        config: { headers: new AxiosHeaders() },
    });

const zipOfTenBytes = () => new File(['0123456789'], 'takeout.zip');

const initResponse = (received_size = 0, resumed = false) => ({
    data: { upload_id: 'u1', chunk_size: 4, received_size, resumed },
});

const completeResponse = { data: { job_id: 1, import_id: 7, message: 'ok' } };

const chunkOffsets = (): number[] =>
    mockedApi.post.mock.calls
        .filter(([url]) => url === '/takeout/upload/chunk')
        .map(([, body]) => Number((body as FormData).get('offset')));

const buildHarness = (connectionLossLimitMs = 10 * 60 * 1000) => {
    let currentTime = 0;
    const sleeps: number[] = [];
    const dependencies: TakeoutUploadDependencies = {
        sleep: (milliseconds) => {
            sleeps.push(milliseconds);
            currentTime += milliseconds;
            return Promise.resolve();
        },
        now: () => currentTime,
        initialRetryDelayMs: 1000,
        maxRetryDelayMs: 4000,
        connectionLossLimitMs,
    };
    const callbacks: TakeoutUploadCallbacks = {
        onProgress: jest.fn(),
        onWaitingConnection: jest.fn(),
        onCompleting: jest.fn(),
        onSessionStarted: jest.fn(),
    };
    return { dependencies, callbacks, sleeps };
};

const scriptPosts = (init: unknown, chunkOutcomes: unknown[]) => {
    const outcomes = [...chunkOutcomes];
    mockedApi.post.mockImplementation((url: string) => {
        if (url === '/takeout/upload/init') return Promise.resolve(init);
        if (url === '/takeout/upload/complete') return Promise.resolve(completeResponse);
        const outcome = outcomes.shift();
        return outcome instanceof Error ? Promise.reject(outcome) : Promise.resolve({ data: {} });
    });
};

describe('components/takeout/uploadTakeoutZipInChunks', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('sends only the missing part when the server resumed the session', async () => {
        const { dependencies, callbacks } = buildHarness();
        scriptPosts(initResponse(8, true), []);

        const result = await uploadTakeoutZipInChunks(
            zipOfTenBytes(),
            callbacks,
            new AbortController().signal,
            dependencies
        );

        expect(chunkOffsets()).toEqual([8]);
        expect(callbacks.onProgress).toHaveBeenNthCalledWith(1, 8);
        expect(result.import_id).toBe(7);
    });

    it('waits, reads the status and continues from the received size after a network drop', async () => {
        const { dependencies, callbacks, sleeps } = buildHarness();
        scriptPosts(initResponse(), [{}, networkError()]);
        mockedApi.get.mockResolvedValue({ data: { upload_id: 'u1', size: 10, received_size: 4 } });

        await uploadTakeoutZipInChunks(
            zipOfTenBytes(),
            callbacks,
            new AbortController().signal,
            dependencies
        );

        expect(sleeps).toEqual([1000]);
        expect(mockedApi.get).toHaveBeenCalledWith('/takeout/upload/u1', {
            signal: expect.any(AbortSignal),
        });
        expect(chunkOffsets()).toEqual([0, 4, 4, 8]);
        expect(callbacks.onWaitingConnection).toHaveBeenNthCalledWith(1, true);
        expect(callbacks.onWaitingConnection).toHaveBeenLastCalledWith(false);
        expect(mockedApi.post).toHaveBeenLastCalledWith(
            '/takeout/upload/complete',
            { upload_id: 'u1' },
            { signal: expect.any(AbortSignal) }
        );
    });

    it('doubles the wait on consecutive drops up to the maximum', async () => {
        const { dependencies, callbacks, sleeps } = buildHarness();
        scriptPosts(initResponse(), [networkError()]);
        mockedApi.get
            .mockRejectedValueOnce(networkError())
            .mockRejectedValueOnce(networkError())
            .mockRejectedValueOnce(networkError())
            .mockResolvedValue({ data: { upload_id: 'u1', size: 10, received_size: 0 } });

        await uploadTakeoutZipInChunks(
            zipOfTenBytes(),
            callbacks,
            new AbortController().signal,
            dependencies
        );

        expect(sleeps).toEqual([1000, 2000, 4000, 4000]);
    });

    it('throws TakeoutConnectionLostError once the loss limit passes without reconnecting', async () => {
        const { dependencies, callbacks } = buildHarness(5000);
        scriptPosts(initResponse(), [networkError()]);
        mockedApi.get.mockRejectedValue(networkError());

        await expect(
            uploadTakeoutZipInChunks(
                zipOfTenBytes(),
                callbacks,
                new AbortController().signal,
                dependencies
            )
        ).rejects.toBeInstanceOf(TakeoutConnectionLostError);
    });

    it('does not retry a server error', async () => {
        const { dependencies, callbacks, sleeps } = buildHarness();
        const failure = serverError(500);
        scriptPosts(initResponse(), [failure]);

        await expect(
            uploadTakeoutZipInChunks(
                zipOfTenBytes(),
                callbacks,
                new AbortController().signal,
                dependencies
            )
        ).rejects.toBe(failure);

        expect(sleeps).toEqual([]);
        expect(mockedApi.get).not.toHaveBeenCalled();
        expect(chunkOffsets()).toEqual([0]);
    });

    it('realigns by the received size when the server rejects the offset', async () => {
        const { dependencies, callbacks, sleeps } = buildHarness();
        scriptPosts(initResponse(), [serverError(400)]);
        mockedApi.get.mockResolvedValue({ data: { upload_id: 'u1', size: 10, received_size: 4 } });

        await uploadTakeoutZipInChunks(
            zipOfTenBytes(),
            callbacks,
            new AbortController().signal,
            dependencies
        );

        expect(sleeps).toEqual([]);
        expect(chunkOffsets()).toEqual([0, 4, 8]);
    });

    it('rethrows a 400 that is not an offset mismatch', async () => {
        const { dependencies, callbacks } = buildHarness();
        const failure = serverError(400);
        scriptPosts(initResponse(), [failure]);
        mockedApi.get.mockResolvedValue({ data: { upload_id: 'u1', size: 10, received_size: 0 } });

        await expect(
            uploadTakeoutZipInChunks(
                zipOfTenBytes(),
                callbacks,
                new AbortController().signal,
                dependencies
            )
        ).rejects.toBe(failure);
    });

    it('stops without completing when aborted in the middle and reports the session', async () => {
        const { dependencies, callbacks } = buildHarness();
        const abortController = new AbortController();
        const abortReason = new Error('canceled by user');
        let chunkCalls = 0;
        mockedApi.post.mockImplementation((url: string) => {
            if (url === '/takeout/upload/init') return Promise.resolve(initResponse());
            chunkCalls += 1;
            if (chunkCalls === 2) {
                abortController.abort(abortReason);
                return Promise.reject(networkError());
            }
            return Promise.resolve({ data: {} });
        });

        await expect(
            uploadTakeoutZipInChunks(
                zipOfTenBytes(),
                callbacks,
                abortController.signal,
                dependencies
            )
        ).rejects.toBe(abortReason);

        expect(callbacks.onSessionStarted).toHaveBeenCalledWith('u1');
        expect(callbacks.onCompleting).not.toHaveBeenCalled();
        expect(mockedApi.get).not.toHaveBeenCalled();
        expect(mockedApi.post.mock.calls.map(([url]) => url)).not.toContain(
            '/takeout/upload/complete'
        );
    });
});

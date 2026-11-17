import { AxiosError, AxiosHeaders, CanceledError } from 'axios';
import {
    cancelFileUpload,
    completeFileUpload,
    getFileUploadSession,
    initFileUpload,
    readMisalignedReceivedSize,
    sendFileUploadChunk,
} from '@/service/fileUpload';
import { computeFileSha256 } from './computeFileSha256';
import type { FileTransferParams } from './fileTransferTypes';
import { UploadConnectionLostError } from './uploadConnectionLostError';
import { uploadFileInChunks } from './uploadFileInChunks';

jest.mock('@/service/fileUpload', () => ({
    initFileUpload: jest.fn(),
    sendFileUploadChunk: jest.fn(),
    getFileUploadSession: jest.fn(),
    completeFileUpload: jest.fn(),
    cancelFileUpload: jest.fn(),
    readMisalignedReceivedSize: jest.fn(),
}));
jest.mock('./computeFileSha256', () => ({ computeFileSha256: jest.fn() }));

const checksum = 'a'.repeat(64);
const lastModified = Date.UTC(2026, 0, 2, 3, 4, 5);
const file = new File(['0123456789'], 'big.bin', { lastModified });
const chunkSize = 4;

const networkError = () => new AxiosError('Network Error', 'ERR_NETWORK');

const httpError = (status: number, data: unknown = {}) =>
    new AxiosError('failed', 'ERR_BAD_REQUEST', undefined, undefined, {
        status,
        statusText: '',
        data,
        headers: {},
        config: { headers: new AxiosHeaders() },
    });

const buildClock = () => {
    let currentTime = 0;
    const sleep = jest.fn(async (milliseconds: number) => {
        currentTime += milliseconds;
    });
    return { now: () => currentTime, sleep };
};

type Harness = ReturnType<typeof buildHarness>;

const buildHarness = (overrides: Partial<FileTransferParams> = {}) => {
    const controller = new AbortController();
    const clock = buildClock();
    const params: FileTransferParams = {
        file,
        targetFolderId: 3,
        relativePath: 'Docs/big.bin',
        onConflict: 'rename',
        signal: controller.signal,
        onPhaseChange: jest.fn(),
        onChecksumComputed: jest.fn(),
        onSentBytes: jest.fn(),
        ...overrides,
    };
    const run = () =>
        uploadFileInChunks(params, {
            sleep: clock.sleep,
            now: clock.now,
            reconnectProbeIntervalMs: 5000,
            connectionLossLimitMs: 600_000,
        });
    return { controller, clock, params, run };
};

const sentBytesCalls = (harness: Harness): number[] =>
    jest.mocked(harness.params.onSentBytes).mock.calls.map(([sentBytes]) => sentBytes);

const phaseCalls = (harness: Harness): string[] =>
    jest.mocked(harness.params.onPhaseChange).mock.calls.map(([phase]) => phase);

const startedSession = (receivedSize = 0, resumed = false) => ({
    uploadId: 'upload-1',
    chunkSize,
    receivedSize,
    resumed,
    alreadyExists: false,
});

const completedFile = (status: 'uploaded' | 'renamed' = 'uploaded') => ({
    file: { name: 'big.bin', path: '/Docs/big.bin', status },
    jobId: 9,
});

describe('uploadFileInChunks', () => {
    beforeEach(() => {
        jest.resetAllMocks();
        jest.mocked(computeFileSha256).mockResolvedValue(checksum);
        jest.mocked(cancelFileUpload).mockResolvedValue();
        jest.mocked(readMisalignedReceivedSize).mockReturnValue(undefined);
        jest.mocked(initFileUpload).mockResolvedValue(startedSession());
        jest.mocked(completeFileUpload).mockResolvedValue(completedFile());
    });

    it('checks, inits with the exact payload, sends every part at its offset and completes', async () => {
        const harness = buildHarness();
        jest.mocked(sendFileUploadChunk)
            .mockImplementationOnce(async ({ onProgress }) => {
                onProgress?.(2);
                return { receivedSize: 4 };
            })
            .mockResolvedValueOnce({ receivedSize: 8 })
            .mockResolvedValueOnce({ receivedSize: 10 });

        const outcome = await harness.run();

        expect(outcome).toEqual({ status: 'uploaded', name: 'big.bin' });
        expect(phaseCalls(harness)).toEqual(['checking', 'uploading']);
        expect(computeFileSha256).toHaveBeenCalledWith(file, { signal: harness.params.signal });
        expect(harness.params.onChecksumComputed).toHaveBeenCalledWith(checksum);
        expect(initFileUpload).toHaveBeenCalledWith({
            fileName: 'big.bin',
            size: 10,
            targetFolderId: 3,
            relativePath: 'Docs/big.bin',
            checksum,
            modifiedAt: '2026-01-02T03:04:05.000Z',
            onConflict: 'rename',
            signal: harness.params.signal,
        });
        const sentChunks = jest.mocked(sendFileUploadChunk).mock.calls.map(([chunkParams]) => ({
            uploadId: chunkParams.uploadId,
            offset: chunkParams.offset,
            size: chunkParams.chunk.size,
        }));
        expect(sentChunks).toEqual([
            { uploadId: 'upload-1', offset: 0, size: 4 },
            { uploadId: 'upload-1', offset: 4, size: 4 },
            { uploadId: 'upload-1', offset: 8, size: 2 },
        ]);
        expect(sentBytesCalls(harness)).toEqual([0, 2, 4, 8, 10]);
        expect(completeFileUpload).toHaveBeenCalledWith('upload-1', harness.params.signal);
    });

    it('returns the renamed status reported by complete', async () => {
        const harness = buildHarness();
        jest.mocked(initFileUpload).mockResolvedValue(startedSession(10));
        jest.mocked(completeFileUpload).mockResolvedValue(completedFile('renamed'));

        await expect(harness.run()).resolves.toEqual({ status: 'renamed', name: 'big.bin' });
        expect(sendFileUploadChunk).not.toHaveBeenCalled();
    });

    it('does not send any part when the server already has the file', async () => {
        const harness = buildHarness();
        jest.mocked(initFileUpload).mockResolvedValue({
            chunkSize,
            receivedSize: 0,
            resumed: false,
            alreadyExists: true,
        });

        await expect(harness.run()).resolves.toEqual({ status: 'alreadyExisted' });

        expect(sendFileUploadChunk).not.toHaveBeenCalled();
        expect(completeFileUpload).not.toHaveBeenCalled();
    });

    it.each([409, 507])(
        'propagates the init HTTP error %s without sending bytes',
        async (status) => {
            const harness = buildHarness();
            const initError = httpError(status, { error: 'mensagem do servidor' });
            jest.mocked(initFileUpload).mockRejectedValue(initError);

            await expect(harness.run()).rejects.toBe(initError);

            expect(sendFileUploadChunk).not.toHaveBeenCalled();
            expect(sentBytesCalls(harness)).toEqual([]);
        }
    );

    it('fails when init answers without an upload id and without an existing file', async () => {
        const harness = buildHarness();
        jest.mocked(initFileUpload).mockResolvedValue({
            chunkSize,
            receivedSize: 0,
            resumed: false,
            alreadyExists: false,
        });

        await expect(harness.run()).rejects.toThrow('without an upload id');
        expect(sendFileUploadChunk).not.toHaveBeenCalled();
    });

    it('resumes from the received size of a resumed session', async () => {
        const harness = buildHarness();
        jest.mocked(initFileUpload).mockResolvedValue(startedSession(8, true));
        jest.mocked(sendFileUploadChunk).mockResolvedValueOnce({ receivedSize: 10 });

        await harness.run();

        const [firstChunk] = jest.mocked(sendFileUploadChunk).mock.calls[0]!;
        expect(firstChunk.offset).toBe(8);
        expect(firstChunk.chunk.size).toBe(2);
        expect(sentBytesCalls(harness)).toEqual([8, 10]);
    });

    it('realigns to the received size reported by a 409 and keeps sending', async () => {
        const harness = buildHarness();
        const misaligned = httpError(409, { received_size: 6 });
        jest.mocked(readMisalignedReceivedSize).mockImplementation((error) =>
            error === misaligned ? 6 : undefined
        );
        jest.mocked(sendFileUploadChunk)
            .mockResolvedValueOnce({ receivedSize: 4 })
            .mockRejectedValueOnce(misaligned)
            .mockResolvedValueOnce({ receivedSize: 10 });

        await harness.run();

        const offsets = jest
            .mocked(sendFileUploadChunk)
            .mock.calls.map(([chunkParams]) => chunkParams.offset);
        expect(offsets).toEqual([0, 4, 6]);
        expect(sentBytesCalls(harness)).toEqual([0, 4, 6, 10]);
    });

    it('propagates a chunk HTTP error that is not a misalignment', async () => {
        const harness = buildHarness();
        const chunkError = httpError(500);
        jest.mocked(sendFileUploadChunk).mockRejectedValue(chunkError);

        await expect(harness.run()).rejects.toBe(chunkError);
        expect(getFileUploadSession).not.toHaveBeenCalled();
    });

    it('waits for the connection, asks the session and continues from its received size', async () => {
        const harness = buildHarness();
        jest.mocked(sendFileUploadChunk)
            .mockResolvedValueOnce({ receivedSize: 4 })
            .mockRejectedValueOnce(networkError())
            .mockResolvedValueOnce({ receivedSize: 10 });
        jest.mocked(getFileUploadSession)
            .mockRejectedValueOnce(networkError())
            .mockResolvedValueOnce({ uploadId: 'upload-1', size: 10, receivedSize: 6 });

        await expect(harness.run()).resolves.toEqual({ status: 'uploaded', name: 'big.bin' });

        expect(phaseCalls(harness)).toEqual([
            'checking',
            'uploading',
            'waitingConnection',
            'uploading',
        ]);
        expect(harness.clock.sleep).toHaveBeenCalledTimes(2);
        expect(harness.clock.sleep).toHaveBeenCalledWith(5000, harness.params.signal);
        expect(getFileUploadSession).toHaveBeenCalledWith('upload-1', harness.params.signal);
        const offsets = jest
            .mocked(sendFileUploadChunk)
            .mock.calls.map(([chunkParams]) => chunkParams.offset);
        expect(offsets).toEqual([0, 4, 6]);
        expect(sentBytesCalls(harness)).toEqual([0, 4, 6, 10]);
    });

    it('gives up with a connection lost error after ten minutes without an answer', async () => {
        const harness = buildHarness();
        jest.mocked(sendFileUploadChunk).mockRejectedValue(networkError());
        jest.mocked(getFileUploadSession).mockRejectedValue(networkError());

        await expect(harness.run()).rejects.toBeInstanceOf(UploadConnectionLostError);

        expect(harness.clock.sleep).toHaveBeenCalledTimes(120);
        expect(phaseCalls(harness)).toEqual(['checking', 'uploading', 'waitingConnection']);
    });

    it('propagates a session error that is not a network failure while waiting', async () => {
        const harness = buildHarness();
        const sessionGone = httpError(404);
        jest.mocked(sendFileUploadChunk).mockRejectedValue(networkError());
        jest.mocked(getFileUploadSession).mockRejectedValue(sessionGone);

        await expect(harness.run()).rejects.toBe(sessionGone);
    });

    it('retries init after a network failure until the server answers', async () => {
        const harness = buildHarness();
        jest.mocked(initFileUpload)
            .mockRejectedValueOnce(networkError())
            .mockResolvedValueOnce(startedSession(10));

        await harness.run();

        expect(initFileUpload).toHaveBeenCalledTimes(2);
        expect(phaseCalls(harness)).toEqual([
            'checking',
            'uploading',
            'waitingConnection',
            'uploading',
        ]);
    });

    it('retries complete after a network failure until the server answers', async () => {
        const harness = buildHarness();
        jest.mocked(initFileUpload).mockResolvedValue(startedSession(10));
        jest.mocked(completeFileUpload)
            .mockRejectedValueOnce(networkError())
            .mockResolvedValueOnce(completedFile());

        await expect(harness.run()).resolves.toEqual({ status: 'uploaded', name: 'big.bin' });

        expect(completeFileUpload).toHaveBeenCalledTimes(2);
    });

    it('does not recompute the checksum when it is already known', async () => {
        const harness = buildHarness({ knownChecksum: 'b'.repeat(64) });
        jest.mocked(initFileUpload).mockResolvedValue(startedSession(10));

        await harness.run();

        expect(computeFileSha256).not.toHaveBeenCalled();
        expect(harness.params.onChecksumComputed).toHaveBeenCalledWith('b'.repeat(64));
        expect(jest.mocked(initFileUpload).mock.calls[0]![0].checksum).toBe('b'.repeat(64));
    });

    it('cancels the server session without waiting when aborted after init', async () => {
        const harness = buildHarness();
        const abortError = new CanceledError();
        jest.mocked(sendFileUploadChunk).mockImplementation(async () => {
            harness.controller.abort();
            throw abortError;
        });

        await expect(harness.run()).rejects.toBe(abortError);

        expect(cancelFileUpload).toHaveBeenCalledWith('upload-1');
        expect(getFileUploadSession).not.toHaveBeenCalled();
    });

    it('swallows a failure of the session cancellation', async () => {
        const harness = buildHarness();
        const abortError = new CanceledError();
        jest.mocked(cancelFileUpload).mockRejectedValue(httpError(500));
        jest.mocked(sendFileUploadChunk).mockImplementation(async () => {
            harness.controller.abort();
            throw abortError;
        });

        await expect(harness.run()).rejects.toBe(abortError);
        await Promise.resolve();
    });

    it('does not call init when aborted while checking', async () => {
        const harness = buildHarness();
        const abortError = new DOMException('Aborted', 'AbortError');
        jest.mocked(computeFileSha256).mockImplementation(async () => {
            harness.controller.abort();
            throw abortError;
        });

        await expect(harness.run()).rejects.toBe(abortError);

        expect(initFileUpload).not.toHaveBeenCalled();
        expect(cancelFileUpload).not.toHaveBeenCalled();
    });

    it('stops waiting for the connection when aborted and cancels the session', async () => {
        const harness = buildHarness();
        const abortError = new DOMException('Aborted', 'AbortError');
        jest.mocked(sendFileUploadChunk).mockRejectedValue(networkError());
        harness.clock.sleep.mockImplementation(async () => {
            harness.controller.abort();
            throw abortError;
        });

        await expect(harness.run()).rejects.toBe(abortError);

        expect(cancelFileUpload).toHaveBeenCalledWith('upload-1');
    });
});

describe('uploadFileInChunks default sleep', () => {
    beforeEach(() => {
        jest.resetAllMocks();
        jest.useFakeTimers();
        jest.mocked(computeFileSha256).mockResolvedValue(checksum);
        jest.mocked(cancelFileUpload).mockResolvedValue();
        jest.mocked(readMisalignedReceivedSize).mockReturnValue(undefined);
        jest.mocked(initFileUpload).mockResolvedValue(startedSession());
        jest.mocked(completeFileUpload).mockResolvedValue(completedFile());
    });

    afterEach(() => jest.useRealTimers());

    const runWithRealSleep = (controller: AbortController) =>
        uploadFileInChunks({
            file,
            onConflict: 'rename',
            signal: controller.signal,
            onPhaseChange: jest.fn(),
            onChecksumComputed: jest.fn(),
            onSentBytes: jest.fn(),
        });

    it('waits the probe interval on the timer before asking the session', async () => {
        const controller = new AbortController();
        jest.mocked(sendFileUploadChunk)
            .mockRejectedValueOnce(networkError())
            .mockResolvedValueOnce({ receivedSize: 10 });
        jest.mocked(getFileUploadSession).mockResolvedValue({
            uploadId: 'upload-1',
            size: 10,
            receivedSize: 4,
        });

        const pendingUpload = runWithRealSleep(controller);
        await jest.advanceTimersByTimeAsync(4999);
        expect(getFileUploadSession).not.toHaveBeenCalled();
        await jest.advanceTimersByTimeAsync(1);
        await expect(pendingUpload).resolves.toEqual({ status: 'uploaded', name: 'big.bin' });

        expect(getFileUploadSession).toHaveBeenCalledTimes(1);
    });

    it('rejects the wait immediately when aborted during the timer', async () => {
        const controller = new AbortController();
        jest.mocked(sendFileUploadChunk).mockRejectedValue(networkError());

        const pendingUpload = runWithRealSleep(controller);
        const assertion = expect(pendingUpload).rejects.toBeDefined();
        await jest.advanceTimersByTimeAsync(1000);
        controller.abort();
        await assertion;

        expect(getFileUploadSession).not.toHaveBeenCalled();
        expect(cancelFileUpload).toHaveBeenCalledWith('upload-1');
    });

    it('does not wait at all when already aborted before the timer starts', async () => {
        const controller = new AbortController();
        jest.mocked(sendFileUploadChunk).mockImplementation(async () => {
            controller.abort();
            throw networkError();
        });

        await expect(runWithRealSleep(controller)).rejects.toBeDefined();

        expect(getFileUploadSession).not.toHaveBeenCalled();
    });
});

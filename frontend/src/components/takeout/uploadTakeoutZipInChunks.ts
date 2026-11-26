import {
    completeTakeoutUpload,
    getTakeoutUploadStatus,
    initTakeoutUpload,
    uploadTakeoutChunk,
} from '@/service/takeout';
import type { CompleteTakeoutUploadResponse } from '@/types/takeout';
import { isAxiosError } from 'axios';
import { TakeoutConnectionLostError } from './takeoutConnectionLostError';

export type TakeoutUploadDependencies = {
    sleep: (milliseconds: number, signal: AbortSignal) => Promise<void>;
    now: () => number;
    initialRetryDelayMs: number;
    maxRetryDelayMs: number;
    connectionLossLimitMs: number;
};

export type TakeoutUploadCallbacks = {
    onProgress: (uploadedBytes: number) => void;
    onWaitingConnection: (isWaiting: boolean) => void;
    onCompleting: () => void;
    onSessionStarted: (uploadId: string) => void;
};

const abortReasonOf = (signal: AbortSignal): unknown =>
    signal.reason ?? new DOMException('Aborted', 'AbortError');

const sleepUnlessAborted = (milliseconds: number, signal: AbortSignal): Promise<void> =>
    new Promise((resolve, reject) => {
        if (signal.aborted) {
            reject(abortReasonOf(signal));
            return;
        }
        const onAbort = () => {
            clearTimeout(timer);
            reject(abortReasonOf(signal));
        };
        const timer = setTimeout(() => {
            signal.removeEventListener('abort', onAbort);
            resolve();
        }, milliseconds);
        signal.addEventListener('abort', onAbort, { once: true });
    });

const OFFSET_MISMATCH_STATUS = 400;

export const defaultTakeoutUploadDependencies: TakeoutUploadDependencies = {
    sleep: sleepUnlessAborted,
    now: () => Date.now(),
    initialRetryDelayMs: 1000,
    maxRetryDelayMs: 30000,
    connectionLossLimitMs: 10 * 60 * 1000,
};

const isConnectionFailure = (error: unknown, signal: AbortSignal): boolean =>
    isAxiosError(error) && !error.response && !signal.aborted;

const isPossibleOffsetMismatch = (error: unknown): boolean =>
    isAxiosError(error) && error.response?.status === OFFSET_MISMATCH_STATUS;

const createConnectionRecovery = (
    dependencies: TakeoutUploadDependencies,
    signal: AbortSignal,
    onWaitingConnection: (isWaiting: boolean) => void
) => {
    let retryDelayMs = dependencies.initialRetryDelayMs;
    let connectionLostAt: number | null = null;

    const waitBeforeRetry = async () => {
        const now = dependencies.now();
        connectionLostAt ??= now;
        if (now - connectionLostAt >= dependencies.connectionLossLimitMs) {
            throw new TakeoutConnectionLostError();
        }
        onWaitingConnection(true);
        await dependencies.sleep(retryDelayMs, signal);
        retryDelayMs = Math.min(retryDelayMs * 2, dependencies.maxRetryDelayMs);
    };

    const markConnected = () => {
        if (connectionLostAt === null) {
            return;
        }
        connectionLostAt = null;
        retryDelayMs = dependencies.initialRetryDelayMs;
        onWaitingConnection(false);
    };

    return { waitBeforeRetry, markConnected };
};

type ConnectionRecovery = ReturnType<typeof createConnectionRecovery>;

const readReceivedSizeWhenConnected = async (
    uploadId: string,
    recovery: ConnectionRecovery,
    signal: AbortSignal
): Promise<number> => {
    for (;;) {
        await recovery.waitBeforeRetry();
        try {
            const status = await getTakeoutUploadStatus(uploadId, signal);
            recovery.markConnected();
            return status.received_size;
        } catch (statusError) {
            if (!isConnectionFailure(statusError, signal)) throw statusError;
        }
    }
};

const readReceivedSizeAfterMismatch = async (
    uploadId: string,
    rejectedOffset: number,
    rejection: unknown,
    signal: AbortSignal
): Promise<number> => {
    const status = await getTakeoutUploadStatus(uploadId, signal);
    if (status.received_size === rejectedOffset) throw rejection;
    return status.received_size;
};

const completeWhenConnected = async (
    uploadId: string,
    recovery: ConnectionRecovery,
    signal: AbortSignal
): Promise<CompleteTakeoutUploadResponse> => {
    for (;;) {
        try {
            const completed = await completeTakeoutUpload(uploadId, signal);
            recovery.markConnected();
            return completed;
        } catch (completeError) {
            if (!isConnectionFailure(completeError, signal)) throw completeError;
            await recovery.waitBeforeRetry();
        }
    }
};

const uploadAndComplete = async (
    file: File,
    callbacks: TakeoutUploadCallbacks,
    signal: AbortSignal,
    dependencies: TakeoutUploadDependencies = defaultTakeoutUploadDependencies
): Promise<CompleteTakeoutUploadResponse> => {
    const recovery = createConnectionRecovery(dependencies, signal, callbacks.onWaitingConnection);
    const session = await initTakeoutUpload(file.name, file.size, signal);
    callbacks.onSessionStarted(session.upload_id);
    const chunkSize = Math.max(1, session.chunk_size);

    let offset = session.resumed ? session.received_size : 0;
    callbacks.onProgress(offset);

    while (offset < file.size) {
        const chunk = file.slice(offset, offset + chunkSize);
        try {
            await uploadTakeoutChunk(session.upload_id, chunk, offset, signal);
            recovery.markConnected();
            offset += chunk.size;
        } catch (chunkError) {
            if (isConnectionFailure(chunkError, signal)) {
                offset = await readReceivedSizeWhenConnected(session.upload_id, recovery, signal);
            } else if (isPossibleOffsetMismatch(chunkError)) {
                offset = await readReceivedSizeAfterMismatch(
                    session.upload_id,
                    offset,
                    chunkError,
                    signal
                );
            } else {
                throw chunkError;
            }
        }
        callbacks.onProgress(offset);
    }

    callbacks.onCompleting();
    return completeWhenConnected(session.upload_id, recovery, signal);
};

export const uploadTakeoutZipInChunks = async (
    file: File,
    callbacks: TakeoutUploadCallbacks,
    signal: AbortSignal,
    dependencies: TakeoutUploadDependencies = defaultTakeoutUploadDependencies
): Promise<CompleteTakeoutUploadResponse> => {
    try {
        return await uploadAndComplete(file, callbacks, signal, dependencies);
    } catch (uploadError) {
        if (signal.aborted) throw abortReasonOf(signal);
        throw uploadError;
    }
};

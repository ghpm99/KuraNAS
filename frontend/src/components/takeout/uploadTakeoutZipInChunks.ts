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
    sleep: (milliseconds: number) => Promise<void>;
    now: () => number;
    initialRetryDelayMs: number;
    maxRetryDelayMs: number;
    connectionLossLimitMs: number;
};

export type TakeoutUploadCallbacks = {
    onProgress: (uploadedBytes: number) => void;
    onWaitingConnection: (isWaiting: boolean) => void;
    onCompleting: () => void;
};

const OFFSET_MISMATCH_STATUS = 400;

export const defaultTakeoutUploadDependencies: TakeoutUploadDependencies = {
    sleep: (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds)),
    now: () => Date.now(),
    initialRetryDelayMs: 1000,
    maxRetryDelayMs: 30000,
    connectionLossLimitMs: 10 * 60 * 1000,
};

const isConnectionFailure = (error: unknown): boolean => isAxiosError(error) && !error.response;

const isPossibleOffsetMismatch = (error: unknown): boolean =>
    isAxiosError(error) && error.response?.status === OFFSET_MISMATCH_STATUS;

const createConnectionRecovery = (
    dependencies: TakeoutUploadDependencies,
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
        await dependencies.sleep(retryDelayMs);
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
    recovery: ConnectionRecovery
): Promise<number> => {
    for (;;) {
        await recovery.waitBeforeRetry();
        try {
            const status = await getTakeoutUploadStatus(uploadId);
            recovery.markConnected();
            return status.received_size;
        } catch (statusError) {
            if (!isConnectionFailure(statusError)) throw statusError;
        }
    }
};

const readReceivedSizeAfterMismatch = async (
    uploadId: string,
    rejectedOffset: number,
    rejection: unknown
): Promise<number> => {
    const status = await getTakeoutUploadStatus(uploadId);
    if (status.received_size === rejectedOffset) throw rejection;
    return status.received_size;
};

const completeWhenConnected = async (
    uploadId: string,
    recovery: ConnectionRecovery
): Promise<CompleteTakeoutUploadResponse> => {
    for (;;) {
        try {
            const completed = await completeTakeoutUpload(uploadId);
            recovery.markConnected();
            return completed;
        } catch (completeError) {
            if (!isConnectionFailure(completeError)) throw completeError;
            await recovery.waitBeforeRetry();
        }
    }
};

export const uploadTakeoutZipInChunks = async (
    file: File,
    callbacks: TakeoutUploadCallbacks,
    dependencies: TakeoutUploadDependencies = defaultTakeoutUploadDependencies
): Promise<CompleteTakeoutUploadResponse> => {
    const recovery = createConnectionRecovery(dependencies, callbacks.onWaitingConnection);
    const session = await initTakeoutUpload(file.name, file.size);
    const chunkSize = Math.max(1, session.chunk_size);

    let offset = session.resumed ? session.received_size : 0;
    callbacks.onProgress(offset);

    while (offset < file.size) {
        const chunk = file.slice(offset, offset + chunkSize);
        try {
            await uploadTakeoutChunk(session.upload_id, chunk, offset);
            recovery.markConnected();
            offset += chunk.size;
        } catch (chunkError) {
            if (isConnectionFailure(chunkError)) {
                offset = await readReceivedSizeWhenConnected(session.upload_id, recovery);
            } else if (isPossibleOffsetMismatch(chunkError)) {
                offset = await readReceivedSizeAfterMismatch(session.upload_id, offset, chunkError);
            } else {
                throw chunkError;
            }
        }
        callbacks.onProgress(offset);
    }

    callbacks.onCompleting();
    return completeWhenConnected(session.upload_id, recovery);
};

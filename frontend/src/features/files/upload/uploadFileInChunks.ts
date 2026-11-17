import axios, { isAxiosError } from 'axios';
import {
    cancelFileUpload,
    completeFileUpload,
    getFileUploadSession,
    initFileUpload,
    readMisalignedReceivedSize,
    sendFileUploadChunk,
} from '@/service/fileUpload';
import type { FileUploadSessionStart } from '@/types/fileUpload';
import { extractBackendMessage } from './backendErrorMessage';
import { computeFileSha256 } from './computeFileSha256';
import type { FileTransferOutcome, FileTransferParams } from './fileTransferTypes';
import { UploadConnectionLostError } from './uploadConnectionLostError';

export type ChunkedUploadDependencies = {
    sleep: (milliseconds: number, signal: AbortSignal) => Promise<void>;
    now: () => number;
    reconnectProbeIntervalMs: number;
    connectionLossLimitMs: number;
};

type ChunkedUploadContext = {
    params: FileTransferParams;
    dependencies: ChunkedUploadDependencies;
};

type StartedUploadSession = FileUploadSessionStart & { uploadId: string };

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

const defaultDependencies: ChunkedUploadDependencies = {
    sleep: sleepUnlessAborted,
    now: Date.now,
    reconnectProbeIntervalMs: 5000,
    connectionLossLimitMs: 10 * 60 * 1000,
};

export const isNetworkFailure = (error: unknown, signal: AbortSignal): boolean =>
    isAxiosError(error) && !error.response && !axios.isCancel(error) && !signal.aborted;

const waitForConnection = async <Probed>(
    probeConnection: () => Promise<Probed>,
    { params, dependencies }: ChunkedUploadContext
): Promise<Probed> => {
    const connectionLostAt = dependencies.now();
    params.onPhaseChange('waitingConnection');
    for (;;) {
        await dependencies.sleep(dependencies.reconnectProbeIntervalMs, params.signal);
        try {
            const probed = await probeConnection();
            params.onPhaseChange('uploading');
            return probed;
        } catch (error) {
            if (!isNetworkFailure(error, params.signal)) throw error;
        }
        if (dependencies.now() - connectionLostAt >= dependencies.connectionLossLimitMs) {
            throw new UploadConnectionLostError();
        }
    }
};

const retryUntilConnected = async <Completed>(
    operation: () => Promise<Completed>,
    context: ChunkedUploadContext
): Promise<Completed> => {
    try {
        return await operation();
    } catch (error) {
        if (!isNetworkFailure(error, context.params.signal)) throw error;
        return waitForConnection(operation, context);
    }
};

const resolveChecksum = async ({
    file,
    signal,
    knownChecksum,
    onPhaseChange,
    onChecksumComputed,
}: FileTransferParams) => {
    onPhaseChange('checking');
    const checksum = knownChecksum ?? (await computeFileSha256(file, { signal }));
    onChecksumComputed(checksum);
    return checksum;
};

const startUploadSession = async (
    checksum: string,
    context: ChunkedUploadContext
): Promise<FileUploadSessionStart> => {
    const { file, targetFolderId, relativePath, onConflict, signal } = context.params;
    context.params.onPhaseChange('uploading');
    return retryUntilConnected(
        () =>
            initFileUpload({
                fileName: file.name,
                size: file.size,
                targetFolderId,
                relativePath,
                checksum,
                modifiedAt: new Date(file.lastModified).toISOString(),
                onConflict,
                signal,
            }),
        context
    );
};

const isNameConflictOnInit = (error: unknown, params: FileTransferParams): boolean =>
    params.onConflict === 'fail' && isAxiosError(error) && error.response?.status === 409;

const startUploadSessionOrReportConflict = async (
    checksum: string,
    context: ChunkedUploadContext
): Promise<FileUploadSessionStart | FileTransferOutcome> => {
    try {
        return await startUploadSession(checksum, context);
    } catch (error) {
        if (!isNameConflictOnInit(error, context.params)) throw error;
        return { status: 'conflict', error: extractBackendMessage(error) };
    }
};

const sendNextChunk = async (
    session: StartedUploadSession,
    offset: number,
    context: ChunkedUploadContext
): Promise<number> => {
    const { file, signal, onSentBytes } = context.params;
    try {
        const { receivedSize } = await sendFileUploadChunk({
            uploadId: session.uploadId,
            chunk: file.slice(offset, offset + session.chunkSize),
            offset,
            signal,
            onProgress: (loaded) => onSentBytes(offset + loaded),
        });
        return receivedSize;
    } catch (error) {
        const misalignedReceivedSize = readMisalignedReceivedSize(error);
        if (misalignedReceivedSize !== undefined) return misalignedReceivedSize;
        if (!isNetworkFailure(error, signal)) throw error;
        const resumedSession = await waitForConnection(
            () => getFileUploadSession(session.uploadId, signal),
            context
        );
        return resumedSession.receivedSize;
    }
};

const sendRemainingChunks = async (
    session: StartedUploadSession,
    context: ChunkedUploadContext
): Promise<void> => {
    const { file, onSentBytes } = context.params;
    let offset = session.receivedSize;
    onSentBytes(offset);
    while (offset < file.size) {
        offset = await sendNextChunk(session, offset, context);
        onSentBytes(offset);
    }
};

const finishUpload = async (
    session: StartedUploadSession,
    context: ChunkedUploadContext
): Promise<FileTransferOutcome> => {
    const completed = await retryUntilConnected(
        () => completeFileUpload(session.uploadId, context.params.signal),
        context
    );
    return { status: completed.file.status, name: completed.file.name };
};

const cancelServerSessionWithoutWaiting = (uploadId: string): void => {
    void cancelFileUpload(uploadId).catch(() => undefined);
};

const requireUploadId = (session: FileUploadSessionStart): StartedUploadSession => {
    if (!session.uploadId) throw new Error('upload session started without an upload id');
    return { ...session, uploadId: session.uploadId };
};

const transferSession = async (
    session: StartedUploadSession,
    context: ChunkedUploadContext
): Promise<FileTransferOutcome> => {
    try {
        await sendRemainingChunks(session, context);
        return await finishUpload(session, context);
    } catch (error) {
        if (context.params.signal.aborted) cancelServerSessionWithoutWaiting(session.uploadId);
        throw error;
    }
};

/**
 * Uploads a file in chunks through the resumable upload session endpoints.
 * Resolves with `alreadyExisted` without sending bytes when the server already holds the file, and
 * with `conflict` when `init` answers 409 under the `fail` policy.
 * Rejects with the HTTP error of `init`/`chunk`/`complete`, with `UploadConnectionLostError` after
 * `connectionLossLimitMs` without connection, or with the abort reason (also cancelling the session).
 */
export const uploadFileInChunks = async (
    params: FileTransferParams,
    dependencyOverrides: Partial<ChunkedUploadDependencies> = {}
): Promise<FileTransferOutcome> => {
    const context: ChunkedUploadContext = {
        params,
        dependencies: { ...defaultDependencies, ...dependencyOverrides },
    };
    const checksum = await resolveChecksum(params);
    const startedSession = await startUploadSessionOrReportConflict(checksum, context);
    if ('status' in startedSession) return startedSession;
    if (startedSession.alreadyExists) return { status: 'alreadyExisted' };
    return transferSession(requireUploadId(startedSession), context);
};

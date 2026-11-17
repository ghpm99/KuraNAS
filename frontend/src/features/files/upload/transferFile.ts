import { uploadSingleFile } from '@/service/files';
import type { FileTransferOutcome, FileTransferParams } from './fileTransferTypes';
import { uploadFileInChunks } from './uploadFileInChunks';

export const chunkedUploadThresholdBytes = 64 * 1024 * 1024;

const uploadInSingleRequest = (params: FileTransferParams): Promise<FileTransferOutcome> => {
    params.onPhaseChange('uploading');
    return uploadSingleFile({
        file: params.file,
        targetFolderId: params.targetFolderId,
        relativePath: params.relativePath,
        onConflict: params.onConflict,
        signal: params.signal,
        onProgress: params.onSentBytes,
    });
};

/**
 * Sends the file in one request up to `chunkedUploadThresholdBytes`, and through the resumable
 * chunked upload above it.
 */
export const transferFile = (params: FileTransferParams): Promise<FileTransferOutcome> =>
    params.file.size <= chunkedUploadThresholdBytes
        ? uploadInSingleRequest(params)
        : uploadFileInChunks(params);

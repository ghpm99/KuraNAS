import { isAxiosError } from 'axios';
import { apiBase } from '@/service';
import type {
    CompletedFileUpload,
    CompletedUploadStatus,
    FileUploadSession,
    FileUploadSessionStart,
    InitFileUploadParams,
    SendFileUploadChunkParams,
} from '@/types/fileUpload';

type InitFileUploadResponse = {
    upload_id?: string;
    chunk_size: number;
    received_size: number;
    resumed: boolean;
    already_exists: boolean;
};

type ChunkResponse = { received_size: number };

type SessionResponse = { upload_id: string; size: number; received_size: number };

type CompleteResponse = {
    file: { name: string; path: string; status: CompletedUploadStatus };
    job_id: number;
};

export const initFileUpload = async ({
    fileName,
    size,
    targetFolderId,
    relativePath,
    checksum,
    modifiedAt,
    onConflict,
    signal,
}: InitFileUploadParams): Promise<FileUploadSessionStart> => {
    const payload = {
        file_name: fileName,
        size,
        on_conflict: onConflict,
        ...(targetFolderId !== undefined && targetFolderId > 0
            ? { target_folder_id: targetFolderId }
            : {}),
        ...(relativePath ? { relative_path: relativePath } : {}),
        ...(checksum ? { checksum } : {}),
        ...(modifiedAt ? { modified_at: modifiedAt } : {}),
    };
    const response = await apiBase.post<InitFileUploadResponse>('/files/upload/init', payload, {
        signal,
    });
    return {
        uploadId: response.data.upload_id,
        chunkSize: response.data.chunk_size,
        receivedSize: response.data.received_size,
        resumed: response.data.resumed,
        alreadyExists: response.data.already_exists,
    };
};

export const sendFileUploadChunk = async ({
    uploadId,
    chunk,
    offset,
    signal,
    onProgress,
}: SendFileUploadChunkParams): Promise<{ receivedSize: number }> => {
    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('upload_id', uploadId);
    formData.append('offset', String(offset));

    const response = await apiBase.post<ChunkResponse>('/files/upload/chunk', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
        signal,
        onUploadProgress: (event) => onProgress?.(event.loaded),
    });
    return { receivedSize: response.data.received_size };
};

export const getFileUploadSession = async (
    uploadId: string,
    signal?: AbortSignal
): Promise<FileUploadSession> => {
    const response = await apiBase.get<SessionResponse>(`/files/upload/${uploadId}`, { signal });
    return {
        uploadId: response.data.upload_id,
        size: response.data.size,
        receivedSize: response.data.received_size,
    };
};

export const completeFileUpload = async (
    uploadId: string,
    signal?: AbortSignal
): Promise<CompletedFileUpload> => {
    const response = await apiBase.post<CompleteResponse>(
        '/files/upload/complete',
        { upload_id: uploadId },
        { signal }
    );
    return { file: response.data.file, jobId: response.data.job_id };
};

export const cancelFileUpload = async (uploadId: string): Promise<void> => {
    await apiBase.delete(`/files/upload/${uploadId}`);
};

export const readMisalignedReceivedSize = (error: unknown): number | undefined => {
    if (!isAxiosError(error) || error.response?.status !== 409) {
        return undefined;
    }
    const receivedSize = (error.response.data as { received_size?: unknown } | undefined)
        ?.received_size;
    return typeof receivedSize === 'number' ? receivedSize : undefined;
};

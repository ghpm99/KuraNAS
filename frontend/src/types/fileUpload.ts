import type { UploadConflictPolicy } from '@/service/files';

export type InitFileUploadParams = {
    fileName: string;
    size: number;
    targetFolderId?: number;
    relativePath?: string;
    checksum?: string;
    modifiedAt?: string;
    onConflict: UploadConflictPolicy;
    signal?: AbortSignal;
};

export type FileUploadSessionStart = {
    uploadId?: string;
    chunkSize: number;
    receivedSize: number;
    resumed: boolean;
    alreadyExists: boolean;
};

export type SendFileUploadChunkParams = {
    uploadId: string;
    chunk: Blob;
    offset: number;
    signal?: AbortSignal;
    onProgress?: (sentBytes: number) => void;
};

export type FileUploadSession = {
    uploadId: string;
    size: number;
    receivedSize: number;
};

export type CompletedUploadStatus = 'uploaded' | 'replaced' | 'renamed' | 'skipped';

export type CompletedFileUpload = {
    file: { name: string; path: string; status: CompletedUploadStatus };
    jobId: number;
};

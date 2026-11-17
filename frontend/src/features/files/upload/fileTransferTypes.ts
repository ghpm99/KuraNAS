import type { UploadConflictPolicy } from '@/service/files';

export type FileTransferPhase = 'checking' | 'uploading' | 'waitingConnection';

export type FileTransferParams = {
    file: File;
    targetFolderId?: number;
    relativePath?: string;
    onConflict: UploadConflictPolicy;
    signal: AbortSignal;
    knownChecksum?: string;
    onPhaseChange: (phase: FileTransferPhase) => void;
    onChecksumComputed: (checksum: string) => void;
    onSentBytes: (sentBytes: number) => void;
};

export type FileTransferOutcome = {
    status:
        | 'uploaded'
        | 'renamed'
        | 'replaced'
        | 'skipped'
        | 'alreadyExisted'
        | 'conflict'
        | 'failed';
    name?: string;
    error?: string;
};

export type TransferFileFn = (params: FileTransferParams) => Promise<FileTransferOutcome>;

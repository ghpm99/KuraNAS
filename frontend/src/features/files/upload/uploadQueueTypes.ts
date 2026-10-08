import type { UploadConflictPolicy } from '@/service/files';

export type UploadEntry = {
    file: File;
    relativePath?: string;
};

export type UploadItemStatus =
    | 'queued'
    | 'uploading'
    | 'done'
    | 'renamed'
    | 'replaced'
    | 'skipped'
    | 'failed'
    | 'canceled';

export type UploadItem = {
    id: string;
    file: File;
    displayName: string;
    relativePath?: string;
    targetFolderId?: number;
    onConflict: UploadConflictPolicy;
    status: UploadItemStatus;
    progress: number;
    error?: string;
    savedName?: string;
};

export type UploadQueue = {
    items: UploadItem[];
    conflictPolicy: UploadConflictPolicy;
    setConflictPolicy: (policy: UploadConflictPolicy) => void;
    enqueue: (entries: UploadEntry[], targetFolderId?: number) => void;
    cancel: (itemId: string) => void;
    cancelAll: () => void;
    retry: (itemId: string) => void;
    retryFailed: () => void;
    clearFinished: () => void;
};

const activeStatuses: UploadItemStatus[] = ['queued', 'uploading'];

export const isActiveItem = (item: UploadItem): boolean => activeStatuses.includes(item.status);

export const isRetryableItem = (item: UploadItem): boolean =>
    item.status === 'failed' || item.status === 'canceled';

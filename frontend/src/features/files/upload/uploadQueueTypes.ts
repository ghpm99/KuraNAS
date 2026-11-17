import type { UploadConflictChoice } from './conflictPolicyPreference';

export type UploadEntry = {
    file: File;
    relativePath?: string;
};

export type UploadItemStatus =
    | 'queued'
    | 'checking'
    | 'uploading'
    | 'waitingConnection'
    | 'done'
    | 'renamed'
    | 'replaced'
    | 'skipped'
    | 'alreadyExisted'
    | 'awaitingDecision'
    | 'failed'
    | 'canceled';

export type ResolvedConflictChoice = Exclude<UploadConflictChoice, 'ask'>;

export type UploadItem = {
    id: string;
    file: File;
    displayName: string;
    relativePath?: string;
    targetFolderId?: number;
    onConflict: UploadConflictChoice;
    status: UploadItemStatus;
    size: number;
    sentBytes: number;
    progress: number;
    bytesPerSecond?: number;
    remainingSeconds?: number;
    checksum?: string;
    error?: string;
    errorKey?: string;
    savedName?: string;
};

export type UploadQueue = {
    items: UploadItem[];
    conflictPolicy: UploadConflictChoice;
    setConflictPolicy: (policy: UploadConflictChoice) => void;
    enqueue: (entries: UploadEntry[], targetFolderId?: number) => void;
    cancel: (itemId: string) => void;
    cancelAll: () => void;
    retry: (itemId: string) => void;
    retryFailed: () => void;
    clearFinished: () => void;
    resolveConflict: (itemId: string, choice: ResolvedConflictChoice, applyToAll: boolean) => void;
    dismissConflict: (itemId: string) => void;
};

const activeStatuses: UploadItemStatus[] = ['queued', 'checking', 'uploading', 'waitingConnection'];

export const isActiveItem = (item: UploadItem): boolean => activeStatuses.includes(item.status);

export const isFinishedItem = (item: UploadItem): boolean =>
    !isActiveItem(item) && item.status !== 'awaitingDecision';

export const isAwaitingDecision = (item: UploadItem): boolean => item.status === 'awaitingDecision';

export const isRetryableItem = (item: UploadItem): boolean =>
    item.status === 'failed' || item.status === 'canceled';

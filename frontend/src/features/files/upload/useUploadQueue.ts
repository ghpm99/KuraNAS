import { useCallback, useRef, useState } from 'react';
import {
    uploadSingleFile,
    type UploadConflictPolicy,
    type UploadOutcome,
    type UploadSingleFileParams,
} from '@/service/files';
import { loadConflictPolicy, saveConflictPolicy } from './conflictPolicyPreference';
import {
    isActiveItem,
    isRetryableItem,
    type UploadEntry,
    type UploadItem,
    type UploadItemStatus,
    type UploadQueue,
} from './uploadQueueTypes';

const maxConcurrentUploads = 3;

type UploadFileFn = (params: UploadSingleFileParams) => Promise<UploadOutcome>;

type UseUploadQueueOptions = {
    uploadFile?: UploadFileFn;
    maxConcurrent?: number;
    onFileFinished?: () => void;
};

const statusByOutcome: Record<UploadOutcome['status'], UploadItemStatus> = {
    uploaded: 'done',
    renamed: 'renamed',
    replaced: 'replaced',
    skipped: 'skipped',
    failed: 'failed',
};

const extractBackendMessage = (error: unknown): string | undefined => {
    const backendMessage = (error as { response?: { data?: { error?: unknown } } })?.response?.data
        ?.error;
    return typeof backendMessage === 'string' && backendMessage !== '' ? backendMessage : undefined;
};

const buildItem = (
    entry: UploadEntry,
    itemId: string,
    targetFolderId: number | undefined,
    onConflict: UploadConflictPolicy
): UploadItem => ({
    id: itemId,
    file: entry.file,
    displayName: entry.relativePath || entry.file.name,
    relativePath: entry.relativePath,
    targetFolderId,
    onConflict,
    status: 'queued',
    progress: 0,
});

export const useUploadQueue = ({
    uploadFile = uploadSingleFile,
    maxConcurrent = maxConcurrentUploads,
    onFileFinished,
}: UseUploadQueueOptions = {}): UploadQueue => {
    const [items, setItems] = useState<UploadItem[]>([]);
    const [conflictPolicy, setConflictPolicyState] = useState<UploadConflictPolicy>(loadConflictPolicy);
    const itemsRef = useRef<UploadItem[]>([]);
    const controllersRef = useRef(new Map<string, AbortController>());
    const runningCountRef = useRef(0);
    const nextIdRef = useRef(0);
    const conflictPolicyRef = useRef(conflictPolicy);
    const onFileFinishedRef = useRef(onFileFinished);
    onFileFinishedRef.current = onFileFinished;

    const commitItems = useCallback((nextItems: UploadItem[]) => {
        itemsRef.current = nextItems;
        setItems(nextItems);
    }, []);

    const patchItem = useCallback(
        (itemId: string, patch: Partial<UploadItem>) => {
            commitItems(
                itemsRef.current.map((item) => (item.id === itemId ? { ...item, ...patch } : item))
            );
        },
        [commitItems]
    );

    const pumpRef = useRef<() => void>(() => undefined);

    const runItem = useCallback(
        async (item: UploadItem) => {
            const controller = new AbortController();
            controllersRef.current.set(item.id, controller);
            runningCountRef.current += 1;
            patchItem(item.id, { status: 'uploading', progress: 0, error: undefined });

            try {
                const outcome = await uploadFile({
                    file: item.file,
                    targetFolderId: item.targetFolderId,
                    relativePath: item.relativePath,
                    onConflict: item.onConflict,
                    signal: controller.signal,
                    onProgress: (percent) => {
                        const current = itemsRef.current.find((candidate) => candidate.id === item.id);
                        if (current && current.status === 'uploading' && current.progress !== percent) {
                            patchItem(item.id, { progress: percent });
                        }
                    },
                });
                patchItem(item.id, {
                    status: statusByOutcome[outcome.status] ?? 'done',
                    progress: 100,
                    error: outcome.error,
                    savedName: outcome.name,
                });
                onFileFinishedRef.current?.();
            } catch (error) {
                if (controller.signal.aborted) {
                    patchItem(item.id, { status: 'canceled', progress: 0 });
                } else {
                    patchItem(item.id, { status: 'failed', error: extractBackendMessage(error) });
                }
            } finally {
                controllersRef.current.delete(item.id);
                runningCountRef.current -= 1;
                pumpRef.current();
            }
        },
        [patchItem, uploadFile]
    );

    const pump = useCallback(() => {
        while (runningCountRef.current < maxConcurrent) {
            const nextItem = itemsRef.current.find((item) => item.status === 'queued');
            if (!nextItem) return;
            void runItem(nextItem);
        }
    }, [maxConcurrent, runItem]);
    pumpRef.current = pump;

    const enqueue = useCallback(
        (entries: UploadEntry[], targetFolderId?: number) => {
            if (entries.length === 0) return;
            const newItems = entries.map((entry) => {
                nextIdRef.current += 1;
                return buildItem(
                    entry,
                    `upload-${nextIdRef.current}`,
                    targetFolderId,
                    conflictPolicyRef.current
                );
            });
            commitItems([...itemsRef.current, ...newItems]);
            pump();
        },
        [commitItems, pump]
    );

    const cancel = useCallback(
        (itemId: string) => {
            const item = itemsRef.current.find((candidate) => candidate.id === itemId);
            if (!item) return;
            if (item.status === 'queued') {
                patchItem(itemId, { status: 'canceled' });
                return;
            }
            controllersRef.current.get(itemId)?.abort();
        },
        [patchItem]
    );

    const cancelAll = useCallback(() => {
        commitItems(
            itemsRef.current.map((item) =>
                item.status === 'queued' ? { ...item, status: 'canceled' } : item
            )
        );
        controllersRef.current.forEach((controller) => controller.abort());
    }, [commitItems]);

    const requeue = useCallback(
        (shouldRequeue: (item: UploadItem) => boolean) => {
            commitItems(
                itemsRef.current.map((item) =>
                    shouldRequeue(item)
                        ? { ...item, status: 'queued', progress: 0, error: undefined }
                        : item
                )
            );
            pump();
        },
        [commitItems, pump]
    );

    const retry = useCallback(
        (itemId: string) => requeue((item) => item.id === itemId && isRetryableItem(item)),
        [requeue]
    );

    const retryFailed = useCallback(
        () => requeue((item) => item.status === 'failed'),
        [requeue]
    );

    const clearFinished = useCallback(() => {
        commitItems(itemsRef.current.filter(isActiveItem));
    }, [commitItems]);

    const setConflictPolicy = useCallback((policy: UploadConflictPolicy) => {
        conflictPolicyRef.current = policy;
        saveConflictPolicy(policy);
        setConflictPolicyState(policy);
    }, []);

    return {
        items,
        conflictPolicy,
        setConflictPolicy,
        enqueue,
        cancel,
        cancelAll,
        retry,
        retryFailed,
        clearFinished,
    };
};

export default useUploadQueue;

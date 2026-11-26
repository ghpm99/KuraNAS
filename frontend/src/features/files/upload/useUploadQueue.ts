import { useCallback, useRef, useState } from 'react';
import { extractBackendMessage } from './backendErrorMessage';
import {
    loadConflictPolicy,
    saveConflictPolicy,
    toWireConflictPolicy,
    type UploadConflictChoice,
} from './conflictPolicyPreference';
import type { FileTransferOutcome, FileTransferPhase, TransferFileFn } from './fileTransferTypes';
import { transferFile as transferFileByChunkingWhenLarge } from './transferFile';
import { createTransferRateTracker } from '@/shared/transfer/transferRateTracker';
import { connectionLostErrorKey, UploadConnectionLostError } from './uploadConnectionLostError';
import {
    isAwaitingDecision,
    isFinishedItem,
    isRetryableItem,
    type ResolvedConflictChoice,
    type UploadEntry,
    type UploadItem,
    type UploadItemStatus,
    type UploadQueue,
} from './uploadQueueTypes';

const maxConcurrentUploads = 3;

const progressPatchIntervalMs = 250;

type UseUploadQueueOptions = {
    transferFile?: TransferFileFn;
    maxConcurrent?: number;
    onFileFinished?: () => void;
};

const statusByOutcome: Record<FileTransferOutcome['status'], UploadItemStatus> = {
    uploaded: 'done',
    renamed: 'renamed',
    replaced: 'replaced',
    skipped: 'skipped',
    alreadyExisted: 'alreadyExisted',
    conflict: 'awaitingDecision',
    failed: 'failed',
};

const buildOutcomePatch = (outcome: FileTransferOutcome, size: number): Partial<UploadItem> => {
    const isAwaitingConflictDecision = outcome.status === 'conflict';
    return {
        status: statusByOutcome[outcome.status] ?? 'done',
        sentBytes: isAwaitingConflictDecision ? 0 : size,
        progress: isAwaitingConflictDecision ? 0 : 100,
        bytesPerSecond: undefined,
        remainingSeconds: undefined,
        error: outcome.error,
        savedName: outcome.name,
    };
};

const buildItem = (
    entry: UploadEntry,
    itemId: string,
    targetFolderId: number | undefined,
    onConflict: UploadConflictChoice
): UploadItem => ({
    id: itemId,
    file: entry.file,
    displayName: entry.relativePath || entry.file.name,
    relativePath: entry.relativePath,
    targetFolderId,
    onConflict,
    status: 'queued',
    size: entry.file.size,
    sentBytes: 0,
    progress: 0,
});

const toProgressPercent = (sentBytes: number, size: number): number =>
    size > 0 ? Math.min(100, Math.floor((sentBytes / size) * 100)) : 0;

const phaseToStatus: Record<FileTransferPhase, UploadItemStatus> = {
    checking: 'checking',
    uploading: 'uploading',
    waitingConnection: 'waitingConnection',
};

const buildFailurePatch = (error: unknown): Partial<UploadItem> =>
    error instanceof UploadConnectionLostError
        ? { status: 'failed', errorKey: connectionLostErrorKey }
        : { status: 'failed', error: extractBackendMessage(error) };

export const useUploadQueue = ({
    transferFile = transferFileByChunkingWhenLarge,
    maxConcurrent = maxConcurrentUploads,
    onFileFinished,
}: UseUploadQueueOptions = {}): UploadQueue => {
    const [items, setItems] = useState<UploadItem[]>([]);
    const [conflictPolicy, setConflictPolicyState] =
        useState<UploadConflictChoice>(loadConflictPolicy);
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
            const rateTracker = createTransferRateTracker();
            let lastPatchAt = 0;
            let lastPercent = -1;
            patchItem(item.id, {
                status: 'uploading',
                sentBytes: 0,
                progress: 0,
                error: undefined,
                errorKey: undefined,
            });

            const handlePhaseChange = (phase: FileTransferPhase) => {
                rateTracker.reset();
                patchItem(item.id, {
                    status: phaseToStatus[phase],
                    bytesPerSecond: undefined,
                    remainingSeconds: undefined,
                });
            };

            const handleSentBytes = (sentBytes: number) => {
                const estimate = rateTracker.record(sentBytes, item.size);
                const current = itemsRef.current.find((candidate) => candidate.id === item.id);
                if (current?.status !== 'uploading') return;
                const percent = toProgressPercent(sentBytes, item.size);
                const patchedAt = Date.now();
                const isPatchDue = patchedAt - lastPatchAt >= progressPatchIntervalMs;
                if (!isPatchDue && percent === lastPercent) return;
                lastPatchAt = patchedAt;
                lastPercent = percent;
                patchItem(item.id, {
                    sentBytes,
                    progress: percent,
                    bytesPerSecond: estimate.bytesPerSecond,
                    remainingSeconds: estimate.remainingSeconds,
                });
            };

            try {
                const outcome = await transferFile({
                    file: item.file,
                    targetFolderId: item.targetFolderId,
                    relativePath: item.relativePath,
                    onConflict: toWireConflictPolicy(item.onConflict),
                    signal: controller.signal,
                    knownChecksum: item.checksum,
                    onPhaseChange: handlePhaseChange,
                    onChecksumComputed: (checksum) => patchItem(item.id, { checksum }),
                    onSentBytes: handleSentBytes,
                });
                patchItem(item.id, buildOutcomePatch(outcome, item.size));
                onFileFinishedRef.current?.();
            } catch (error) {
                if (controller.signal.aborted) {
                    patchItem(item.id, {
                        status: 'canceled',
                        sentBytes: 0,
                        progress: 0,
                        bytesPerSecond: undefined,
                        remainingSeconds: undefined,
                    });
                } else {
                    patchItem(item.id, {
                        ...buildFailurePatch(error),
                        bytesPerSecond: undefined,
                        remainingSeconds: undefined,
                    });
                }
            } finally {
                controllersRef.current.delete(item.id);
                runningCountRef.current -= 1;
                pumpRef.current();
            }
        },
        [patchItem, transferFile]
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
                        ? {
                              ...item,
                              status: 'queued',
                              sentBytes: 0,
                              progress: 0,
                              bytesPerSecond: undefined,
                              remainingSeconds: undefined,
                              error: undefined,
                              errorKey: undefined,
                          }
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

    const retryFailed = useCallback(() => requeue((item) => item.status === 'failed'), [requeue]);

    const clearFinished = useCallback(() => {
        commitItems(itemsRef.current.filter((item) => !isFinishedItem(item)));
    }, [commitItems]);

    const setConflictPolicy = useCallback((policy: UploadConflictChoice) => {
        conflictPolicyRef.current = policy;
        saveConflictPolicy(policy);
        setConflictPolicyState(policy);
    }, []);

    const resolveConflict = useCallback(
        (itemId: string, choice: ResolvedConflictChoice, applyToAll: boolean) => {
            const isResolvedTarget = (item: UploadItem): boolean =>
                item.id === itemId && isAwaitingDecision(item);
            const isCoveredByApplyToAll = (item: UploadItem): boolean =>
                applyToAll &&
                item.id !== itemId &&
                (isAwaitingDecision(item) ||
                    (item.status === 'queued' && item.onConflict === 'ask'));
            commitItems(
                itemsRef.current.map((item) =>
                    isResolvedTarget(item) || isCoveredByApplyToAll(item)
                        ? {
                              ...item,
                              onConflict: choice,
                              status: 'queued',
                              sentBytes: 0,
                              progress: 0,
                              error: undefined,
                              errorKey: undefined,
                          }
                        : item
                )
            );
            pump();
        },
        [commitItems, pump]
    );

    const dismissConflict = useCallback(
        (itemId: string) => {
            const item = itemsRef.current.find((candidate) => candidate.id === itemId);
            if (!item || !isAwaitingDecision(item)) return;
            patchItem(itemId, { status: 'canceled', error: undefined });
        },
        [patchItem]
    );

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
        resolveConflict,
        dismissConflict,
    };
};

export default useUploadQueue;

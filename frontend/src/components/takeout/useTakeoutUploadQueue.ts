import useI18n from '@/components/i18n/provider/i18nContext';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { useCallback, useRef, useState } from 'react';
import {
    createQueueEntry,
    findNextQueuedEntry,
    isQueueBusy,
    removeSettledEntries,
    updateQueueEntry,
    type TakeoutQueueEntry,
} from './takeoutUploadQueue';
import { createTransferRateTracker } from '@/shared/transfer/transferRateTracker';
import { cancelTakeoutUpload } from '@/service/takeout';
import { TakeoutConnectionLostError } from './takeoutConnectionLostError';
import { uploadTakeoutZipInChunks } from './uploadTakeoutZipInChunks';

const useTakeoutUploadQueue = () => {
    const { t } = useI18n();
    const [entries, setEntries] = useState<TakeoutQueueEntry[]>([]);
    const latestEntriesRef = useRef<TakeoutQueueEntry[]>([]);
    const nextEntryNumberRef = useRef(0);
    const isProcessingRef = useRef(false);
    const uploadSessionsRef = useRef(
        new Map<string, { controller: AbortController; uploadId: string | null }>()
    );

    const commitEntries = useCallback((nextEntries: TakeoutQueueEntry[]) => {
        latestEntriesRef.current = nextEntries;
        setEntries(nextEntries);
    }, []);

    const changeEntry = useCallback(
        (id: string, changes: Partial<Omit<TakeoutQueueEntry, 'id' | 'file'>>) => {
            commitEntries(updateQueueEntry(latestEntriesRef.current, id, changes));
        },
        [commitEntries]
    );

    const addFile = useCallback(
        (file: File) => {
            nextEntryNumberRef.current += 1;
            const entry = createQueueEntry(file, `takeout-upload-${nextEntryNumberRef.current}`);
            const entryToAdd =
                entry.state === 'rejected'
                    ? { ...entry, errorMessage: t('TAKEOUT_FILE_TYPE_ERROR') }
                    : entry;
            commitEntries([...latestEntriesRef.current, entryToAdd]);
        },
        [commitEntries, t]
    );

    const describeUploadFailure = useCallback(
        (uploadError: unknown): string => {
            if (uploadError instanceof TakeoutConnectionLostError) {
                return t('TAKEOUT_UPLOAD_CONNECTION_LOST');
            }
            return extractBackendErrorMessage(uploadError) ?? t('TAKEOUT_UPLOAD_ERROR');
        },
        [t]
    );

    const uploadEntry = useCallback(
        async (entry: TakeoutQueueEntry) => {
            const { file } = entry;
            const uploadSession = {
                controller: new AbortController(),
                uploadId: null as string | null,
            };
            uploadSessionsRef.current.set(entry.id, uploadSession);
            const rateTracker = createTransferRateTracker();
            try {
                changeEntry(entry.id, { state: 'uploading', uploadedBytes: 0, errorMessage: '' });
                let stateAfterReconnect: 'uploading' | 'completing' = 'uploading';
                const completeResult = await uploadTakeoutZipInChunks(
                    file,
                    {
                        onSessionStarted: (uploadId) => {
                            uploadSession.uploadId = uploadId;
                        },
                        onProgress: (uploadedBytes) =>
                            changeEntry(entry.id, {
                                uploadedBytes,
                                ...rateTracker.record(uploadedBytes, file.size),
                            }),
                        onWaitingConnection: (isWaiting) => {
                            if (!isWaiting) {
                                changeEntry(entry.id, { state: stateAfterReconnect });
                                return;
                            }
                            rateTracker.reset();
                            changeEntry(entry.id, {
                                state: 'waitingConnection',
                                bytesPerSecond: undefined,
                                remainingSeconds: undefined,
                            });
                        },
                        onCompleting: () => {
                            stateAfterReconnect = 'completing';
                            changeEntry(entry.id, { state: 'completing' });
                        },
                    },
                    uploadSession.controller.signal
                );
                changeEntry(entry.id, {
                    state: 'done',
                    importId: completeResult.import_id ?? null,
                });
            } catch (uploadError) {
                if (uploadSession.controller.signal.aborted) {
                    return;
                }
                changeEntry(entry.id, {
                    state: 'error',
                    errorMessage: describeUploadFailure(uploadError),
                });
            } finally {
                uploadSessionsRef.current.delete(entry.id);
            }
        },
        [changeEntry, describeUploadFailure]
    );

    const startQueue = useCallback(async () => {
        if (isProcessingRef.current) {
            return;
        }
        isProcessingRef.current = true;
        try {
            let nextEntry = findNextQueuedEntry(latestEntriesRef.current);
            while (nextEntry) {
                await uploadEntry(nextEntry);
                nextEntry = findNextQueuedEntry(latestEntriesRef.current);
            }
        } finally {
            isProcessingRef.current = false;
        }
    }, [uploadEntry]);

    const deleteReceivedPartsOf = useCallback(
        async (entryId: string, uploadId: string) => {
            try {
                await cancelTakeoutUpload(uploadId);
            } catch (cancelError) {
                changeEntry(entryId, {
                    errorMessage:
                        extractBackendErrorMessage(cancelError) ?? t('TAKEOUT_CANCEL_ERROR'),
                });
            }
        },
        [changeEntry, t]
    );

    const cancelEntry = useCallback(
        async (id: string) => {
            const entry = latestEntriesRef.current.find((candidate) => candidate.id === id);
            if (!entry) return;
            if (entry.state === 'queued') {
                changeEntry(id, { state: 'canceled' });
                return;
            }
            const uploadSession = uploadSessionsRef.current.get(id);
            if (
                !uploadSession ||
                (entry.state !== 'uploading' && entry.state !== 'waitingConnection')
            ) {
                return;
            }
            uploadSession.controller.abort();
            changeEntry(id, { state: 'canceled', errorMessage: '' });
            if (uploadSession.uploadId === null) return;
            await deleteReceivedPartsOf(id, uploadSession.uploadId);
        },
        [changeEntry, deleteReceivedPartsOf]
    );

    const clearSettled = useCallback(() => {
        commitEntries(removeSettledEntries(latestEntriesRef.current));
    }, [commitEntries]);

    return {
        entries,
        addFile,
        startQueue,
        clearSettled,
        cancelEntry,
        isBusy: isQueueBusy(entries),
        hasQueuedEntries: findNextQueuedEntry(entries) !== undefined,
    };
};

export default useTakeoutUploadQueue;

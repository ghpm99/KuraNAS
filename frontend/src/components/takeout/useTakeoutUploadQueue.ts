import useI18n from '@/components/i18n/provider/i18nContext';
import { completeTakeoutUpload, initTakeoutUpload, uploadTakeoutChunk } from '@/service/takeout';
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

const useTakeoutUploadQueue = () => {
    const { t } = useI18n();
    const [entries, setEntries] = useState<TakeoutQueueEntry[]>([]);
    const latestEntriesRef = useRef<TakeoutQueueEntry[]>([]);
    const nextEntryNumberRef = useRef(0);
    const isProcessingRef = useRef(false);

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

    const uploadEntry = useCallback(
        async (entry: TakeoutQueueEntry) => {
            const { file } = entry;
            try {
                changeEntry(entry.id, { state: 'uploading', uploadedBytes: 0, errorMessage: '' });
                const initResult = await initTakeoutUpload(file.name, file.size);
                const chunkSize = Math.max(1, initResult.chunk_size);

                let offset = 0;
                while (offset < file.size) {
                    const chunk = file.slice(offset, offset + chunkSize);
                    await uploadTakeoutChunk(initResult.upload_id, chunk, offset);
                    offset += chunk.size;
                    changeEntry(entry.id, { uploadedBytes: offset });
                }

                changeEntry(entry.id, { state: 'completing' });
                const completeResult = await completeTakeoutUpload(initResult.upload_id);
                changeEntry(entry.id, {
                    state: 'done',
                    importId: completeResult.import_id ?? null,
                });
            } catch (uploadError) {
                changeEntry(entry.id, {
                    state: 'error',
                    errorMessage:
                        extractBackendErrorMessage(uploadError) ?? t('TAKEOUT_UPLOAD_ERROR'),
                });
            }
        },
        [changeEntry, t]
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

    const clearSettled = useCallback(() => {
        commitEntries(removeSettledEntries(latestEntriesRef.current));
    }, [commitEntries]);

    return {
        entries,
        addFile,
        startQueue,
        clearSettled,
        isBusy: isQueueBusy(entries),
        hasQueuedEntries: findNextQueuedEntry(entries) !== undefined,
    };
};

export default useTakeoutUploadQueue;

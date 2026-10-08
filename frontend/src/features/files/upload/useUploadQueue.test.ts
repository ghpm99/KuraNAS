import { act, renderHook } from '@testing-library/react';
import type { UploadOutcome, UploadSingleFileParams } from '@/service/files';
import { conflictPolicyStorageKey } from './conflictPolicyPreference';
import useUploadQueue from './useUploadQueue';

type PendingUpload = {
    params: UploadSingleFileParams;
    resolve: (outcome: UploadOutcome) => void;
    reject: (error: unknown) => void;
};

const createControlledUploader = () => {
    const pendingUploads: PendingUpload[] = [];
    const uploadFile = jest.fn(
        (params: UploadSingleFileParams) =>
            new Promise<UploadOutcome>((resolve, reject) => {
                params.signal?.addEventListener('abort', () => reject(new Error('aborted')));
                pendingUploads.push({ params, resolve, reject });
            })
    );
    return { uploadFile, pendingUploads };
};

const buildEntries = (count: number) =>
    Array.from({ length: count }, (_, index) => ({ file: new File(['x'], `f${index}.txt`) }));

describe('useUploadQueue', () => {
    beforeEach(() => window.localStorage.clear());

    it('mounts with an empty queue and the default rename policy', () => {
        const { result } = renderHook(() => useUploadQueue());

        expect(result.current.items).toEqual([]);
        expect(result.current.conflictPolicy).toBe('rename');
    });

    it('runs at most three uploads at a time and starts the next as one finishes', async () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile }));

        act(() => result.current.enqueue(buildEntries(5), 9));

        expect(uploadFile).toHaveBeenCalledTimes(3);
        expect(result.current.items.map((item) => item.status)).toEqual([
            'uploading',
            'uploading',
            'uploading',
            'queued',
            'queued',
        ]);

        await act(async () => pendingUploads[0]!.resolve({ status: 'uploaded' }));

        expect(uploadFile).toHaveBeenCalledTimes(4);
        expect(result.current.items[0]!.status).toBe('done');
        expect(result.current.items[0]!.progress).toBe(100);
        expect(uploadFile.mock.calls[0]![0].targetFolderId).toBe(9);
    });

    it('forwards progress, relative path and the chosen conflict policy', () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile }));

        act(() => result.current.setConflictPolicy('skip'));
        act(() =>
            result.current.enqueue([{ file: new File(['x'], 'a.mp3'), relativePath: 'Album/a.mp3' }])
        );
        act(() => pendingUploads[0]!.params.onProgress?.(42));

        expect(uploadFile.mock.calls[0]![0]).toEqual(
            expect.objectContaining({ relativePath: 'Album/a.mp3', onConflict: 'skip' })
        );
        expect(result.current.items[0]!.progress).toBe(42);
        expect(result.current.items[0]!.displayName).toBe('Album/a.mp3');
        expect(window.localStorage.getItem(conflictPolicyStorageKey)).toBe('skip');
    });

    it('maps backend outcomes to item statuses and keeps the backend error verbatim', async () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile }));

        act(() => result.current.enqueue(buildEntries(3)));
        await act(async () => {
            pendingUploads[0]!.resolve({ status: 'skipped' });
            pendingUploads[1]!.resolve({ status: 'renamed', name: 'f1 (2).txt' });
            pendingUploads[2]!.resolve({ status: 'failed', error: 'Falha ao enviar arquivo' });
        });

        expect(result.current.items.map((item) => item.status)).toEqual([
            'skipped',
            'renamed',
            'failed',
        ]);
        expect(result.current.items[1]!.savedName).toBe('f1 (2).txt');
        expect(result.current.items[2]!.error).toBe('Falha ao enviar arquivo');
    });

    it('marks a rejected request as failed using the backend error message', async () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile }));

        act(() => result.current.enqueue(buildEntries(1)));
        await act(async () =>
            pendingUploads[0]!.reject({ response: { data: { error: 'Já existe arquivo' } } })
        );
        expect(result.current.items[0]).toEqual(
            expect.objectContaining({ status: 'failed', error: 'Já existe arquivo' })
        );

        act(() => result.current.enqueue(buildEntries(1)));
        await act(async () => pendingUploads[1]!.reject(new Error('network')));
        expect(result.current.items[1]!.status).toBe('failed');
        expect(result.current.items[1]!.error).toBeUndefined();
    });

    it('cancels a running upload through its abort signal and a queued one directly', async () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile, maxConcurrent: 1 }));

        act(() => result.current.enqueue(buildEntries(2)));
        const [runningItem, queuedItem] = result.current.items;

        act(() => result.current.cancel(queuedItem!.id));
        expect(result.current.items[1]!.status).toBe('canceled');

        await act(async () => result.current.cancel(runningItem!.id));
        expect(pendingUploads[0]!.params.signal?.aborted).toBe(true);
        expect(result.current.items[0]!.status).toBe('canceled');
        expect(uploadFile).toHaveBeenCalledTimes(1);
    });

    it('cancels everything with cancelAll', async () => {
        const { uploadFile } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile, maxConcurrent: 1 }));

        act(() => result.current.enqueue(buildEntries(3)));
        await act(async () => result.current.cancelAll());

        expect(result.current.items.map((item) => item.status)).toEqual([
            'canceled',
            'canceled',
            'canceled',
        ]);
    });

    it('ignores cancel for an unknown id', () => {
        const { result } = renderHook(() => useUploadQueue());
        expect(() => act(() => result.current.cancel('missing'))).not.toThrow();
    });

    it('retries a failed item and only failed ones with retryFailed', async () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile }));

        act(() => result.current.enqueue(buildEntries(3)));
        await act(async () => {
            pendingUploads[0]!.reject(new Error('boom'));
            pendingUploads[1]!.reject(new Error('boom'));
            pendingUploads[2]!.resolve({ status: 'uploaded' });
        });

        act(() => result.current.retry(result.current.items[0]!.id));
        expect(uploadFile).toHaveBeenCalledTimes(4);
        expect(result.current.items[0]!.status).toBe('uploading');
        expect(result.current.items[0]!.error).toBeUndefined();

        act(() => result.current.retry(result.current.items[2]!.id));
        expect(uploadFile).toHaveBeenCalledTimes(4);

        act(() => result.current.retryFailed());
        expect(uploadFile).toHaveBeenCalledTimes(5);
        expect(result.current.items[1]!.status).toBe('uploading');
    });

    it('clears only finished items', async () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ uploadFile, maxConcurrent: 1 }));

        act(() => result.current.enqueue(buildEntries(2)));
        await act(async () => pendingUploads[0]!.resolve({ status: 'uploaded' }));
        act(() => result.current.clearFinished());

        expect(result.current.items).toHaveLength(1);
        expect(result.current.items[0]!.displayName).toBe('f1.txt');
    });

    it('notifies each completed file and ignores empty enqueue', async () => {
        const { uploadFile, pendingUploads } = createControlledUploader();
        const onFileFinished = jest.fn();
        const { result } = renderHook(() => useUploadQueue({ uploadFile, onFileFinished }));

        act(() => result.current.enqueue([]));
        expect(uploadFile).not.toHaveBeenCalled();

        act(() => result.current.enqueue(buildEntries(1)));
        await act(async () => pendingUploads[0]!.resolve({ status: 'uploaded' }));

        expect(onFileFinished).toHaveBeenCalledTimes(1);
    });
});

import { act, renderHook } from '@testing-library/react';
import { conflictPolicyStorageKey } from './conflictPolicyPreference';
import type { FileTransferOutcome, FileTransferParams } from './fileTransferTypes';
import { UploadConnectionLostError } from './uploadConnectionLostError';
import useUploadQueue from './useUploadQueue';

type PendingUpload = {
    params: FileTransferParams;
    resolve: (outcome: FileTransferOutcome) => void;
    reject: (error: unknown) => void;
};

const createControlledUploader = () => {
    const pendingUploads: PendingUpload[] = [];
    const transferFile = jest.fn(
        (params: FileTransferParams) =>
            new Promise<FileTransferOutcome>((resolve, reject) => {
                params.signal?.addEventListener('abort', () => reject(new Error('aborted')));
                pendingUploads.push({ params, resolve, reject });
            })
    );
    return { transferFile, pendingUploads };
};

const buildEntries = (count: number, size = 1) =>
    Array.from({ length: count }, (_, index) => ({
        file: new File(['x'.repeat(size)], `f${index}.txt`),
    }));

describe('useUploadQueue', () => {
    beforeEach(() => window.localStorage.clear());

    it('mounts with an empty queue and the default ask policy', () => {
        const { result } = renderHook(() => useUploadQueue());

        expect(result.current.items).toEqual([]);
        expect(result.current.conflictPolicy).toBe('ask');
    });

    it('runs at most three uploads at a time and starts the next as one finishes', async () => {
        const { transferFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile }));

        act(() => result.current.enqueue(buildEntries(5), 9));

        expect(transferFile).toHaveBeenCalledTimes(3);
        expect(result.current.items.map((item) => item.status)).toEqual([
            'uploading',
            'uploading',
            'uploading',
            'queued',
            'queued',
        ]);

        await act(async () => pendingUploads[0]!.resolve({ status: 'uploaded' }));

        expect(transferFile).toHaveBeenCalledTimes(4);
        expect(result.current.items[0]!.status).toBe('done');
        expect(result.current.items[0]!.progress).toBe(100);
        expect(transferFile.mock.calls[0]![0].targetFolderId).toBe(9);
    });

    it('forwards progress, relative path and the chosen conflict policy', () => {
        const { transferFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile }));

        act(() => result.current.setConflictPolicy('skip'));
        act(() =>
            result.current.enqueue([
                { file: new File(['x'.repeat(200)], 'a.mp3'), relativePath: 'Album/a.mp3' },
            ])
        );
        act(() => pendingUploads[0]!.params.onSentBytes(84));

        expect(transferFile.mock.calls[0]![0]).toEqual(
            expect.objectContaining({ relativePath: 'Album/a.mp3', onConflict: 'skip' })
        );
        expect(result.current.items[0]!.progress).toBe(42);
        expect(result.current.items[0]!.sentBytes).toBe(84);
        expect(result.current.items[0]!.size).toBe(200);
        expect(result.current.items[0]!.displayName).toBe('Album/a.mp3');
        expect(window.localStorage.getItem(conflictPolicyStorageKey)).toBe('skip');
    });

    it('maps backend outcomes to item statuses and keeps the backend error verbatim', async () => {
        const { transferFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile }));

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
        const { transferFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile }));

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
        const { transferFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile, maxConcurrent: 1 }));

        act(() => result.current.enqueue(buildEntries(2)));
        const [runningItem, queuedItem] = result.current.items;

        act(() => result.current.cancel(queuedItem!.id));
        expect(result.current.items[1]!.status).toBe('canceled');

        await act(async () => result.current.cancel(runningItem!.id));
        expect(pendingUploads[0]!.params.signal?.aborted).toBe(true);
        expect(result.current.items[0]!.status).toBe('canceled');
        expect(transferFile).toHaveBeenCalledTimes(1);
    });

    it('cancels everything with cancelAll', async () => {
        const { transferFile } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile, maxConcurrent: 1 }));

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
        const { transferFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile }));

        act(() => result.current.enqueue(buildEntries(3)));
        await act(async () => {
            pendingUploads[0]!.reject(new Error('boom'));
            pendingUploads[1]!.reject(new Error('boom'));
            pendingUploads[2]!.resolve({ status: 'uploaded' });
        });

        act(() => result.current.retry(result.current.items[0]!.id));
        expect(transferFile).toHaveBeenCalledTimes(4);
        expect(result.current.items[0]!.status).toBe('uploading');
        expect(result.current.items[0]!.error).toBeUndefined();

        act(() => result.current.retry(result.current.items[2]!.id));
        expect(transferFile).toHaveBeenCalledTimes(4);

        act(() => result.current.retryFailed());
        expect(transferFile).toHaveBeenCalledTimes(5);
        expect(result.current.items[1]!.status).toBe('uploading');
    });

    it('clears only finished items', async () => {
        const { transferFile, pendingUploads } = createControlledUploader();
        const { result } = renderHook(() => useUploadQueue({ transferFile, maxConcurrent: 1 }));

        act(() => result.current.enqueue(buildEntries(2)));
        await act(async () => pendingUploads[0]!.resolve({ status: 'uploaded' }));
        act(() => result.current.clearFinished());

        expect(result.current.items).toHaveLength(1);
        expect(result.current.items[0]!.displayName).toBe('f1.txt');
    });

    it('notifies each completed file and ignores empty enqueue', async () => {
        const { transferFile, pendingUploads } = createControlledUploader();
        const onFileFinished = jest.fn();
        const { result } = renderHook(() => useUploadQueue({ transferFile, onFileFinished }));

        act(() => result.current.enqueue([]));
        expect(transferFile).not.toHaveBeenCalled();

        act(() => result.current.enqueue(buildEntries(1)));
        await act(async () => pendingUploads[0]!.resolve({ status: 'uploaded' }));

        expect(onFileFinished).toHaveBeenCalledTimes(1);
    });

    describe('large file transfer states', () => {
        afterEach(() => jest.restoreAllMocks());

        it('follows the transfer phases and stores the computed checksum', () => {
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1)));
            const { params } = pendingUploads[0]!;

            act(() => params.onPhaseChange('checking'));
            expect(result.current.items[0]!.status).toBe('checking');

            act(() => params.onChecksumComputed('abc123'));
            act(() => params.onPhaseChange('uploading'));
            expect(result.current.items[0]).toEqual(
                expect.objectContaining({ status: 'uploading', checksum: 'abc123' })
            );

            act(() => params.onPhaseChange('waitingConnection'));
            expect(result.current.items[0]!.status).toBe('waitingConnection');
        });

        it('maps the already existed outcome and treats it as finished and not retryable', async () => {
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1)));

            await act(async () => pendingUploads[0]!.resolve({ status: 'alreadyExisted' }));
            act(() => result.current.retry(result.current.items[0]!.id));

            expect(result.current.items[0]).toEqual(
                expect.objectContaining({ status: 'alreadyExisted', progress: 100, sentBytes: 1 })
            );
            expect(transferFile).toHaveBeenCalledTimes(1);
        });

        it('keeps the checksum on retry, reuses it and zeroes the transfer measures', async () => {
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1, 1000)));
            act(() => pendingUploads[0]!.params.onChecksumComputed('abc123'));
            act(() => pendingUploads[0]!.params.onSentBytes(500));
            await act(async () => pendingUploads[0]!.reject(new Error('boom')));

            expect(transferFile.mock.calls[0]![0].knownChecksum).toBeUndefined();
            expect(result.current.items[0]!.checksum).toBe('abc123');

            act(() => result.current.retry(result.current.items[0]!.id));

            expect(transferFile.mock.calls[1]![0].knownChecksum).toBe('abc123');
            expect(result.current.items[0]).toEqual(
                expect.objectContaining({
                    status: 'uploading',
                    sentBytes: 0,
                    progress: 0,
                    checksum: 'abc123',
                })
            );
        });

        it('counts items checking or waiting for connection against the concurrency limit', () => {
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(4)));

            act(() => pendingUploads[0]!.params.onPhaseChange('checking'));
            act(() => pendingUploads[1]!.params.onPhaseChange('waitingConnection'));

            expect(transferFile).toHaveBeenCalledTimes(3);
            expect(result.current.items.map((item) => item.status)).toEqual([
                'checking',
                'waitingConnection',
                'uploading',
                'queued',
            ]);
        });

        it('shows connection lost with a translation key instead of backend text', async () => {
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1)));

            await act(async () => pendingUploads[0]!.reject(new UploadConnectionLostError()));

            expect(result.current.items[0]).toEqual(
                expect.objectContaining({
                    status: 'failed',
                    errorKey: 'ERROR_UPLOAD_CONNECTION_LOST',
                    error: undefined,
                })
            );

            act(() => result.current.retry(result.current.items[0]!.id));
            expect(result.current.items[0]!.errorKey).toBeUndefined();
        });

        it('ignores sent bytes reported outside the uploading phase', () => {
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1, 100)));
            act(() => pendingUploads[0]!.params.onPhaseChange('waitingConnection'));

            act(() => pendingUploads[0]!.params.onSentBytes(50));

            expect(result.current.items[0]!.sentBytes).toBe(0);
        });

        it('patches progress at most every 250 ms unless the whole percentage changes', () => {
            let currentTime = 10_000;
            jest.spyOn(Date, 'now').mockImplementation(() => currentTime);
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1, 10_000)));
            const { onSentBytes } = pendingUploads[0]!.params;

            act(() => onSentBytes(10));
            const itemsAfterFirstPatch = result.current.items;
            currentTime += 100;
            act(() => onSentBytes(20));
            expect(result.current.items).toBe(itemsAfterFirstPatch);

            currentTime += 150;
            act(() => onSentBytes(30));
            expect(result.current.items[0]!.sentBytes).toBe(30);

            currentTime += 10;
            act(() => onSentBytes(200));
            expect(result.current.items[0]).toEqual(
                expect.objectContaining({ sentBytes: 200, progress: 2 })
            );
        });

        it('publishes speed and remaining time and clears them when waiting for connection', () => {
            let currentTime = 10_000;
            jest.spyOn(Date, 'now').mockImplementation(() => currentTime);
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1, 10_000)));
            const { params } = pendingUploads[0]!;

            act(() => params.onSentBytes(0));
            currentTime += 2000;
            act(() => params.onSentBytes(2000));

            expect(result.current.items[0]).toEqual(
                expect.objectContaining({ bytesPerSecond: 1000, remainingSeconds: 8 })
            );

            act(() => params.onPhaseChange('waitingConnection'));
            expect(result.current.items[0]!.bytesPerSecond).toBeUndefined();
            expect(result.current.items[0]!.remainingSeconds).toBeUndefined();
        });

        it('does not estimate a negative speed after the transfer realigns backwards', () => {
            let currentTime = 10_000;
            jest.spyOn(Date, 'now').mockImplementation(() => currentTime);
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));
            act(() => result.current.enqueue(buildEntries(1, 10_000)));
            const { params } = pendingUploads[0]!;

            act(() => params.onSentBytes(0));
            currentTime += 2000;
            act(() => params.onSentBytes(4000));
            currentTime += 2000;
            act(() => params.onSentBytes(1000));

            expect(result.current.items[0]!.bytesPerSecond).toBeUndefined();
        });
    });

    describe('name conflicts', () => {
        const enqueueAndConflict = async (count: number) => {
            const uploader = createControlledUploader();
            const hook = renderHook(() => useUploadQueue({ ...uploader, maxConcurrent: 3 }));
            act(() => hook.result.current.enqueue(buildEntries(count)));
            return { ...uploader, ...hook };
        };

        const conflict: FileTransferOutcome = {
            status: 'conflict',
            error: 'Já existe um arquivo com esse nome',
        };

        const statuses = (items: { status: string }[]) => items.map((item) => item.status);

        it('sends ask to the server as fail', () => {
            const { transferFile, pendingUploads } = createControlledUploader();
            const { result } = renderHook(() => useUploadQueue({ transferFile }));

            act(() => result.current.enqueue(buildEntries(1)));

            expect(pendingUploads[0]!.params.onConflict).toBe('fail');
            expect(result.current.items[0]!.onConflict).toBe('ask');
        });

        it('holds a conflicting item as awaiting decision keeping the server message', async () => {
            const { result, pendingUploads } = await enqueueAndConflict(1);

            await act(async () => pendingUploads[0]!.resolve(conflict));

            expect(result.current.items[0]).toEqual(
                expect.objectContaining({
                    status: 'awaitingDecision',
                    error: 'Já existe um arquivo com esse nome',
                })
            );
        });

        it('does not occupy a concurrency slot while awaiting a decision', async () => {
            const { result, pendingUploads, transferFile } = await enqueueAndConflict(4);

            await act(async () => pendingUploads[0]!.resolve(conflict));

            expect(statuses(result.current.items)).toEqual([
                'awaitingDecision',
                'uploading',
                'uploading',
                'uploading',
            ]);
            expect(transferFile).toHaveBeenCalledTimes(4);
        });

        it('keeps the awaiting item on clearFinished and out of the finished count', async () => {
            const { result, pendingUploads } = await enqueueAndConflict(2);
            await act(async () => {
                pendingUploads[0]!.resolve(conflict);
                pendingUploads[1]!.resolve({ status: 'uploaded' });
            });

            act(() => result.current.clearFinished());

            expect(statuses(result.current.items)).toEqual(['awaitingDecision']);
        });

        it('leaves the progress empty while awaiting the decision', async () => {
            const { result, pendingUploads } = await enqueueAndConflict(1);
            await act(async () => pendingUploads[0]!.resolve(conflict));

            expect(result.current.items[0]).toMatchObject({
                status: 'awaitingDecision',
                progress: 0,
                sentBytes: 0,
            });
        });

        it('is not retryable through the retry actions', async () => {
            const { result, pendingUploads, transferFile } = await enqueueAndConflict(1);
            await act(async () => pendingUploads[0]!.resolve(conflict));

            act(() => result.current.retry(result.current.items[0]!.id));
            act(() => result.current.retryFailed());

            expect(result.current.items[0]!.status).toBe('awaitingDecision');
            expect(transferFile).toHaveBeenCalledTimes(1);
        });

        it('resolves one conflict by requeueing it with the choice and the known checksum', async () => {
            const { result, pendingUploads, transferFile } = await enqueueAndConflict(1);
            act(() => pendingUploads[0]!.params.onChecksumComputed('abc123'));
            await act(async () => pendingUploads[0]!.resolve(conflict));

            act(() =>
                result.current.resolveConflict(result.current.items[0]!.id, 'replace', false)
            );

            expect(transferFile).toHaveBeenCalledTimes(2);
            expect(transferFile.mock.calls[1]![0]).toEqual(
                expect.objectContaining({ onConflict: 'replace', knownChecksum: 'abc123' })
            );
            expect(result.current.items[0]!.status).toBe('uploading');
            expect(result.current.items[0]!.error).toBeUndefined();
        });

        it('resolves only the chosen conflict without applyToAll', async () => {
            const { result, pendingUploads, transferFile } = await enqueueAndConflict(2);
            await act(async () => {
                pendingUploads[0]!.resolve(conflict);
                pendingUploads[1]!.resolve(conflict);
            });

            act(() => result.current.resolveConflict(result.current.items[0]!.id, 'skip', false));

            expect(statuses(result.current.items)).toEqual(['uploading', 'awaitingDecision']);
            expect(transferFile).toHaveBeenCalledTimes(3);
        });

        it('switches queued items still asking to the choice with applyToAll', async () => {
            const uploader = createControlledUploader();
            const { result } = renderHook(() =>
                useUploadQueue({ transferFile: uploader.transferFile, maxConcurrent: 1 })
            );
            act(() => result.current.enqueue(buildEntries(3)));
            await act(async () => uploader.pendingUploads[0]!.resolve(conflict));
            expect(statuses(result.current.items)).toEqual([
                'awaitingDecision',
                'uploading',
                'queued',
            ]);

            act(() => result.current.resolveConflict(result.current.items[0]!.id, 'rename', true));

            expect(result.current.items.map((item) => item.onConflict)).toEqual([
                'rename',
                'ask',
                'rename',
            ]);
            expect(result.current.items[0]!.status).toBe('queued');
        });

        it('resolves every awaiting conflict with applyToAll', async () => {
            const { result, pendingUploads, transferFile } = await enqueueAndConflict(3);
            await act(async () => {
                pendingUploads[0]!.resolve(conflict);
                pendingUploads[1]!.resolve(conflict);
            });

            act(() => result.current.resolveConflict(result.current.items[0]!.id, 'replace', true));

            expect(statuses(result.current.items)).toEqual(['uploading', 'uploading', 'uploading']);
            expect(transferFile.mock.calls.slice(3).map(([params]) => params.onConflict)).toEqual([
                'replace',
                'replace',
            ]);
        });

        it('does not touch items already done or running when applying to all', async () => {
            const { result, pendingUploads } = await enqueueAndConflict(3);
            await act(async () => {
                pendingUploads[0]!.resolve(conflict);
                pendingUploads[1]!.resolve({ status: 'uploaded' });
            });

            act(() => result.current.resolveConflict(result.current.items[0]!.id, 'skip', true));

            expect(result.current.items[1]).toEqual(
                expect.objectContaining({ status: 'done', onConflict: 'ask' })
            );
            expect(result.current.items[2]).toEqual(
                expect.objectContaining({ status: 'uploading', onConflict: 'ask' })
            );
        });

        it('cancels a dismissed conflict and ignores dismissing anything else', async () => {
            const { result, pendingUploads } = await enqueueAndConflict(2);
            await act(async () => pendingUploads[0]!.resolve(conflict));

            act(() => result.current.dismissConflict(result.current.items[0]!.id));
            act(() => result.current.dismissConflict(result.current.items[1]!.id));

            expect(statuses(result.current.items)).toEqual(['canceled', 'uploading']);
        });
    });
});

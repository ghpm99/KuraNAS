export type TakeoutQueueEntryState =
    | 'queued'
    | 'uploading'
    | 'completing'
    | 'done'
    | 'error'
    | 'rejected';

export type TakeoutQueueEntry = {
    id: string;
    file: File;
    state: TakeoutQueueEntryState;
    uploadedBytes: number;
    importId: number | null;
    errorMessage: string;
};

export const createQueueEntry = (file: File, id: string): TakeoutQueueEntry => ({
    id,
    file,
    state: file.name.toLowerCase().endsWith('.zip') ? 'queued' : 'rejected',
    uploadedBytes: 0,
    importId: null,
    errorMessage: '',
});

export const updateQueueEntry = (
    entries: TakeoutQueueEntry[],
    id: string,
    changes: Partial<Omit<TakeoutQueueEntry, 'id' | 'file'>>
): TakeoutQueueEntry[] =>
    entries.map((entry) => (entry.id === id ? { ...entry, ...changes } : entry));

export const findNextQueuedEntry = (entries: TakeoutQueueEntry[]): TakeoutQueueEntry | undefined =>
    entries.find((entry) => entry.state === 'queued');

export const isQueueBusy = (entries: TakeoutQueueEntry[]): boolean =>
    entries.some((entry) => entry.state === 'uploading' || entry.state === 'completing');

export const removeSettledEntries = (entries: TakeoutQueueEntry[]): TakeoutQueueEntry[] =>
    entries.filter((entry) => entry.state === 'uploading' || entry.state === 'completing');

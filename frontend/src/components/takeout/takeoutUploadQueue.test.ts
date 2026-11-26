import {
    createQueueEntry,
    findNextQueuedEntry,
    isQueueBusy,
    removeSettledEntries,
    updateQueueEntry,
    type TakeoutQueueEntry,
} from './takeoutUploadQueue';

const zipFile = (name: string) => new File(['x'], name);

const entryWithState = (id: string, state: TakeoutQueueEntry['state']): TakeoutQueueEntry => ({
    ...createQueueEntry(zipFile(`${id}.zip`), id),
    state,
});

describe('takeoutUploadQueue', () => {
    it('queues a zip and rejects another format', () => {
        expect(createQueueEntry(zipFile('a.zip'), '1').state).toBe('queued');
        const rejectedEntry = createQueueEntry(zipFile('a.rar'), '2');
        expect(rejectedEntry.state).toBe('rejected');
        expect(rejectedEntry.errorMessage).toBe('');
    });

    it('accepts the zip extension in upper case', () => {
        expect(createQueueEntry(zipFile('TAKEOUT.ZIP'), '1').state).toBe('queued');
    });

    it('updates only the entry with the given id', () => {
        const entries = [entryWithState('a', 'queued'), entryWithState('b', 'queued')];
        const updated = updateQueueEntry(entries, 'b', { state: 'uploading', uploadedBytes: 3 });
        expect(updated[0]).toBe(entries[0]);
        expect(updated[1]).toMatchObject({ id: 'b', state: 'uploading', uploadedBytes: 3 });
    });

    it('finds the first queued entry in order', () => {
        const entries = [
            entryWithState('a', 'done'),
            entryWithState('b', 'queued'),
            entryWithState('c', 'queued'),
        ];
        expect(findNextQueuedEntry(entries)?.id).toBe('b');
        expect(findNextQueuedEntry([entryWithState('a', 'done')])).toBeUndefined();
    });

    it('is busy only while uploading or completing', () => {
        expect(isQueueBusy([entryWithState('a', 'queued'), entryWithState('b', 'done')])).toBe(
            false
        );
        expect(isQueueBusy([entryWithState('a', 'uploading')])).toBe(true);
        expect(isQueueBusy([entryWithState('a', 'completing')])).toBe(true);
    });

    it('removes settled entries and keeps the one being sent', () => {
        const entries = [
            entryWithState('a', 'done'),
            entryWithState('b', 'uploading'),
            entryWithState('c', 'error'),
            entryWithState('d', 'rejected'),
            entryWithState('e', 'queued'),
        ];
        expect(removeSettledEntries(entries).map((entry) => entry.id)).toEqual(['b']);
    });
});

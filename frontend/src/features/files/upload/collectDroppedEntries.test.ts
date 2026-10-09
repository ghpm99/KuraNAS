import { collectDroppedEntries } from './collectDroppedEntries';

type FakeEntry = {
    isFile: boolean;
    isDirectory: boolean;
    name: string;
    fullPath: string;
    file?: (onSuccess: (file: File) => void, onError: (error: unknown) => void) => void;
    createReader?: () => {
        readEntries: (onSuccess: (entries: FakeEntry[]) => void, onError: (error: unknown) => void) => void;
    };
};

const fileEntry = (fullPath: string): FakeEntry => ({
    isFile: true,
    isDirectory: false,
    name: fullPath.split('/').pop()!,
    fullPath,
    file: (onSuccess) => onSuccess(new File(['x'], fullPath.split('/').pop()!)),
});

const directoryEntry = (fullPath: string, children: FakeEntry[], batchSize = 100): FakeEntry => ({
    isFile: false,
    isDirectory: true,
    name: fullPath.split('/').pop()!,
    fullPath,
    createReader: () => {
        let cursor = 0;
        return {
            readEntries: (onSuccess) => {
                const batch = children.slice(cursor, cursor + batchSize);
                cursor += batchSize;
                onSuccess(batch);
            },
        };
    },
});

const buildDataTransfer = (entries: Array<FakeEntry | null>, files: File[] = []): DataTransfer =>
    ({
        items: entries.map((entry) => ({ kind: 'file', webkitGetAsEntry: () => entry })),
        files,
    }) as unknown as DataTransfer;

describe('collectDroppedEntries', () => {
    it('walks dropped directories recursively keeping relative paths', async () => {
        const album = directoryEntry('/Album', [
            fileEntry('/Album/cover.jpg'),
            directoryEntry('/Album/Disc1', [fileEntry('/Album/Disc1/a.mp3')]),
        ]);

        const entries = await collectDroppedEntries(buildDataTransfer([album]));

        expect(entries.map((entry) => [entry.file.name, entry.relativePath])).toEqual([
            ['cover.jpg', 'Album/cover.jpg'],
            ['a.mp3', 'Album/Disc1/a.mp3'],
        ]);
    });

    it('sends top-level dropped files without a relative path', async () => {
        const entries = await collectDroppedEntries(buildDataTransfer([fileEntry('/solo.txt')]));

        expect(entries).toHaveLength(1);
        expect(entries[0]!.relativePath).toBeUndefined();
    });

    it('reads directories in several batches until an empty batch', async () => {
        const children = ['a', 'b', 'c'].map((name) => fileEntry(`/Dir/${name}.txt`));
        const entries = await collectDroppedEntries(
            buildDataTransfer([directoryEntry('/Dir', children, 2)])
        );

        expect(entries.map((entry) => entry.relativePath)).toEqual([
            'Dir/a.txt',
            'Dir/b.txt',
            'Dir/c.txt',
        ]);
    });

    it('mixes files and folders in one drop', async () => {
        const entries = await collectDroppedEntries(
            buildDataTransfer([
                fileEntry('/one.txt'),
                directoryEntry('/Dir', [fileEntry('/Dir/two.txt')]),
            ])
        );

        expect(entries.map((entry) => entry.relativePath)).toEqual([undefined, 'Dir/two.txt']);
    });

    it('falls back to the plain file list when entries are unavailable', async () => {
        const plainFile = new File(['x'], 'plain.txt');

        const entries = await collectDroppedEntries(buildDataTransfer([null], [plainFile]));

        expect(entries).toEqual([{ file: plainFile, relativePath: undefined }]);
    });

    it('falls back to files when the drop has no items', async () => {
        const plainFile = new File(['x'], 'plain.txt');
        const dataTransfer = { files: [plainFile] } as unknown as DataTransfer;

        await expect(collectDroppedEntries(dataTransfer)).resolves.toEqual([
            { file: plainFile, relativePath: undefined },
        ]);
    });

    it('ignores entries that are neither file nor directory', async () => {
        const odd: FakeEntry = { isFile: false, isDirectory: false, name: 'x', fullPath: '/x' };

        await expect(collectDroppedEntries(buildDataTransfer([odd]))).resolves.toEqual([]);
    });

    it('skips non-file items', async () => {
        const dataTransfer = {
            items: [{ kind: 'string' }],
            files: [],
        } as unknown as DataTransfer;

        await expect(collectDroppedEntries(dataTransfer)).resolves.toEqual([]);
    });
});

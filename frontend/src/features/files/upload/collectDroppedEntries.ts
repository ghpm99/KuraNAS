import { entriesFromFileList } from './entriesFromFileList';
import type { UploadEntry } from './uploadQueueTypes';

type FileSystemEntryLike = {
    isFile: boolean;
    isDirectory: boolean;
    name: string;
    fullPath: string;
    file?: (onSuccess: (file: File) => void, onError: (error: unknown) => void) => void;
    createReader?: () => {
        readEntries: (
            onSuccess: (entries: FileSystemEntryLike[]) => void,
            onError: (error: unknown) => void
        ) => void;
    };
};

type DataTransferItemWithEntry = DataTransferItem & {
    webkitGetAsEntry?: () => FileSystemEntryLike | null;
};

const stripLeadingSlash = (fullPath: string): string => fullPath.replace(/^\/+/, '');

const readFile = (entry: FileSystemEntryLike): Promise<File> =>
    new Promise((resolve, reject) => {
        entry.file?.(resolve, reject);
    });

const readDirectoryBatch = (
    reader: NonNullable<ReturnType<NonNullable<FileSystemEntryLike['createReader']>>>
): Promise<FileSystemEntryLike[]> =>
    new Promise((resolve, reject) => {
        reader.readEntries(resolve, reject);
    });

const readAllDirectoryEntries = async (
    directory: FileSystemEntryLike
): Promise<FileSystemEntryLike[]> => {
    const reader = directory.createReader?.();
    if (!reader) return [];
    const collected: FileSystemEntryLike[] = [];
    let batch = await readDirectoryBatch(reader);
    while (batch.length > 0) {
        collected.push(...batch);
        batch = await readDirectoryBatch(reader);
    }
    return collected;
};

const walkEntry = async (entry: FileSystemEntryLike, isTopLevel: boolean): Promise<UploadEntry[]> => {
    if (entry.isFile) {
        const file = await readFile(entry);
        return [{ file, relativePath: isTopLevel ? undefined : stripLeadingSlash(entry.fullPath) }];
    }
    if (!entry.isDirectory) return [];
    const children = await readAllDirectoryEntries(entry);
    const nested = await Promise.all(children.map((child) => walkEntry(child, false)));
    return nested.flat();
};

export const collectDroppedEntries = async (dataTransfer: DataTransfer): Promise<UploadEntry[]> => {
    const rootEntries = Array.from(dataTransfer.items ?? [])
        .filter((item) => item.kind === 'file')
        .map((item) => (item as DataTransferItemWithEntry).webkitGetAsEntry?.() ?? null);

    if (rootEntries.length === 0 || rootEntries.some((entry) => entry === null)) {
        return entriesFromFileList(dataTransfer.files ?? []);
    }

    const walked = await Promise.all(
        (rootEntries as FileSystemEntryLike[]).map((entry) => walkEntry(entry, true))
    );
    return walked.flat();
};

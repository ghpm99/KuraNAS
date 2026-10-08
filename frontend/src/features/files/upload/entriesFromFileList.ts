import type { UploadEntry } from './uploadQueueTypes';

type FileWithRelativePath = File & { webkitRelativePath?: string };

export const entriesFromFileList = (fileList: FileList | File[]): UploadEntry[] =>
    Array.from(fileList).map((file) => ({
        file,
        relativePath: (file as FileWithRelativePath).webkitRelativePath || undefined,
    }));

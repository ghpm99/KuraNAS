import type { FileContextType } from '@/features/files/providers/fileProvider/fileContext';

export const createFileContextStub = (overrides: Partial<FileContextType> = {}): FileContextType =>
    ({
        selectedItem: null,
        files: [],
        moveFile: jest.fn().mockResolvedValue(undefined),
        copyFile: jest.fn().mockResolvedValue(undefined),
        deleteFile: jest.fn().mockResolvedValue(undefined),
        renameFile: jest.fn().mockResolvedValue(undefined),
        toggleStarred: jest.fn().mockResolvedValue(undefined),
        ...overrides,
    }) as unknown as FileContextType;

import type { FileContextType } from '@/features/files/providers/fileProvider/fileContext';

export const createFileContextStub = (overrides: Partial<FileContextType> = {}): FileContextType =>
    ({
        selectedItem: null,
        files: [],
        moveFile: jest.fn().mockResolvedValue(undefined),
        copyFile: jest.fn().mockResolvedValue(undefined),
        deleteFile: jest.fn().mockResolvedValue(undefined),
        renameFile: jest.fn().mockResolvedValue(undefined),
        promoteFileToHot: jest.fn().mockResolvedValue(undefined),
        setFilesStarred: jest.fn().mockResolvedValue([]),
        ...overrides,
    }) as unknown as FileContextType;

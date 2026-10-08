import { act, renderHook } from '@testing-library/react';
import { FileType } from '@/utils';
import { UploadQueueContextProvider } from './uploadQueueContext';
import type { UploadQueue } from './uploadQueueTypes';
import useUploadToCurrentFolder, { resolveUploadFolderId } from './useUploadToCurrentFolder';

const mockUseFile = jest.fn();
const mockGetFileByPath = jest.fn();
const mockEnqueueSnackbar = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));
jest.mock('@/service/files', () => ({
    getFileByPath: (path: string) => mockGetFileByPath(path),
}));
jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

const directory = { id: 10, type: FileType.Directory, parent_path: '/media' };
const fileInFolder = { id: 20, type: FileType.File, parent_path: '/media/photos' };
const fileAtRoot = { id: 21, type: FileType.File, parent_path: '/' };

describe('resolveUploadFolderId', () => {
    beforeEach(() => jest.clearAllMocks());

    it('uses the root when nothing is open', async () => {
        await expect(resolveUploadFolderId(null)).resolves.toBeUndefined();
    });

    it('uses the opened folder itself', async () => {
        await expect(resolveUploadFolderId(directory as never)).resolves.toBe(10);
    });

    it('uses the parent folder when a file is the current item', async () => {
        mockGetFileByPath.mockResolvedValue({ id: 77 });

        await expect(resolveUploadFolderId(fileInFolder as never)).resolves.toBe(77);
        expect(mockGetFileByPath).toHaveBeenCalledWith('/media/photos');
    });

    it('uses the root for a file that lives at the root', async () => {
        await expect(resolveUploadFolderId(fileAtRoot as never)).resolves.toBeUndefined();
        expect(mockGetFileByPath).not.toHaveBeenCalled();
    });

    it('fails when the parent folder cannot be found', async () => {
        mockGetFileByPath.mockResolvedValue(null);

        await expect(resolveUploadFolderId(fileInFolder as never)).rejects.toThrow();
    });
});

describe('useUploadToCurrentFolder', () => {
    const enqueue = jest.fn();
    const wrapper = ({ children }: { children: React.ReactNode }) => (
        <UploadQueueContextProvider value={{ enqueue } as unknown as UploadQueue}>
            {children}
        </UploadQueueContextProvider>
    );
    const entries = [{ file: new File(['x'], 'a.txt') }];

    beforeEach(() => jest.clearAllMocks());

    it('enqueues into the folder being viewed', async () => {
        mockUseFile.mockReturnValue({ selectedItem: directory });
        const { result } = renderHook(() => useUploadToCurrentFolder(), { wrapper });

        await act(async () => result.current.uploadEntries(entries));

        expect(enqueue).toHaveBeenCalledWith(entries, 10);
    });

    it('enqueues into the parent folder when a file is open', async () => {
        mockUseFile.mockReturnValue({ selectedItem: fileInFolder });
        mockGetFileByPath.mockResolvedValue({ id: 77 });
        const { result } = renderHook(() => useUploadToCurrentFolder(), { wrapper });

        await act(async () => result.current.uploadEntries(entries));

        expect(enqueue).toHaveBeenCalledWith(entries, 77);
    });

    it('reports an error when the destination cannot be resolved', async () => {
        mockUseFile.mockReturnValue({ selectedItem: fileInFolder });
        mockGetFileByPath.mockRejectedValue(new Error('network'));
        const { result } = renderHook(() => useUploadToCurrentFolder(), { wrapper });

        await act(async () => result.current.uploadEntries(entries));

        expect(enqueue).not.toHaveBeenCalled();
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_UPLOAD_FAILED', { variant: 'error' });
    });

    it('does nothing for an empty drop', async () => {
        mockUseFile.mockReturnValue({ selectedItem: directory });
        const { result } = renderHook(() => useUploadToCurrentFolder(), { wrapper });

        await act(async () => result.current.uploadEntries([]));

        expect(enqueue).not.toHaveBeenCalled();
    });
});

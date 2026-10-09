import { act, renderHook } from '@testing-library/react';
import { createTestFile } from '@/features/files/selection/testFileFactory';
import useFileOperations from './useFileOperations';

const mockEnqueueSnackbar = jest.fn();
const mockTriggerBrowserDownload = jest.fn();
const mockFileContext = {
    moveFile: jest.fn(),
    copyFile: jest.fn(),
    deleteFile: jest.fn(),
    renameFile: jest.fn(),
    toggleStarred: jest.fn(),
};

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));
jest.mock('@/service/browserDownload', () => ({
    triggerBrowserDownload: (...args: unknown[]) => mockTriggerBrowserDownload(...args),
}));
jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockFileContext,
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options ? `${key}:${JSON.stringify(options)}` : key,
    }),
}));

const files = [1, 2, 3].map((id) => createTestFile(id));

describe('useFileOperations', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        Object.values(mockFileContext).forEach((operation) => operation.mockResolvedValue(undefined));
    });

    it('reports one success summary with the count when every delete works', async () => {
        const { result } = renderHook(() => useFileOperations());

        await act(async () => {
            await result.current.deleteFiles(files);
        });

        expect(mockFileContext.deleteFile).toHaveBeenCalledTimes(3);
        expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1);
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
            'FILES_BULK_DELETE_SUCCESS:{"count":"3"}',
            { variant: 'success' }
        );
    });

    it('passes the permanent flag to every delete', async () => {
        const { result } = renderHook(() => useFileOperations());

        await act(async () => {
            await result.current.deleteFiles(files, true);
        });

        expect(mockFileContext.deleteFile).toHaveBeenCalledWith(files[0]!.id, true);
        expect(mockFileContext.deleteFile).toHaveBeenCalledTimes(3);
    });

    it('uses the single-item message for one file', async () => {
        const { result } = renderHook(() => useFileOperations());

        await act(async () => {
            await result.current.deleteFiles([files[0]!]);
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ACTION_DELETE_SUCCESS', {
            variant: 'success',
        });
    });

    it('summarizes a bulk delete with one failure using the backend message verbatim', async () => {
        mockFileContext.deleteFile.mockImplementation(async (fileId: number) => {
            if (fileId === 2) throw { response: { data: { error: 'Pasta protegida' } } };
        });
        const { result } = renderHook(() => useFileOperations());

        let outcome: Awaited<ReturnType<typeof result.current.deleteFiles>> | undefined;
        await act(async () => {
            outcome = await result.current.deleteFiles(files);
        });

        expect(outcome?.failedFiles.map((file) => file.id)).toEqual([2]);
        expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1);
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
            'FILES_BULK_PARTIAL_SUMMARY:{"succeeded":"2","failed":"1","message":"Pasta protegida"}',
            { variant: 'warning' }
        );
    });

    it('reports a single failure with the backend message or the fallback key', async () => {
        const { result } = renderHook(() => useFileOperations());

        mockFileContext.moveFile.mockRejectedValueOnce({ response: { data: { error: 'Destino inválido' } } });
        await act(async () => {
            await result.current.moveFiles([files[0]!], { folderId: 9 });
        });
        expect(mockEnqueueSnackbar).toHaveBeenLastCalledWith('Destino inválido', { variant: 'error' });
        expect(mockFileContext.moveFile).toHaveBeenCalledWith(1, 9, undefined);

        mockFileContext.copyFile.mockRejectedValueOnce(new Error('network'));
        await act(async () => {
            await result.current.copyFiles([files[0]!], { path: '/dest' });
        });
        expect(mockEnqueueSnackbar).toHaveBeenLastCalledWith('ERROR_COPY_FAILED', { variant: 'error' });
        expect(mockFileContext.copyFile).toHaveBeenCalledWith(1, undefined, '/dest');
    });

    it('reports an all-failed summary with the failure count', async () => {
        mockFileContext.moveFile.mockRejectedValue({ response: { data: { error: 'Sem permissão' } } });
        const { result } = renderHook(() => useFileOperations());

        await act(async () => {
            await result.current.moveFiles(files, { folderId: 9 });
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
            'FILES_BULK_FAILED_SUMMARY:{"failed":"3","message":"Sem permissão"}',
            { variant: 'error' }
        );
    });

    it('renames one file', async () => {
        const { result } = renderHook(() => useFileOperations());

        await act(async () => {
            await result.current.renameSingleFile(files[0]!, 'novo.txt');
        });

        expect(mockFileContext.renameFile).toHaveBeenCalledWith(1, 'novo.txt');
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ACTION_RENAME_SUCCESS', { variant: 'success' });
    });

    it('favorites only unstarred files, and unfavorites when all are starred', async () => {
        const { result } = renderHook(() => useFileOperations());
        const mixedFiles = [createTestFile(1, { starred: true }), createTestFile(2), createTestFile(3)];

        await act(async () => {
            await result.current.toggleFavorites(mixedFiles);
        });
        expect(mockFileContext.toggleStarred.mock.calls.map(([fileId]) => fileId)).toEqual([2, 3]);

        mockFileContext.toggleStarred.mockClear();
        await act(async () => {
            await result.current.toggleFavorites([mixedFiles[0]!]);
        });
        expect(mockFileContext.toggleStarred).toHaveBeenCalledWith(1);
    });

    it('downloads one file by its url and several as a zip', () => {
        const { result } = renderHook(() => useFileOperations());

        result.current.downloadFiles([]);
        expect(mockTriggerBrowserDownload).not.toHaveBeenCalled();

        result.current.downloadFiles([files[0]!]);
        expect(mockTriggerBrowserDownload).toHaveBeenLastCalledWith(
            expect.stringContaining('/files/download/1'),
            'file-1.txt'
        );

        result.current.downloadFiles(files);
        expect(mockTriggerBrowserDownload).toHaveBeenLastCalledWith(
            expect.stringContaining('1,2,3')
        );
    });

    it('copies paths when the clipboard is available and reports when it is not', async () => {
        const { result } = renderHook(() => useFileOperations());

        Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true });
        await act(async () => {
            await result.current.copyPaths(files);
        });
        expect(mockEnqueueSnackbar).toHaveBeenLastCalledWith('ERROR_COPY_PATH_FAILED', {
            variant: 'error',
        });

        const writeText = jest.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
        await act(async () => {
            await result.current.copyPaths(files);
        });
        expect(writeText).toHaveBeenCalledWith(
            '/library/file-1.txt\n/library/file-2.txt\n/library/file-3.txt'
        );
        expect(mockEnqueueSnackbar).toHaveBeenLastCalledWith('FILES_COPY_PATH_SUCCESS', {
            variant: 'success',
        });

        writeText.mockRejectedValueOnce(new Error('denied'));
        await act(async () => {
            await result.current.copyPaths(files);
        });
        expect(mockEnqueueSnackbar).toHaveBeenLastCalledWith('ERROR_COPY_PATH_FAILED', {
            variant: 'error',
        });
    });
});

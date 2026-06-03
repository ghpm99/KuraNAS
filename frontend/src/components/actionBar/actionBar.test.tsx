import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import ActionBar from './actionBar';
import { FileType } from '@/utils';

const mockUseFile = jest.fn();
const mockNavigate = jest.fn();
const mockEnqueueSnackbar = jest.fn();
const mockDownloadFileBlob = jest.fn();

// Mock FolderPicker — renders a minimal dialog with confirm/cancel buttons.
// The confirm button calls onSelect with a result stored in mockFolderPickerResult.
let mockFolderPickerResult = { folderId: 99 };
jest.mock('@/components/folderPicker/folderPicker', () => ({
    __esModule: true,
    default: ({ open, onClose, onSelect }: { open: boolean; onClose: () => void; onSelect: (r: any) => void }) => {
        if (!open) return null;
        return (
            <div role="dialog" aria-label="FOLDER_PICKER">
                <button onClick={() => onSelect(mockFolderPickerResult)}>CONFIRM_PICKER</button>
                <button onClick={onClose}>CANCEL_PICKER</button>
            </div>
        );
    },
}));

const createFileContext = (overrides = {}) => ({
    selectedItem: null,
    uploadFiles: jest.fn(),
    createFolder: jest.fn().mockResolvedValue(undefined),
    moveFile: jest.fn().mockResolvedValue(undefined),
    copyFile: jest.fn().mockResolvedValue(undefined),
    renameFile: jest.fn().mockResolvedValue(undefined),
    deleteFile: jest.fn().mockResolvedValue(undefined),
    rescanFiles: jest.fn(),
    fileListFilter: 'recent',
    ...overrides,
});

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('../i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => {
            const map: Record<string, string> = {
                FILES: 'FILES',
                RECENT_FILES: 'RECENT_FILES',
                STARRED_FILES: 'STARRED_FILES',
                NEW_FILE: 'NEW_FILE',
                UPLOAD_FILE: 'UPLOAD_FILE',
                NEW_FOLDER: 'NEW_FOLDER',
                MOVE: 'MOVE',
                COPY: 'COPY',
                RENAME: 'RENAME',
                DELETE: 'DELETE',
                DOWNLOAD: 'DOWNLOAD',
                NAME: 'NAME',
                PATH: 'PATH',
                ACTION_CANCEL: 'ACTION_CANCEL',
                CONFIRM_DELETE: 'CONFIRM_DELETE',
                ACTION_CREATE_FOLDER_SUCCESS: 'ACTION_CREATE_FOLDER_SUCCESS',
                ACTION_MOVE_SUCCESS: 'ACTION_MOVE_SUCCESS',
                ACTION_COPY_SUCCESS: 'ACTION_COPY_SUCCESS',
                ACTION_RENAME_SUCCESS: 'ACTION_RENAME_SUCCESS',
                ACTION_DELETE_SUCCESS: 'ACTION_DELETE_SUCCESS',
                ACTION_UPLOAD_SUCCESS: 'ACTION_UPLOAD_SUCCESS',
                ERROR_LOADING_FILES: 'ERROR_LOADING_FILES',
                ERROR_UPLOAD_FAILED: 'ERROR_UPLOAD_FAILED',
                ERROR_CREATE_FOLDER_FAILED: 'ERROR_CREATE_FOLDER_FAILED',
                ERROR_MOVE_FAILED: 'ERROR_MOVE_FAILED',
                ERROR_COPY_FAILED: 'ERROR_COPY_FAILED',
                ERROR_RENAME_FAILED: 'ERROR_RENAME_FAILED',
                ERROR_DELETE_FAILED: 'ERROR_DELETE_FAILED',
                COPY_SUFFIX: '_copy',
            };
            return map[key] ?? key;
        },
    }),
}));

jest.mock('react-router-dom', () => {
    const actual = jest.requireActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

jest.mock('notistack', () => ({
    useSnackbar: () => ({
        enqueueSnackbar: mockEnqueueSnackbar,
    }),
}));

jest.mock('@/service/files', () => ({
    downloadFileBlob: (...args: unknown[]) => mockDownloadFileBlob(...args),
}));

describe('components/actionBar', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockDownloadFileBlob.mockResolvedValue(new Blob(['file']));
        mockUseFile.mockReturnValue(createFileContext());
    });

    it('shows the filtered list title and creates folders from the dialog', async () => {
        const createFolder = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(createFileContext({ createFolder }));

        render(<ActionBar />);

        expect(screen.getByText('RECENT_FILES')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'MOVE' })).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));

        const dialog = screen.getByRole('dialog');
        fireEvent.change(within(dialog).getByLabelText('NAME'), {
            target: { value: 'Docs' },
        });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'NEW_FOLDER' })[0]!);

        await waitFor(() => {
            expect(createFolder).toHaveBeenCalledWith('Docs', undefined);
        });
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ACTION_CREATE_FOLDER_SUCCESS', {
            variant: 'success',
        });
    });

    it('opens move/copy/rename/delete flows and downloads the selected file', async () => {
        const moveFile = jest.fn().mockResolvedValue(undefined);
        const copyFile = jest.fn().mockResolvedValue(undefined);
        const renameFile = jest.fn().mockResolvedValue(undefined);
        const deleteFile = jest.fn().mockResolvedValue(undefined);
        Object.assign(URL, {
            createObjectURL: URL.createObjectURL ?? jest.fn(),
            revokeObjectURL: URL.revokeObjectURL ?? jest.fn(),
        });
        const createObjectURLSpy = jest.spyOn(URL, 'createObjectURL').mockReturnValue('blob:url');
        const revokeObjectURLSpy = jest
            .spyOn(URL, 'revokeObjectURL')
            .mockImplementation(() => undefined);
        const clickSpy = jest.fn();
        const removeSpy = jest.fn();
        const originalCreateElement = document.createElement.bind(document);
        const createElementSpy = jest
            .spyOn(document, 'createElement')
            .mockImplementation((tagName: string) => {
                if (tagName === 'a') {
                    const anchor = originalCreateElement('a');
                    anchor.click = clickSpy;
                    anchor.remove = removeSpy;
                    return anchor;
                }

                return originalCreateElement(tagName);
            });

        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
                fileListFilter: 'all',
                moveFile,
                copyFile,
                renameFile,
                deleteFile,
            })
        );

        render(<ActionBar />);

        // Move via FolderPicker
        mockFolderPickerResult = { folderId: 20 };
        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        fireEvent.click(screen.getByRole('button', { name: 'CONFIRM_PICKER' }));
        await waitFor(() => {
            expect(moveFile).toHaveBeenCalledWith(7, 20, undefined);
        });
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'FOLDER_PICKER' })).not.toBeInTheDocument();
        });

        // Copy via FolderPicker
        mockFolderPickerResult = { folderId: 30 };
        fireEvent.click(screen.getByRole('button', { name: 'COPY' }));
        fireEvent.click(screen.getByRole('button', { name: 'CONFIRM_PICKER' }));
        await waitFor(() => {
            expect(copyFile).toHaveBeenCalledWith(7, 30, undefined);
        });
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'FOLDER_PICKER' })).not.toBeInTheDocument();
        });

        // Rename
        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
        let dialog = screen.getByRole('dialog');
        fireEvent.change(within(dialog).getByLabelText('NAME'), {
            target: { value: 'movie-new.mp4' },
        });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'RENAME' })[0]!);
        await waitFor(() => {
            expect(renameFile).toHaveBeenCalledWith(7, 'movie-new.mp4');
        });
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'RENAME' })).not.toBeInTheDocument();
        });

        // Delete
        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        dialog = screen.getByRole('dialog');
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'DELETE' })[0]!);
        await waitFor(() => {
            expect(deleteFile).toHaveBeenCalledWith(7);
        });
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'DELETE' })).not.toBeInTheDocument();
        });

        // Download
        fireEvent.click(screen.getByRole('button', { name: 'DOWNLOAD' }));
        await waitFor(() => {
            expect(mockDownloadFileBlob).toHaveBeenCalledWith(7);
        });

        expect(createObjectURLSpy).toHaveBeenCalled();
        expect(clickSpy).toHaveBeenCalled();
        expect(removeSpy).toHaveBeenCalled();
        expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:url');

        createElementSpy.mockRestore();
        createObjectURLSpy.mockRestore();
        revokeObjectURLSpy.mockRestore();
    });

    it('shows error snackbars when operations fail', async () => {
        const error = new Error('boom');
        const context = createFileContext({
            selectedItem: {
                id: 7,
                name: 'movie.mp4',
                path: '/media/movie.mp4',
                parent_path: '/media',
                type: FileType.File,
            },
            createFolder: jest.fn().mockRejectedValue(error),
            moveFile: jest.fn().mockRejectedValue(error),
            copyFile: jest.fn().mockRejectedValue(error),
            renameFile: jest.fn().mockRejectedValue(error),
            deleteFile: jest.fn().mockRejectedValue(error),
        });

        mockUseFile.mockReturnValue(context);

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));
        let dialog = screen.getByRole('dialog', { name: 'NEW_FOLDER' });
        fireEvent.change(within(dialog).getByLabelText('NAME'), {
            target: { value: 'Docs' },
        });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'NEW_FOLDER' })[0]!);
        await waitFor(() => expect(context.createFolder).toHaveBeenCalled());
        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_CREATE_FOLDER_FAILED', {
                variant: 'error',
            })
        );
        fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
        await waitFor(() =>
            expect(screen.queryByRole('dialog', { name: 'NEW_FOLDER' })).not.toBeInTheDocument()
        );

        mockFolderPickerResult = { folderId: 10 };
        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        fireEvent.click(screen.getByRole('button', { name: 'CONFIRM_PICKER' }));
        await waitFor(() => expect(context.moveFile).toHaveBeenCalled());
        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_MOVE_FAILED', {
                variant: 'error',
            })
        );

        fireEvent.click(screen.getByRole('button', { name: 'COPY' }));
        fireEvent.click(screen.getByRole('button', { name: 'CONFIRM_PICKER' }));
        await waitFor(() => expect(context.copyFile).toHaveBeenCalled());
        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_COPY_FAILED', {
                variant: 'error',
            })
        );

        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
        dialog = screen.getByRole('dialog', { name: 'RENAME' });
        fireEvent.change(within(dialog).getByLabelText('NAME'), {
            target: { value: 'movie-new.mp4' },
        });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'RENAME' })[0]!);
        await waitFor(() => expect(context.renameFile).toHaveBeenCalled());
        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_RENAME_FAILED', {
                variant: 'error',
            })
        );
        fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
        await waitFor(() =>
            expect(screen.queryByRole('dialog', { name: 'RENAME' })).not.toBeInTheDocument()
        );

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        dialog = screen.getByRole('dialog', { name: 'DELETE' });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'DELETE' })[0]!);
        await waitFor(() => expect(context.deleteFile).toHaveBeenCalled());
        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_DELETE_FAILED', {
                variant: 'error',
            })
        );
        fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
        await waitFor(() =>
            expect(screen.queryByRole('dialog', { name: 'DELETE' })).not.toBeInTheDocument()
        );
    });

    it('shows upload errors and resets the file input', async () => {
        const uploadFiles = jest.fn().mockRejectedValue(new Error('upload failed'));
        mockUseFile.mockReturnValue(createFileContext({ uploadFiles }));

        render(<ActionBar />);

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
        const blob = new File(['content'], 'doc.txt', { type: 'text/plain' });
        fireEvent.change(fileInput, { target: { files: [blob] } });

        await waitFor(() => {
            expect(uploadFiles).toHaveBeenCalledWith([blob], undefined);
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_UPLOAD_FAILED', {
            variant: 'error',
        });
        expect(fileInput.value).toBe('');
    });

    it('shows STARRED_FILES title when fileListFilter is starred', () => {
        mockUseFile.mockReturnValue(createFileContext({ fileListFilter: 'starred' }));
        render(<ActionBar />);
        expect(screen.getByText('STARRED_FILES')).toBeInTheDocument();
    });

    it('shows FILES title when fileListFilter is all', () => {
        mockUseFile.mockReturnValue(createFileContext({ fileListFilter: 'all' }));
        render(<ActionBar />);
        expect(screen.getByText('FILES')).toBeInTheDocument();
    });

    it('uses directory id as currentFolderId when selectedItem is a directory', async () => {
        const createFolder = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 10,
                    name: 'photos',
                    path: '/media/photos',
                    parent_path: '/media',
                    type: FileType.Directory,
                },
                createFolder,
            })
        );

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));
        const dialog = screen.getByRole('dialog');
        fireEvent.change(within(dialog).getByLabelText('NAME'), {
            target: { value: 'Sub' },
        });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'NEW_FOLDER' })[0]!);

        await waitFor(() => {
            expect(createFolder).toHaveBeenCalledWith('Sub', 10);
        });
    });

    it('navigates to parent path when back button is clicked', () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 10,
                    name: 'photos',
                    path: '/media/photos',
                    parent_path: '/media',
                    type: FileType.Directory,
                },
            })
        );

        render(<ActionBar />);

        const backButton = screen.getAllByRole('button')[0]!;
        fireEvent.click(backButton);

        expect(mockNavigate).toHaveBeenCalledWith('/files/media');
    });

    it('navigates to files root when parent_path is /', () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 10,
                    name: 'root-item',
                    path: '/root-item',
                    parent_path: '/',
                    type: FileType.Directory,
                },
            })
        );

        render(<ActionBar />);

        const backButton = screen.getAllByRole('button')[0]!;
        fireEvent.click(backButton);

        expect(mockNavigate).toHaveBeenCalledWith('/files');
    });

    it('navigates to files root when parent_path is empty', () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 10,
                    name: 'root-item',
                    path: '/root-item',
                    parent_path: '',
                    type: FileType.Directory,
                },
            })
        );

        render(<ActionBar />);

        const backButton = screen.getAllByRole('button')[0]!;
        fireEvent.click(backButton);

        expect(mockNavigate).toHaveBeenCalledWith('/files');
    });

    it('does not show download button for directory items', () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 10,
                    name: 'photos',
                    path: '/media/photos',
                    parent_path: '/media',
                    type: FileType.Directory,
                },
            })
        );

        render(<ActionBar />);

        expect(screen.queryByRole('button', { name: 'DOWNLOAD' })).not.toBeInTheDocument();
    });

    it('shows successful upload snackbar', async () => {
        const uploadFiles = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(createFileContext({ uploadFiles }));

        render(<ActionBar />);

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
        const blob = new File(['content'], 'doc.txt', { type: 'text/plain' });
        fireEvent.change(fileInput, { target: { files: [blob] } });

        await waitFor(() => {
            expect(uploadFiles).toHaveBeenCalled();
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ACTION_UPLOAD_SUCCESS', {
            variant: 'success',
        });
    });

    it('ignores upload when no files are selected', async () => {
        const uploadFiles = jest.fn();
        mockUseFile.mockReturnValue(createFileContext({ uploadFiles }));

        render(<ActionBar />);

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
        fireEvent.change(fileInput, { target: { files: null } });

        expect(uploadFiles).not.toHaveBeenCalled();
    });

    it('ignores upload when file list is empty', async () => {
        const uploadFiles = jest.fn();
        mockUseFile.mockReturnValue(createFileContext({ uploadFiles }));

        render(<ActionBar />);

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
        fireEvent.change(fileInput, { target: { files: [] } });

        expect(uploadFiles).not.toHaveBeenCalled();
    });

    it('does not create folder when name is empty', async () => {
        const createFolder = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(createFileContext({ createFolder }));

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));
        const dialog = screen.getByRole('dialog');
        // Leave folder name empty (default) and click create
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'NEW_FOLDER' })[0]!);

        expect(createFolder).not.toHaveBeenCalled();
    });

    it('does not rename when new name is same as current name', async () => {
        const renameFile = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
                renameFile,
            })
        );

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
        const dialog = screen.getByRole('dialog');
        // Name is pre-filled with 'movie.mp4', don't change it
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'RENAME' })[0]!);

        expect(renameFile).not.toHaveBeenCalled();
    });

    it('shows download error snackbar when download fails', async () => {
        mockDownloadFileBlob.mockRejectedValue(new Error('download failed'));
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
            })
        );

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'DOWNLOAD' }));

        await waitFor(() => {
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_LOADING_FILES', {
                variant: 'error',
            });
        });
    });

    it('uses undefined as currentFolderId when selectedItem is a file', async () => {
        const createFolder = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
                createFolder,
            })
        );

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));
        const dialog = screen.getByRole('dialog');
        fireEvent.change(within(dialog).getByLabelText('NAME'), {
            target: { value: 'Sub' },
        });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'NEW_FOLDER' })[0]!);

        await waitFor(() => {
            expect(createFolder).toHaveBeenCalledWith('Sub', undefined);
        });
    });

    it('displays selected item name instead of filter title', () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 10,
                    name: 'photos',
                    path: '/media/photos',
                    parent_path: '/media',
                    type: FileType.Directory,
                },
            })
        );

        render(<ActionBar />);
        expect(screen.getByText('photos')).toBeInTheDocument();
    });

    it('closes move picker via cancel', async () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
            })
        );

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        expect(screen.getByRole('dialog', { name: 'FOLDER_PICKER' })).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'CANCEL_PICKER' }));
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'FOLDER_PICKER' })).not.toBeInTheDocument();
        });
    });

    it('closes copy picker via cancel', async () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
            })
        );

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'COPY' }));
        expect(screen.getByRole('dialog', { name: 'FOLDER_PICKER' })).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'CANCEL_PICKER' }));
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'FOLDER_PICKER' })).not.toBeInTheDocument();
        });
    });

    it('does not rename when new name is empty', async () => {
        const renameFile = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
                renameFile,
            })
        );

        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
        const dialog = screen.getByRole('dialog');
        fireEvent.change(within(dialog).getByLabelText('NAME'), {
            target: { value: '   ' },
        });
        fireEvent.click(within(dialog).getAllByRole('button', { name: 'RENAME' })[0]!);

        expect(renameFile).not.toHaveBeenCalled();
    });

    it('closes create folder dialog via Escape key', async () => {
        mockUseFile.mockReturnValue(createFileContext());
        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));
        expect(screen.getByRole('dialog')).toBeInTheDocument();

        fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'NEW_FOLDER' })).not.toBeInTheDocument();
        });
    });

    it('opens move folder picker dialog', () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
            })
        );
        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        expect(screen.getByRole('dialog', { name: 'FOLDER_PICKER' })).toBeInTheDocument();
    });

    it('opens copy folder picker dialog', () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
            })
        );
        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'COPY' }));
        expect(screen.getByRole('dialog', { name: 'FOLDER_PICKER' })).toBeInTheDocument();
    });

    it('closes rename dialog via Escape key', async () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
            })
        );
        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
        expect(screen.getByRole('dialog')).toBeInTheDocument();

        fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'RENAME' })).not.toBeInTheDocument();
        });
    });

    it('closes delete dialog via Escape key', async () => {
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 7,
                    name: 'movie.mp4',
                    path: '/media/movie.mp4',
                    parent_path: '/media',
                    type: FileType.File,
                },
            })
        );
        render(<ActionBar />);

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        expect(screen.getByRole('dialog')).toBeInTheDocument();

        fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: 'DELETE' })).not.toBeInTheDocument();
        });
    });

    it('uploads files with folder id when selectedItem is a directory', async () => {
        const uploadFiles = jest.fn().mockResolvedValue(undefined);
        mockUseFile.mockReturnValue(
            createFileContext({
                selectedItem: {
                    id: 10,
                    name: 'photos',
                    path: '/media/photos',
                    parent_path: '/media',
                    type: FileType.Directory,
                },
                uploadFiles,
            })
        );

        render(<ActionBar />);

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
        const blob = new File(['content'], 'doc.txt', { type: 'text/plain' });
        fireEvent.change(fileInput, { target: { files: [blob] } });

        await waitFor(() => {
            expect(uploadFiles).toHaveBeenCalledWith([blob], 10);
        });
    });
});

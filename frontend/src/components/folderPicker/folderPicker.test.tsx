import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import FolderPicker from './folderPicker';

const mockGetFilesTree = jest.fn();

jest.mock('@/service/files', () => ({
    getFilesTree: (...args: unknown[]) => mockGetFilesTree(...args),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => {
            const map: Record<string, string> = {
                SELECT_DESTINATION: 'SELECT_DESTINATION',
                ROOT_FOLDER: 'ROOT_FOLDER',
                PATH: 'PATH',
                ACTION_CANCEL: 'ACTION_CANCEL',
                MOVE: 'MOVE',
                EMPTY_FILE_LIST: 'EMPTY_FILE_LIST',
            };
            return map[key] ?? key;
        },
    }),
}));

const makePaginationResponse = (items: Array<{ id: number; name: string; path: string; type: number }>) => ({
    items: items.map((item) => ({
        id: item.id,
        name: item.name,
        path: item.path,
        parent_path: '/',
        type: item.type,
        format: '',
        size: 0,
        updated_at: '',
        created_at: '',
        deleted_at: '',
        last_interaction: '',
        last_backup: '',
        check_sum: '',
        directory_content_count: 0,
        starred: false,
    })),
    pagination: { hasNext: false, hasPrevious: false, page: 1, pageSize: 200 },
});

describe('FolderPicker', () => {
    const onClose = jest.fn();
    const onSelect = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));
    });

    it('renders dialog with title when open', async () => {
        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);
        expect(screen.getByText('SELECT_DESTINATION')).toBeInTheDocument();
        expect(screen.getByText('ROOT_FOLDER')).toBeInTheDocument();
    });

    it('does not render dialog when closed', () => {
        render(<FolderPicker open={false} onClose={onClose} onSelect={onSelect} />);
        expect(screen.queryByText('SELECT_DESTINATION')).not.toBeInTheDocument();
    });

    it('fetches and displays folders', async () => {
        mockGetFilesTree.mockResolvedValue(
            makePaginationResponse([
                { id: 1, name: 'Documents', path: '/Documents', type: 1 },
                { id: 2, name: 'Photos', path: '/Photos', type: 1 },
                { id: 3, name: 'readme.txt', path: '/readme.txt', type: 2 },
            ])
        );

        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        await waitFor(() => {
            expect(screen.getByText('Documents')).toBeInTheDocument();
            expect(screen.getByText('Photos')).toBeInTheDocument();
        });
        // Files should be filtered out
        expect(screen.queryByText('readme.txt')).not.toBeInTheDocument();
    });

    it('navigates into a folder and updates breadcrumbs', async () => {
        mockGetFilesTree
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 1, name: 'Documents', path: '/Documents', type: 1 },
                ])
            )
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 10, name: 'Sub', path: '/Documents/Sub', type: 1 },
                ])
            );

        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        await waitFor(() => expect(screen.getByText('Documents')).toBeInTheDocument());
        fireEvent.click(screen.getByText('Documents'));

        await waitFor(() => {
            expect(screen.getByText('Sub')).toBeInTheDocument();
        });
        // Breadcrumb should show Documents
        expect(screen.getByText('Documents')).toBeInTheDocument();
        expect(screen.getByLabelText('PATH')).toHaveValue('/Documents');
    });

    it('navigates back via root breadcrumb', async () => {
        mockGetFilesTree
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 1, name: 'Docs', path: '/Docs', type: 1 },
                ])
            )
            .mockResolvedValueOnce(makePaginationResponse([]))
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 1, name: 'Docs', path: '/Docs', type: 1 },
                ])
            );

        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);
        await waitFor(() => expect(screen.getByText('Docs')).toBeInTheDocument());

        fireEvent.click(screen.getByText('Docs'));
        await waitFor(() => expect(screen.getByText('EMPTY_FILE_LIST')).toBeInTheDocument());

        fireEvent.click(screen.getByText('ROOT_FOLDER'));
        await waitFor(() => expect(screen.getByText('Docs')).toBeInTheDocument());
        expect(screen.getByLabelText('PATH')).toHaveValue('');
    });

    it('selects folder by id when a folder was navigated into', async () => {
        mockGetFilesTree
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 5, name: 'Target', path: '/Target', type: 1 },
                ])
            )
            .mockResolvedValueOnce(makePaginationResponse([]));

        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);
        await waitFor(() => expect(screen.getByText('Target')).toBeInTheDocument());

        fireEvent.click(screen.getByText('Target'));
        await waitFor(() => expect(screen.getByLabelText('PATH')).toHaveValue('/Target'));

        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        expect(onSelect).toHaveBeenCalledWith({ folderId: 5 });
    });

    it('selects by path when user types a custom path', async () => {
        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        await waitFor(() => expect(screen.getByLabelText('PATH')).toBeInTheDocument());

        fireEvent.change(screen.getByLabelText('PATH'), {
            target: { value: '/new/custom/path' },
        });

        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        expect(onSelect).toHaveBeenCalledWith({ path: '/new/custom/path' });
    });

    it('selects empty result when path is empty and no folder selected', async () => {
        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        await waitFor(() => expect(screen.getByLabelText('PATH')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        expect(onSelect).toHaveBeenCalledWith({});
    });

    it('calls onClose when cancel is clicked', async () => {
        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));
        expect(onClose).toHaveBeenCalled();
    });

    it('clears selected folder when user edits the path input', async () => {
        mockGetFilesTree
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 5, name: 'Target', path: '/Target', type: 1 },
                ])
            )
            .mockResolvedValueOnce(makePaginationResponse([]));

        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);
        await waitFor(() => expect(screen.getByText('Target')).toBeInTheDocument());

        fireEvent.click(screen.getByText('Target'));
        await waitFor(() => expect(screen.getByLabelText('PATH')).toHaveValue('/Target'));

        // Modify the path manually
        fireEvent.change(screen.getByLabelText('PATH'), {
            target: { value: '/Target/new-sub' },
        });

        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        expect(onSelect).toHaveBeenCalledWith({ path: '/Target/new-sub' });
    });

    it('shows empty state when no folders exist', async () => {
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));

        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        await waitFor(() => {
            expect(screen.getByText('EMPTY_FILE_LIST')).toBeInTheDocument();
        });
    });

    it('handles API error gracefully', async () => {
        mockGetFilesTree.mockRejectedValue(new Error('API error'));

        render(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        await waitFor(() => {
            expect(screen.getByText('EMPTY_FILE_LIST')).toBeInTheDocument();
        });
    });

    it('resets state when reopened', async () => {
        mockGetFilesTree
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 1, name: 'Docs', path: '/Docs', type: 1 },
                ])
            )
            .mockResolvedValueOnce(makePaginationResponse([]))
            .mockResolvedValueOnce(
                makePaginationResponse([
                    { id: 1, name: 'Docs', path: '/Docs', type: 1 },
                ])
            );

        const { rerender } = render(
            <FolderPicker open onClose={onClose} onSelect={onSelect} />
        );

        await waitFor(() => expect(screen.getByText('Docs')).toBeInTheDocument());
        fireEvent.click(screen.getByText('Docs'));
        await waitFor(() => expect(screen.getByLabelText('PATH')).toHaveValue('/Docs'));

        // Close and reopen
        rerender(<FolderPicker open={false} onClose={onClose} onSelect={onSelect} />);
        rerender(<FolderPicker open onClose={onClose} onSelect={onSelect} />);

        await waitFor(() => {
            expect(screen.getByLabelText('PATH')).toHaveValue('');
        });
    });
});

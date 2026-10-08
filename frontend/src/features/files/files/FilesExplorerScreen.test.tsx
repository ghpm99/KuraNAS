import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import FilesExplorerScreen from './FilesExplorerScreen';

const mockUseFile = jest.fn();
const mockSetFilesSort = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

jest.mock('@/components/actionBar', () => () => <div>ActionBarMock</div>);
jest.mock('@/features/files/fileContent', () => ({ viewMode, showHeading }: any) => (
    <div
        data-testid="file-content"
        data-view-mode={viewMode}
        data-show-heading={String(showHeading)}
    >
        FileContentMock
    </div>
));
jest.mock('@/features/files/fileDetails', () => () => <div>FileDetailsMock</div>);
jest.mock('@/components/layout/Sidebar/components/folderTree', () => () => (
    <div>FolderTreeMock</div>
));
jest.mock('@/components/tabs', () => () => <div>TabsMock</div>);
jest.mock(
    '@/features/files/findByDiskPath/findByDiskPathDialog',
    () =>
        ({ open, onFileFound }: any) =>
            open ? (
                <button onClick={() => onFileFound({ id: 9, path: '/found' })}>
                    FindDialogMock
                </button>
            ) : null
);

describe('FilesExplorerScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseFile.mockReturnValue({
            files: [
                {
                    id: 1,
                    name: 'media',
                    path: '/media',
                    parent_path: '/',
                    type: 1,
                    file_children: [
                        {
                            id: 2,
                            name: 'movie.mp4',
                            path: '/media/movie.mp4',
                            parent_path: '/media',
                            type: 2,
                            format: '.mp4',
                            size: 42,
                        },
                    ],
                },
            ],
            selectedItem: null,
            handleSelectItem: jest.fn(),
            fileListFilter: 'all',
            filesSort: { key: 'name', order: 'asc' },
            setFilesSort: mockSetFilesSort,
        });
    });

    it('renders explorer structure and switches view mode', () => {
        render(
            <MemoryRouter>
                <FilesExplorerScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('FILES_PAGE_TITLE')).toBeInTheDocument();
        expect(screen.getByText('FILES_PAGE_DESCRIPTION')).toBeInTheDocument();
        expect(screen.getByText('FILES_PAGE_TITLE')).toBeInTheDocument();
        expect(screen.getByTestId('file-content')).toHaveAttribute('data-view-mode', 'grid');
        expect(screen.getByTestId('file-content')).toHaveAttribute('data-show-heading', 'false');

        fireEvent.click(screen.getByRole('button', { name: 'FILES_VIEW_LIST' }));
        expect(screen.getByTestId('file-content')).toHaveAttribute('data-view-mode', 'list');
    });

    it('renders the sort control and forwards order toggles to the provider', () => {
        render(
            <MemoryRouter>
                <FilesExplorerScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SORT_ORDER_ASCENDING' }));

        expect(mockSetFilesSort).toHaveBeenCalledWith({ key: 'name', order: 'desc' });
    });

    it('renders breadcrumb and preview when a file is selected', () => {
        mockUseFile.mockReturnValue({
            ...mockUseFile(),
            selectedItem: {
                id: 2,
                name: 'movie.mp4',
                path: '/media/movie.mp4',
                parent_path: '/media',
                type: 2,
                format: '.mp4',
                size: 42,
            },
        });

        render(
            <MemoryRouter>
                <FilesExplorerScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('media')).toBeInTheDocument();
        expect(screen.getAllByText('movie.mp4').length).toBeGreaterThan(0);
        expect(screen.getByText('FileDetailsMock')).toBeInTheDocument();
    });

    it('opens the tree drawer action', () => {
        render(
            <MemoryRouter>
                <FilesExplorerScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('button', { name: 'FILES_OPEN_TREE' }));
        expect(screen.getByText('FolderTreeMock')).toBeInTheDocument();
    });

    it('opens the find-by-disk-path dialog and navigates to the found file', () => {
        const handleSelectItem = jest.fn();
        mockUseFile.mockReturnValue({ ...mockUseFile(), handleSelectItem });
        render(
            <MemoryRouter>
                <FilesExplorerScreen />
            </MemoryRouter>
        );

        expect(screen.queryByText('FindDialogMock')).toBeNull();
        fireEvent.click(screen.getByRole('button', { name: 'FILES_FIND_BY_DISK_PATH' }));
        fireEvent.click(screen.getByText('FindDialogMock'));

        expect(handleSelectItem).toHaveBeenCalledWith({ id: 9, path: '/found' });
    });
});

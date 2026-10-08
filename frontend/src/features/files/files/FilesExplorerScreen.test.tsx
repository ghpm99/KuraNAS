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

jest.mock('@/features/files/providers/fileProvider/useFileAncestors', () => ({
    __esModule: true,
    default: () => ({ data: undefined }),
}));

jest.mock('@/features/files/search/useFileSearchResults', () => ({
    __esModule: true,
    default: () => ({
        items: [],
        status: 'pending',
        hasNextPage: false,
        isFetchingNextPage: false,
        fetchNextPage: jest.fn(),
    }),
}));

jest.mock('@/components/actionBar', () => () => {
    const useFileDetails = jest.requireActual(
        '@/features/files/fileDetails/useFileDetails'
    ).default;
    const OpenFolderDetails = () => {
        const { openDetails } = useFileDetails();
        return (
            <button
                onClick={() =>
                    openDetails({ id: 1, name: 'media', path: '/media', parent_path: '/', type: 1 })
                }
            >
                ActionBarMock
            </button>
        );
    };
    return <OpenFolderDetails />;
});
jest.mock('@/features/files/fileContent', () => ({ viewMode, showHeading }: any) => (
    <div
        data-testid="file-content"
        data-view-mode={viewMode}
        data-show-heading={String(showHeading)}
    >
        FileContentMock
    </div>
));
jest.mock('@/features/files/fileDetails', () => ({ file, onClose }: any) => (
    <div>
        FileDetailsMock:{file.name}
        <button onClick={onClose}>CloseDetailsMock</button>
    </div>
));
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
        expect(screen.getByText(/FileDetailsMock:movie.mp4/)).toBeInTheDocument();
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

    describe('details panel', () => {
        const openedFile = {
            id: 2,
            name: 'movie.mp4',
            path: '/media/movie.mp4',
            parent_path: '/media',
            type: 2,
            format: '.mp4',
            size: 42,
        };

        afterEach(() => {
            delete (window as any).matchMedia;
        });

        const mockPhoneViewport = (isPhone: boolean) => {
            (window as any).matchMedia = (query: string) => ({
                matches: isPhone,
                media: query,
                addEventListener: jest.fn(),
                removeEventListener: jest.fn(),
                addListener: jest.fn(),
                removeListener: jest.fn(),
            });
        };

        it('leaves the opened file by closing the auto-opened side panel', () => {
            const handleSelectItem = jest.fn();
            mockUseFile.mockReturnValue({
                ...mockUseFile(),
                selectedItem: openedFile,
                handleSelectItem,
            });
            render(
                <MemoryRouter>
                    <FilesExplorerScreen />
                </MemoryRouter>
            );

            fireEvent.click(screen.getByText('CloseDetailsMock'));

            expect(handleSelectItem).toHaveBeenCalledWith(null);
        });

        it('shows the details of a folder in the side column and closes it without navigating', () => {
            const handleSelectItem = jest.fn();
            mockUseFile.mockReturnValue({ ...mockUseFile(), handleSelectItem });
            render(
                <MemoryRouter>
                    <FilesExplorerScreen />
                </MemoryRouter>
            );
            expect(screen.queryByText(/FileDetailsMock/)).toBeNull();

            fireEvent.click(screen.getByText('ActionBarMock'));
            expect(screen.getByText(/FileDetailsMock:media/)).toBeInTheDocument();

            fireEvent.click(screen.getByText('CloseDetailsMock'));
            expect(screen.queryByText(/FileDetailsMock/)).toBeNull();
            expect(handleSelectItem).not.toHaveBeenCalled();
        });

        it('opens the details of a folder as a bottom drawer on phones', () => {
            mockPhoneViewport(true);
            render(
                <MemoryRouter>
                    <FilesExplorerScreen />
                </MemoryRouter>
            );

            fireEvent.click(screen.getByText('ActionBarMock'));

            const drawer = screen.getByRole('presentation');
            expect(drawer.querySelector('.MuiDrawer-paperAnchorBottom')).not.toBeNull();
            expect(screen.getByText(/FileDetailsMock:media/)).toBeInTheDocument();
        });

        it('does not auto-open the details on phones for an opened file', () => {
            mockPhoneViewport(true);
            mockUseFile.mockReturnValue({ ...mockUseFile(), selectedItem: openedFile });

            render(
                <MemoryRouter>
                    <FilesExplorerScreen />
                </MemoryRouter>
            );

            expect(screen.queryByText(/FileDetailsMock/)).toBeNull();
        });
    });
});

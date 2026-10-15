import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import FavoritesScreen from './FavoritesScreen';

const mockUseFavoritesScreen = jest.fn();
const mockSetActiveFilter = jest.fn();
const mockSetViewMode = jest.fn();
const mockSetStarredSort = jest.fn();

jest.mock('./useFavoritesScreen', () => ({
    __esModule: true,
    default: () => mockUseFavoritesScreen(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

jest.mock('@/features/files/fileContent', () => ({ viewMode, title, emptyStateMessage }: any) => (
    <div
        data-testid="favorites-file-content"
        data-view-mode={viewMode}
        data-title={title}
        data-empty-state={emptyStateMessage}
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

describe('FavoritesScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseFavoritesScreen.mockReturnValue({
            activeFilter: 'all',
            activeFilterLabel: 'FAVORITES_FILTER_ALL',
            breadcrumbSegments: [
                { id: null, label: 'STARRED_FILES', path: null, isCurrent: false },
                { id: 10, label: 'Projects', path: '/projects', isCurrent: true },
            ],
            contextPath: '/projects',
            currentTitle: 'Projects',
            filterOptions: [
                { value: 'all', label: 'FAVORITES_FILTER_ALL', count: 3 },
                { value: 'folders', label: 'FAVORITES_FILTER_FOLDERS', count: 1 },
                { value: 'files', label: 'FAVORITES_FILTER_FILES', count: 1 },
                { value: 'media', label: 'FAVORITES_FILTER_MEDIA', count: 1 },
            ],
            listedItems: [{ id: 1, name: 'Invoice.pdf' }],
            itemCountLabel: '3 ITENS',
            selectedItem: null,
            isRootListing: true,
            setActiveFilter: mockSetActiveFilter,
            setStarredSort: mockSetStarredSort,
            setViewMode: mockSetViewMode,
            sortOptions: [
                { value: 'name', label: 'FAVORITES_SORT_NAME' },
                { value: 'starred_at', label: 'FAVORITES_SORT_STARRED_AT' },
            ],
            starredSort: 'name',
            viewMode: 'grid',
        });
    });

    it('loads the next page through the sentinel at the top of the favorites list', () => {
        const fetchNextPage = jest.fn();
        mockUseFavoritesScreen.mockReturnValue({
            ...mockUseFavoritesScreen(),
            selectedItem: null,
            hasNextPage: true,
            isFetchingNextPage: false,
            fetchNextPage,
        });

        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));
        expect(fetchNextPage).toHaveBeenCalledTimes(1);
    });

    it('keeps the sentinel while a folder is open so large folders paginate', () => {
        mockUseFavoritesScreen.mockReturnValue({
            ...mockUseFavoritesScreen(),
            selectedItem: { id: 10, type: 1, name: 'Projects' },
            hasNextPage: true,
            fetchNextPage: jest.fn(),
        });

        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );

        expect(screen.getByRole('button', { name: 'LOAD_MORE' })).toBeInTheDocument();
    });

    it('renders the dedicated favorites layout and delegates filter changes', () => {
        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('FAVORITES_PAGE_TITLE')).toBeInTheDocument();
        expect(screen.getByText('FAVORITES_PAGE_DESCRIPTION')).toBeInTheDocument();
        expect(screen.getByText('FAVORITES_PAGE_TITLE')).toBeInTheDocument();
        expect(screen.getByTestId('favorites-file-content')).toHaveAttribute(
            'data-view-mode',
            'grid'
        );
        expect(screen.getByTestId('favorites-file-content')).toHaveAttribute(
            'data-title',
            'Projects'
        );
        expect(screen.getByTestId('favorites-file-content')).toHaveAttribute(
            'data-empty-state',
            'FAVORITES_EMPTY_STATE'
        );

        fireEvent.click(screen.getByRole('button', { name: 'FAVORITES_FILTER_FILES' }));
        expect(mockSetActiveFilter).toHaveBeenCalledWith('files');

        fireEvent.click(screen.getByRole('button', { name: 'FILES_VIEW_LIST' }));
        expect(mockSetViewMode).toHaveBeenCalledWith('list');
    });

    it('shows the preview column when a file is selected', () => {
        mockUseFavoritesScreen.mockReturnValue({
            activeFilter: 'all',
            activeFilterLabel: 'FAVORITES_FILTER_ALL',
            breadcrumbSegments: [
                { id: null, label: 'STARRED_FILES', path: null, isCurrent: false },
                { id: 10, label: 'Projects', path: '/projects', isCurrent: false },
                {
                    id: 99,
                    label: 'Invoice.pdf',
                    path: '/projects/Invoice.pdf',
                    isCurrent: true,
                },
            ],
            contextPath: '/projects',
            currentTitle: 'Invoice.pdf',
            filterOptions: [
                { value: 'all', label: 'FAVORITES_FILTER_ALL', count: 3 },
                { value: 'folders', label: 'FAVORITES_FILTER_FOLDERS', count: 1 },
                { value: 'files', label: 'FAVORITES_FILTER_FILES', count: 1 },
                { value: 'media', label: 'FAVORITES_FILTER_MEDIA', count: 1 },
            ],
            listedItems: [],
            itemCountLabel: '1 ITEM',
            selectedItem: { id: 99, type: 2, name: 'Invoice.pdf' },
            isRootListing: true,
            setActiveFilter: mockSetActiveFilter,
            setStarredSort: mockSetStarredSort,
            setViewMode: mockSetViewMode,
            sortOptions: [
                { value: 'name', label: 'FAVORITES_SORT_NAME' },
                { value: 'starred_at', label: 'FAVORITES_SORT_STARRED_AT' },
            ],
            starredSort: 'name',
            viewMode: 'grid',
        });

        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );

        expect(screen.getByText(/FileDetailsMock:Invoice.pdf/)).toBeInTheDocument();
    });

    it('closes the selected item from the details panel', () => {
        const closeSelectedItem = jest.fn();
        mockUseFavoritesScreen.mockReturnValue({
            ...mockUseFavoritesScreen(),
            selectedItem: { id: 99, type: 2, name: 'Invoice.pdf' },
            closeSelectedItem,
        });

        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );
        fireEvent.click(screen.getByText('CloseDetailsMock'));

        expect(closeSelectedItem).toHaveBeenCalledTimes(1);
    });

    it('navigates breadcrumb segments inside favorites', () => {
        mockUseFavoritesScreen.mockReturnValue({
            ...mockUseFavoritesScreen(),
            breadcrumbSegments: [
                { id: null, label: 'STARRED_FILES', path: null, isCurrent: false },
                { id: 10, label: 'My Projects', path: '/my projects', isCurrent: false },
                { id: 11, label: 'src', path: '/my projects/src', isCurrent: true },
            ],
        });

        const LocationProbe = () => <div data-testid="location">{useLocation().pathname}</div>;
        render(
            <MemoryRouter initialEntries={['/favorites/my%20projects/src']}>
                <FavoritesScreen />
                <Routes>
                    <Route path="*" element={<LocationProbe />} />
                </Routes>
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('button', { name: 'My Projects' }));
        expect(screen.getByTestId('location')).toHaveTextContent('/favorites/my%20projects');

        fireEvent.click(screen.getByRole('button', { name: 'STARRED_FILES' }));
        expect(screen.getByTestId('location')).toHaveTextContent(/^\/favorites$/);
    });

    it('delegates sort changes to the sort selector', () => {
        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('button', { name: 'FAVORITES_SORT_STARRED_AT' }));

        expect(mockSetStarredSort).toHaveBeenCalledWith('starred_at');
    });

    it('hides filter chips and sort selector inside a starred folder', () => {
        mockUseFavoritesScreen.mockReturnValue({
            ...mockUseFavoritesScreen(),
            isRootListing: false,
            selectedItem: { id: 10, type: 1, name: 'Projects' },
        });

        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );

        expect(screen.queryByRole('button', { name: 'FAVORITES_FILTER_FILES' })).toBeNull();
        expect(screen.queryByRole('button', { name: 'FAVORITES_SORT_NAME' })).toBeNull();
    });

    it('omits the chip count while the server counts are unavailable', () => {
        mockUseFavoritesScreen.mockReturnValue({
            ...mockUseFavoritesScreen(),
            filterOptions: [{ value: 'all', label: 'FAVORITES_FILTER_ALL', count: null }],
        });

        render(
            <MemoryRouter>
                <FavoritesScreen />
            </MemoryRouter>
        );

        expect(screen.getByRole('button', { name: 'FAVORITES_FILTER_ALL' })).toHaveTextContent(
            /^FAVORITES_FILTER_ALL$/
        );
    });
});

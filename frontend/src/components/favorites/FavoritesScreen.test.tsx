import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import FavoritesScreen from './FavoritesScreen';

const mockUseFavoritesScreen = jest.fn();
const mockSetActiveFilter = jest.fn();
const mockSetViewMode = jest.fn();

jest.mock('./useFavoritesScreen', () => ({
	__esModule: true,
	default: () => mockUseFavoritesScreen(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

jest.mock('@/components/fileContent', () => ({ viewMode, title, emptyStateMessage }: any) => (
	<div
		data-testid='favorites-file-content'
		data-view-mode={viewMode}
		data-title={title}
		data-empty-state={emptyStateMessage}
	>
		FileContentMock
	</div>
));

jest.mock('@/components/fileDetails', () => () => <div>FileDetailsMock</div>);

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
			filteredItems: [{ id: 1, name: 'Invoice.pdf' }],
			itemCountLabel: '3 ITENS',
			selectedItem: null,
			setActiveFilter: mockSetActiveFilter,
			setViewMode: mockSetViewMode,
			viewMode: 'grid',
		});
	});

	it('renders the dedicated favorites layout and delegates filter changes', () => {
		render(<MemoryRouter><FavoritesScreen /></MemoryRouter>);

		expect(screen.getByText('FAVORITES_PAGE_TITLE')).toBeInTheDocument();
		expect(screen.getByText('FAVORITES_PAGE_DESCRIPTION')).toBeInTheDocument();
		expect(screen.getByText('FAVORITES_EYEBROW')).toBeInTheDocument();
		expect(screen.getByTestId('favorites-file-content')).toHaveAttribute('data-view-mode', 'grid');
		expect(screen.getByTestId('favorites-file-content')).toHaveAttribute('data-title', 'Projects');
		expect(screen.getByTestId('favorites-file-content')).toHaveAttribute('data-empty-state', 'FAVORITES_EMPTY_STATE');

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
				{ id: 99, label: 'Invoice.pdf', path: '/projects/Invoice.pdf', isCurrent: true },
			],
			contextPath: '/projects',
			currentTitle: 'Invoice.pdf',
			filterOptions: [
				{ value: 'all', label: 'FAVORITES_FILTER_ALL', count: 3 },
				{ value: 'folders', label: 'FAVORITES_FILTER_FOLDERS', count: 1 },
				{ value: 'files', label: 'FAVORITES_FILTER_FILES', count: 1 },
				{ value: 'media', label: 'FAVORITES_FILTER_MEDIA', count: 1 },
			],
			filteredItems: [],
			itemCountLabel: '1 ITEM',
			selectedItem: { id: 99, type: 2, name: 'Invoice.pdf' },
			setActiveFilter: mockSetActiveFilter,
			setViewMode: mockSetViewMode,
			viewMode: 'grid',
		});

		render(<MemoryRouter><FavoritesScreen /></MemoryRouter>);

		expect(screen.getByText('FileDetailsMock')).toBeInTheDocument();
	});
});

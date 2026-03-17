import { render } from '@testing-library/react';
import GlobalSearchProvider from './GlobalSearchProvider';

const mockOpenSearch = jest.fn();
const mockCloseSearch = jest.fn();
const mockSetQuery = jest.fn();
const mockSetActiveIndex = jest.fn();
const mockHandleInputKeyDown = jest.fn();
const mockActivateItem = jest.fn();

const mockSections = [
	{
		id: 'actions',
		title: 'Actions',
		items: [{ id: 'item-1', kind: 'action', label: 'Action', description: 'desc', onSelect: () => {} }],
	},
];

jest.mock('./useGlobalSearchProvider', () => ({
	__esModule: true,
	default: () => ({
		open: true,
		query: 'test',
		sections: mockSections,
		isFetching: false,
		activeItemId: 'item-1',
		shortcut: 'Ctrl+K',
		showEmptyState: false,
		openSearch: mockOpenSearch,
		closeSearch: mockCloseSearch,
		setQuery: mockSetQuery,
		setActiveIndex: mockSetActiveIndex,
		handleInputKeyDown: mockHandleInputKeyDown,
		activateItem: mockActivateItem,
	}),
}));

let lastProps: any;
jest.mock('./GlobalSearchDialog', () => ({
	__esModule: true,
	default: (props: any) => {
		lastProps = props;
		return null;
	},
}));

describe('GlobalSearchProvider', () => {
	beforeEach(() => {
		mockSetActiveIndex.mockReset();
	});

	it('updates active index on hover when item exists', () => {
		render(
			<GlobalSearchProvider>
				<div>child</div>
			</GlobalSearchProvider>,
		);
		lastProps.onItemHover('item-1');
		expect(mockSetActiveIndex).toHaveBeenCalled();

		lastProps.onItemHover('missing');
		expect(mockSetActiveIndex).toHaveBeenCalledTimes(1);
	});
});

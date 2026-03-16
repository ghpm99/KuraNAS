import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route, useLocation } from 'react-router-dom';
import GlobalSearchProvider from './GlobalSearchProvider';
import useGlobalSearch from './useGlobalSearch';

const mockSearchGlobal = jest.fn();

jest.mock('@/service/search', () => ({
	searchGlobal: (...args: any[]) => mockSearchGlobal(...args),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, options?: Record<string, string>) =>
			Object.entries(options ?? {}).reduce((message, [optionKey, value]) => {
				return message.replace(`{{${optionKey}}}`, value);
			}, key),
	}),
}));

const SearchLauncher = () => {
	const { openSearch } = useGlobalSearch();
	return <button onClick={openSearch}>Launch search</button>;
};

const LocationProbe = () => {
	const location = useLocation();
	return <div data-testid='location-probe'>{`${location.pathname}${location.search}`}</div>;
};

const renderSearchProvider = () => {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				retry: false,
			},
		},
	});

	return render(
		<QueryClientProvider client={queryClient}>
			<MemoryRouter initialEntries={['/home']}>
				<GlobalSearchProvider>
					<SearchLauncher />
					<LocationProbe />
					<Routes>
						<Route path='*' element={<div />} />
					</Routes>
				</GlobalSearchProvider>
			</MemoryRouter>
		</QueryClientProvider>,
	);
};

describe('components/search/GlobalSearchProvider', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockSearchGlobal.mockResolvedValue({
			query: 'mix',
			files: [
				{
					id: 11,
					name: 'mix.mp3',
					path: '/music/mix.mp3',
					parent_path: '/music',
					format: '.mp3',
					starred: true,
				},
			],
			folders: [],
			artists: [],
			albums: [],
			playlists: [],
			videos: [],
			images: [],
		});
	});

	it('opens the dialog, shows quick actions, and navigates when an action is selected', async () => {
		renderSearchProvider();

		fireEvent.click(screen.getByText('Launch search'));

		expect(screen.getByPlaceholderText('GLOBAL_SEARCH_PLACEHOLDER')).toBeInTheDocument();
		expect(screen.getByText('GLOBAL_SEARCH_SECTION_ACTIONS')).toBeInTheDocument();

		fireEvent.click(screen.getByText('SETTINGS'));

		await waitFor(() => {
			expect(screen.getByTestId('location-probe')).toHaveTextContent('/settings');
		});
	});

	it('queries the backend and opens file targets from search results', async () => {
		renderSearchProvider();

		fireEvent.click(screen.getByText('Launch search'));
		fireEvent.change(screen.getByLabelText('GLOBAL_SEARCH_OPEN'), { target: { value: 'mix' } });

		await waitFor(() => {
			expect(mockSearchGlobal).toHaveBeenCalledWith('mix', 6);
		});

		expect(await screen.findByText('GLOBAL_SEARCH_SECTION_FILES')).toBeInTheDocument();
		fireEvent.click(screen.getByText('mix.mp3'));

		await waitFor(() => {
			expect(screen.getByTestId('location-probe')).toHaveTextContent('/files/music/mix.mp3');
		});
	});

	it('renders the empty state when no search results are available', async () => {
		mockSearchGlobal.mockResolvedValueOnce({
			query: 'void',
			files: [],
			folders: [],
			artists: [],
			albums: [],
			playlists: [],
			videos: [],
			images: [],
		});

		renderSearchProvider();

		fireEvent.click(screen.getByText('Launch search'));
		fireEvent.change(screen.getByLabelText('GLOBAL_SEARCH_OPEN'), { target: { value: 'void' } });

		expect(await screen.findByText('GLOBAL_SEARCH_EMPTY_TITLE')).toBeInTheDocument();
		expect(screen.getByText('GLOBAL_SEARCH_EMPTY_DESCRIPTION')).toBeInTheDocument();
	});
});

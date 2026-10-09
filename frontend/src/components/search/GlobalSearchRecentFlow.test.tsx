import { act, fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import GlobalSearchProvider from './GlobalSearchProvider';
import { searchHistoryStorageKey } from './searchHistoryStorage';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params ? `${key}:${JSON.stringify(params)}` : key,
    }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const emptyResponse = {
    query: '',
    files: [],
    folders: [],
    artists: [],
    albums: [],
    playlists: [],
    videos: [],
    images: [],
};

const renderProvider = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <GlobalSearchProvider>
                    <div />
                </GlobalSearchProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const openSearch = () => fireEvent.keyDown(window, { key: 'k', ctrlKey: true });

const typeQuery = (text: string) =>
    fireEvent.change(screen.getByRole('combobox'), { target: { value: text } });

const settle = async () => {
    await act(async () => {
        await jest.advanceTimersByTimeAsync(250);
    });
    await act(async () => {
        await jest.advanceTimersByTimeAsync(0);
    });
};

const readStoredHistory = () =>
    JSON.parse(window.localStorage.getItem(searchHistoryStorageKey) ?? '[]');

describe('components/search/GlobalSearchProvider recent searches and fuzzy', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        window.localStorage.clear();
        mockedApi.get.mockReset();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('opens with a rejecting backend and no stored history', () => {
        mockedApi.get.mockRejectedValue(new Error('offline'));
        renderProvider();

        openSearch();

        expect(screen.getByRole('combobox')).toBeInTheDocument();
        expect(screen.queryByText('GLOBAL_SEARCH_SECTION_RECENT')).not.toBeInTheDocument();
    });

    it('records the query on Enter, lists it when reopened and re-runs it on click', async () => {
        mockedApi.get.mockResolvedValue({ data: emptyResponse });
        renderProvider();
        openSearch();
        typeQuery('ferias');
        await settle();

        fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' });
        expect(readStoredHistory()).toEqual(['ferias']);

        typeQuery('');
        expect(screen.getByText('ferias')).toBeInTheDocument();

        fireEvent.click(screen.getByText('ferias'));
        expect(screen.getByRole('combobox')).toHaveValue('ferias');
    });

    it('does not record queries shorter than two characters', async () => {
        mockedApi.get.mockResolvedValue({ data: emptyResponse });
        renderProvider();
        openSearch();
        typeQuery('f');
        await settle();

        fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' });

        expect(readStoredHistory()).toEqual([]);
    });

    it('removes and clears stored searches from the dialog', async () => {
        window.localStorage.setItem(
            searchHistoryStorageKey,
            JSON.stringify(['ferias', 'contrato'])
        );
        mockedApi.get.mockResolvedValue({ data: emptyResponse });
        renderProvider();
        openSearch();

        fireEvent.click(screen.getByLabelText('GLOBAL_SEARCH_RECENT_REMOVE:{"query":"ferias"}'));
        expect(readStoredHistory()).toEqual(['contrato']);

        fireEvent.click(screen.getByText('GLOBAL_SEARCH_RECENT_CLEAR'));
        expect(readStoredHistory()).toEqual([]);
        expect(screen.queryByText('GLOBAL_SEARCH_SECTION_RECENT')).not.toBeInTheDocument();
    });

    it('shows the fuzzy notice when the backend flags a typo-tolerant response', async () => {
        mockedApi.get.mockResolvedValue({ data: { ...emptyResponse, fuzzy: true } });
        renderProvider();
        openSearch();
        typeQuery('fereas');
        await settle();

        expect(screen.getByText('GLOBAL_SEARCH_FUZZY_NOTICE')).toBeInTheDocument();
    });

    it('does not show the fuzzy notice for exact responses', async () => {
        mockedApi.get.mockResolvedValue({ data: emptyResponse });
        renderProvider();
        openSearch();
        typeQuery('ferias');
        await settle();

        expect(screen.queryByText('GLOBAL_SEARCH_FUZZY_NOTICE')).not.toBeInTheDocument();
    });
});

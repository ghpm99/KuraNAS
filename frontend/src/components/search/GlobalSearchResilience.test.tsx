import { act, fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import GlobalSearchProvider from './GlobalSearchProvider';
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

const buildFile = (name: string) => ({
    id: name.length,
    name,
    path: `/${name}`,
    parent_path: '/',
    format: '.jpg',
    starred: false,
});

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

const advance = async (ms: number) => {
    await act(async () => {
        await jest.advanceTimersByTimeAsync(ms);
    });
};

const typeAndSettle = async (text: string) => {
    typeQuery(text);
    await advance(250);
    await advance(0);
};

const searchCalls = () => mockedApi.get.mock.calls.filter(([url]) => url === '/search/global');

describe('components/search/GlobalSearchProvider resilience', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        mockedApi.get.mockReset();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('renders and opens without any backend response', async () => {
        mockedApi.get.mockRejectedValue(new Error('offline'));
        renderProvider();

        openSearch();

        expect(screen.getByRole('combobox')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('waits 250ms of silence before requesting and only requests the last query', async () => {
        mockedApi.get.mockResolvedValue({ data: emptyResponse });
        renderProvider();
        openSearch();

        typeQuery('tr');
        await advance(200);
        typeQuery('tri');
        await advance(200);
        typeQuery('trip');
        await advance(249);
        expect(searchCalls()).toHaveLength(0);

        await advance(1);

        expect(searchCalls()).toHaveLength(1);
        expect(searchCalls()[0][1].params).toEqual({ q: 'trip', limit: 6 });
    });

    it('passes an abort signal that is aborted when the query changes mid-flight', async () => {
        mockedApi.get.mockImplementation(() => new Promise(() => undefined));
        renderProvider();
        openSearch();

        typeQuery('trip');
        await advance(250);
        const firstSignal: AbortSignal = searchCalls()[0][1].signal;
        expect(firstSignal.aborted).toBe(false);

        typeQuery('trips');
        await advance(250);

        expect(firstSignal.aborted).toBe(true);
        expect(searchCalls()).toHaveLength(2);
    });

    it('keeps the previous results visible and shows the updating indicator while fetching', async () => {
        mockedApi.get.mockResolvedValueOnce({
            data: { ...emptyResponse, files: [buildFile('trip.jpg')] },
        });
        renderProvider();
        openSearch();
        await typeAndSettle('trip');
        expect(screen.getByRole('option', { name: /trip\s*\.jpg/ })).toBeInTheDocument();
        expect(screen.queryByText('GLOBAL_SEARCH_UPDATING')).not.toBeInTheDocument();

        mockedApi.get.mockImplementation(() => new Promise(() => undefined));
        typeQuery('trips');
        await advance(250);

        expect(screen.getByRole('option', { name: /trip\s*\.jpg/ })).toBeInTheDocument();
        expect(screen.getByText('GLOBAL_SEARCH_UPDATING')).toBeInTheDocument();
        expect(screen.queryByText('GLOBAL_SEARCH_EMPTY_TITLE')).not.toBeInTheDocument();
    });

    it('shows the backend message verbatim on failure, not the empty state, and retries', async () => {
        mockedApi.get.mockRejectedValueOnce({
            response: { data: { error: 'Busca indisponivel' } },
        });
        renderProvider();
        openSearch();
        await typeAndSettle('trip');

        expect(screen.getByRole('alert')).toHaveTextContent('GLOBAL_SEARCH_ERROR_TITLE');
        expect(screen.getByRole('alert')).toHaveTextContent('Busca indisponivel');
        expect(screen.queryByText('GLOBAL_SEARCH_EMPTY_TITLE')).not.toBeInTheDocument();

        mockedApi.get.mockResolvedValueOnce({
            data: { ...emptyResponse, files: [buildFile('trip.jpg')] },
        });
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));
        await advance(0);

        expect(screen.getByRole('option', { name: /trip\s*\.jpg/ })).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(searchCalls()).toHaveLength(2);
    });

    it('shows the empty state, not an error, when the search finds nothing', async () => {
        mockedApi.get.mockResolvedValue({ data: emptyResponse });
        renderProvider();
        openSearch();
        await typeAndSettle('zzzz');

        expect(screen.getByText('GLOBAL_SEARCH_EMPTY_TITLE')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('shows the error without a backend message when the failure carries none', async () => {
        mockedApi.get.mockRejectedValueOnce(new Error('network'));
        renderProvider();
        openSearch();
        await typeAndSettle('trip');

        expect(screen.getByRole('alert')).toHaveTextContent('GLOBAL_SEARCH_ERROR_TITLE');
    });
});

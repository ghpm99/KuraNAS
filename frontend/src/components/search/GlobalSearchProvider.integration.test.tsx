import { fireEvent, render, screen, waitFor } from '@testing-library/react';
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
    files: [],
    folders: [],
    artists: [],
    albums: [],
    playlists: [],
    videos: [],
    images: [],
};

const baseFile = {
    id: 1,
    name: 'trip.jpg',
    path: '/trip.jpg',
    parent_path: '/',
    format: '.jpg',
    starred: false,
};

const aiFile = { ...baseFile, id: 2, name: 'holiday.jpg', path: '/holiday.jpg' };

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

const openAndType = (text: string) => {
    fireEvent.keyDown(window, { key: 'k', ctrlKey: true });
    fireEvent.change(screen.getByRole('combobox'), { target: { value: text } });
};

describe('components/search/GlobalSearchProvider (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/search/global') {
                return Promise.resolve({
                    data: { ...emptyResponse, query: 'my trip', files: [baseFile] },
                });
            }
            if (url === '/search/global/ai') {
                return Promise.resolve({
                    data: {
                        ...emptyResponse,
                        query: 'my trip',
                        suggestion: 'Try the holiday folder',
                        files: [baseFile, aiFile],
                    },
                });
            }
            return Promise.reject(new Error(`unexpected GET ${url}`));
        });
    });

    it('shows base results without calling the AI endpoint', async () => {
        renderProvider();
        openAndType('my trip');

        expect(await screen.findByText('trip.jpg')).toBeInTheDocument();
        expect(mockedApi.get).toHaveBeenCalledWith('/search/global', {
            params: { q: 'my trip', limit: 6 },
        });
        expect(mockedApi.get).not.toHaveBeenCalledWith('/search/global/ai', expect.anything());
    });

    it('does not offer the AI search for a single-word query', async () => {
        renderProvider();
        openAndType('trip');

        expect(await screen.findByText('trip.jpg')).toBeInTheDocument();
        expect(screen.queryByText('GLOBAL_SEARCH_WITH_AI')).not.toBeInTheDocument();
    });

    it('calls the AI endpoint on demand, merges results and renders the suggestion verbatim', async () => {
        renderProvider();
        openAndType('my trip');

        fireEvent.click(await screen.findByText('GLOBAL_SEARCH_WITH_AI'));

        expect(await screen.findByText('holiday.jpg')).toBeInTheDocument();
        expect(mockedApi.get).toHaveBeenCalledWith('/search/global/ai', {
            params: { q: 'my trip', limit: 6 },
        });
        expect(screen.getByText('Try the holiday folder')).toBeInTheDocument();
        await waitFor(() =>
            expect(screen.queryByText('GLOBAL_SEARCH_WITH_AI')).not.toBeInTheDocument()
        );
    });
});

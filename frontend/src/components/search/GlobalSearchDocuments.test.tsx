import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, useLocation } from 'react-router-dom';
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

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useOptionalGlobalMusic: () => undefined,
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const documentResult = {
    file_id: 9,
    name: 'contrato.pdf',
    path: '/Docs/Juridico/contrato.pdf',
    parent_path: '/Docs/Juridico',
    format: '.pdf',
    size: 2048,
    updated_at: '2026-01-02T10:00:00Z',
    snippet: 'a clausula de rescisao do contrato',
};

const baseResponse = {
    query: 'rescisao',
    files: [],
    folders: [],
    artists: [],
    albums: [],
    playlists: [],
    videos: [],
    images: [],
};

const LocationProbe = () => {
    const location = useLocation();
    return <span data-testid="location">{`${location.pathname}${location.search}`}</span>;
};

const renderProvider = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <GlobalSearchProvider>
                    <LocationProbe />
                </GlobalSearchProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const openAndType = (text: string) => {
    fireEvent.keyDown(window, { key: 'k', ctrlKey: true });
    fireEvent.change(screen.getByRole('combobox'), { target: { value: text } });
};

describe('components/search/GlobalSearchProvider documents', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: { ...baseResponse, documents: [documentResult] } });
    });

    it('lists documents with an ellipsized, highlighted snippet and the parent path', async () => {
        renderProvider();
        openAndType('rescisao');

        expect(await screen.findByText('GLOBAL_SEARCH_SECTION_DOCUMENTS')).toBeInTheDocument();
        expect(screen.getByText('contrato.pdf')).toBeInTheDocument();
        expect(screen.getByText('/Docs/Juridico')).toBeInTheDocument();

        const highlightedTerm = screen.getByText('rescisao', { selector: 'mark' });
        const snippetElement = highlightedTerm.parentElement as HTMLElement;
        expect(snippetElement.textContent).toBe('…a clausula de rescisao do contrato…');
    });

    it('opens the document in the files explorer when clicked', async () => {
        renderProvider();
        openAndType('rescisao');

        fireEvent.click(await screen.findByText('contrato.pdf'));

        expect(screen.getByTestId('location')).toHaveTextContent('/files/Docs/Juridico/contrato.pdf');
        await waitFor(() => expect(screen.queryByRole('combobox')).not.toBeInTheDocument());
    });

    it('opens the document with Enter when it is the best result', async () => {
        renderProvider();
        openAndType('rescisao');
        await screen.findByText('contrato.pdf');

        fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' });

        expect(screen.getByTestId('location')).toHaveTextContent('/files/Docs/Juridico/contrato.pdf');
    });

    it('offers a see-all entry that opens the files screen in content mode', async () => {
        renderProvider();
        openAndType('rescisao');

        fireEvent.click(await screen.findByText('GLOBAL_SEARCH_SEE_ALL_DOCUMENTS'));

        expect(screen.getByTestId('location')).toHaveTextContent('/files?q=rescisao&in=content');
    });

    it('renders results from an older backend that sends no documents', async () => {
        mockedApi.get.mockResolvedValue({ data: baseResponse });
        renderProvider();
        openAndType('rescisao');

        expect(await screen.findByText('GLOBAL_SEARCH_EMPTY_TITLE')).toBeInTheDocument();
        expect(screen.queryByText('GLOBAL_SEARCH_SECTION_DOCUMENTS')).toBeNull();
    });
});

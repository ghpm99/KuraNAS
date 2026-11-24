import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsScreen from './documentsScreenInFileProvider';
import { documentsSortStorageKey } from './documentsSortPreference';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const mockNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useNavigate: () => mockNavigate,
}));

const pagination = { page: 1, page_size: 20, has_next: false, has_prev: false };

const listItem = {
    id: 1,
    name: 'lista.pdf',
    path: '/docs/lista.pdf',
    parent_path: '/docs',
    format: 'pdf',
    size: 10,
    modified_at: '2026-01-01T00:00:00Z',
    created_at: '2026-01-01T00:00:00Z',
    starred: false,
};

const searchResult = {
    file_id: 9,
    name: 'contrato.pdf',
    path: '/docs/contrato.pdf',
    parent_path: '/docs',
    format: 'pdf',
    size: 2048,
    updated_at: '2026-03-01T10:00:00Z',
    snippet: 'cláusula de rescisão do contrato',
    page: 4,
};

const requestsTo = (url: string) =>
    mockGet.mock.calls.filter(([requestedUrl]) => requestedUrl === url);

const lastParamsOf = (url: string) => {
    const calls = requestsTo(url);
    return calls[calls.length - 1]?.[1]?.params;
};

const routeRequests = (searchResponse: () => Promise<unknown> = () =>
    Promise.resolve({ data: { items: [searchResult], pagination } })) => {
    mockGet.mockImplementation((url: string) => {
        if (url === '/documents/count') return Promise.resolve({ data: { total: 1 } });
        if (url === '/documents') {
            return Promise.resolve({ data: { items: [listItem], pagination } });
        }
        if (url === '/documents/search') return searchResponse();
        return Promise.reject(new Error('no backend'));
    });
};

const searchInput = () => screen.getByRole('textbox', { name: 'DOCUMENTS_SEARCH_PLACEHOLDER' });

describe('DocumentsScreen search', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
    });

    it('renders the search field without any backend', async () => {
        mockGet.mockRejectedValue(new Error('no backend'));

        expect(() => renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato'])).not.toThrow();

        expect(searchInput()).toHaveValue('contrato');
        expect(await screen.findByRole('alert')).toBeInTheDocument();
    });

    it('searches with q and the current filters, folder and relevance order, with the exact params', async () => {
        routeRequests();

        renderWithoutBackend(<DocumentsScreen />, [
            '/documents?q=contrato&format=pdf&favorites_only=true&folder_id=5&sort=name',
        ]);

        await waitFor(() => expect(requestsTo('/documents/search')).toHaveLength(1));
        expect(lastParamsOf('/documents/search')).toEqual({
            q: 'contrato',
            format: 'pdf',
            folder_id: 5,
            favorites_only: 'true',
            page: 1,
            page_size: 20,
        });
        expect(requestsTo('/documents')).toHaveLength(0);
        expect(screen.getByText('FILES_SEARCH_SORT_RELEVANCE')).toBeInTheDocument();
        expect(await screen.findByText('DOCUMENTS_SEARCH_PAGE')).toBeInTheDocument();
        expect(screen.getAllByText('contrato', { selector: 'mark' }).length).toBeGreaterThan(0);
    });

    it('debounces typing and ignores queries shorter than two characters', async () => {
        routeRequests();
        renderWithoutBackend(<DocumentsScreen />);
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));

        fireEvent.change(searchInput(), { target: { value: 'c' } });
        await new Promise((resolve) => setTimeout(resolve, 400));
        expect(requestsTo('/documents/search')).toHaveLength(0);

        fireEvent.change(searchInput(), { target: { value: 'co' } });
        fireEvent.change(searchInput(), { target: { value: 'con' } });
        expect(requestsTo('/documents/search')).toHaveLength(0);

        await waitFor(() => expect(requestsTo('/documents/search')).toHaveLength(1));
        expect(lastParamsOf('/documents/search')).toMatchObject({ q: 'con' });
    });

    it('reproduces the same search when remounted with q in the URL (F5)', async () => {
        routeRequests();

        const firstMount = renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato']);
        await waitFor(() => expect(requestsTo('/documents/search')).toHaveLength(1));
        const paramsBeforeReload = lastParamsOf('/documents/search');
        firstMount.unmount();

        renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato']);
        await waitFor(() => expect(requestsTo('/documents/search')).toHaveLength(2));

        expect(lastParamsOf('/documents/search')).toEqual(paramsBeforeReload);
    });

    it('clearing the search brings back the list with the remembered sort', async () => {
        window.localStorage.setItem(
            documentsSortStorageKey,
            JSON.stringify({ sort: 'size', order: 'asc' })
        );
        routeRequests();
        renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato']);
        expect(await screen.findByText('DOCUMENTS_SEARCH_PAGE')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_CLEAR' }));

        expect(await screen.findByText('lista.pdf')).toBeInTheDocument();
        expect(lastParamsOf('/documents')).toMatchObject({ sort: 'size', order: 'asc' });
        expect(searchInput()).toHaveValue('');
    });

    it('shows the empty-search state', async () => {
        routeRequests(() => Promise.resolve({ data: { items: [], pagination } }));

        renderWithoutBackend(<DocumentsScreen />, ['/documents?q=inexistente']);

        expect(await screen.findByText('DOCUMENTS_SEARCH_EMPTY_TITLE')).toBeInTheDocument();
    });

    it('keeps loaded results and shows the error at the end when the next page fails', async () => {
        window.IntersectionObserver = class {
            constructor(private readonly onIntersect: IntersectionObserverCallback) {}
            observe() {
                this.onIntersect(
                    [{ isIntersecting: true } as IntersectionObserverEntry],
                    this as unknown as IntersectionObserver
                );
            }
            unobserve() {}
            disconnect() {}
            takeRecords() {
                return [];
            }
        } as unknown as typeof IntersectionObserver;
        mockGet.mockImplementation((url: string, config?: { params?: { page?: number } }) => {
            if (url === '/documents/count') return Promise.resolve({ data: { total: 1 } });
            if (url !== '/documents/search') return Promise.reject(new Error('no backend'));
            if (config?.params?.page === 2) return Promise.reject(new Error('boom'));
            return Promise.resolve({
                data: { items: [searchResult], pagination: { ...pagination, has_next: true } },
            });
        });

        renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato']);

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.getByText('DOCUMENTS_SEARCH_PAGE')).toBeInTheDocument();
    });

    it('opens the result in Files keeping the page in the URL', async () => {
        routeRequests();
        renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato']);

        await screen.findByText('DOCUMENTS_SEARCH_PAGE');
        fireEvent.click(screen.getByRole('button', { name: /DOCUMENTS_SEARCH_PAGE/ }));

        expect(mockNavigate).toHaveBeenCalledWith('/files/docs/contrato.pdf?page=4');
    });
});

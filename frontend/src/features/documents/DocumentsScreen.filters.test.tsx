import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsScreen from './documentsScreenInFileProvider';
import { documentsSortStorageKey } from './documentsSortPreference';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const emptyPage = {
    data: { items: [], pagination: { page: 1, page_size: 50, has_next: false, has_prev: false } },
};

const requestsTo = (url: string) =>
    mockGet.mock.calls.filter(([requestedUrl]) => requestedUrl === url);

const lastParamsOf = (url: string) => {
    const calls = requestsTo(url);
    return calls[calls.length - 1]?.[1]?.params;
};

const routeRequests = (listResponse: () => Promise<unknown> = () => Promise.resolve(emptyPage)) => {
    mockGet.mockImplementation((url: string) => {
        if (url === '/documents/count') return Promise.resolve({ data: { total: 0 } });
        if (url === '/documents') return listResponse();
        return Promise.reject(new Error('no backend'));
    });
};

const combinedUrl =
    '/documents?format=pdf&modified_from=2026-01-01&modified_to=2026-02-01&favorites_only=true&index_state=failed&sort=name&order=desc';

describe('DocumentsScreen filters and sort', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
    });

    it('renders the filter bar and sort control without any backend', async () => {
        mockGet.mockRejectedValue(new Error('no backend'));

        expect(() => renderWithoutBackend(<DocumentsScreen />)).not.toThrow();

        expect(screen.getByRole('button', { name: 'DOCUMENTS_FILTER_FORMAT' })).toBeInTheDocument();
        expect(screen.getByRole('combobox')).toBeInTheDocument();
        expect(await screen.findByRole('alert')).toBeInTheDocument();
    });

    it('asks for modified_at desc by default with no filters', async () => {
        routeRequests();

        renderWithoutBackend(<DocumentsScreen />);

        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));
        expect(lastParamsOf('/documents')).toEqual({
            sort: 'modified_at',
            order: 'desc',
            page: 1,
            page_size: 50,
        });
        expect(lastParamsOf('/documents/count')).toEqual({});
    });

    it('sends the exact combined filter and sort params and counts with the same filters', async () => {
        routeRequests();

        renderWithoutBackend(<DocumentsScreen />, [combinedUrl]);

        await waitFor(() => expect(requestsTo('/documents/count')).toHaveLength(1));
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));
        const filterParams = {
            format: 'pdf',
            modified_from: '2026-01-01',
            modified_to: '2026-02-01',
            favorites_only: 'true',
            index_state: 'failed',
        };
        expect(lastParamsOf('/documents')).toEqual({
            ...filterParams,
            sort: 'name',
            order: 'desc',
            page: 1,
            page_size: 50,
        });
        expect(lastParamsOf('/documents/count')).toEqual(filterParams);
    });

    it('reproduces the same requests when remounted with the same URL (F5)', async () => {
        routeRequests();

        const firstMount = renderWithoutBackend(<DocumentsScreen />, [combinedUrl]);
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));
        const paramsBeforeReload = lastParamsOf('/documents');
        firstMount.unmount();

        renderWithoutBackend(<DocumentsScreen />, [combinedUrl]);
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(2));

        expect(lastParamsOf('/documents')).toEqual(paramsBeforeReload);
    });

    it('falls back to defaults for invalid URL values', async () => {
        routeRequests();

        renderWithoutBackend(<DocumentsScreen />, ['/documents?format=exe&sort=color&order=up']);

        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));
        expect(lastParamsOf('/documents')).toEqual({
            sort: 'modified_at',
            order: 'desc',
            page: 1,
            page_size: 50,
        });
    });

    it('uses the sort remembered on the device when the URL has none, and the URL wins otherwise', async () => {
        window.localStorage.setItem(
            documentsSortStorageKey,
            JSON.stringify({ sort: 'size', order: 'asc' })
        );
        routeRequests();

        const remembered = renderWithoutBackend(<DocumentsScreen />);
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));
        expect(lastParamsOf('/documents')).toMatchObject({ sort: 'size', order: 'asc' });
        remembered.unmount();

        renderWithoutBackend(<DocumentsScreen />, ['/documents?sort=name']);
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(2));
        expect(lastParamsOf('/documents')).toMatchObject({ sort: 'name', order: 'asc' });
    });

    it('refetches with the new sort and remembers it on the device', async () => {
        routeRequests();
        renderWithoutBackend(<DocumentsScreen />);
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(within(screen.getByRole('listbox')).getByText('DOCUMENTS_SORT_NAME'));

        await waitFor(() =>
            expect(lastParamsOf('/documents')).toMatchObject({ sort: 'name', order: 'asc' })
        );
        expect(JSON.parse(window.localStorage.getItem(documentsSortStorageKey) ?? '')).toEqual({
            sort: 'name',
            order: 'asc',
        });
    });

    it('applies a filter through the bar and refetches list and count', async () => {
        routeRequests();
        renderWithoutBackend(<DocumentsScreen />);
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTER_FAVORITES' }));

        await waitFor(() =>
            expect(lastParamsOf('/documents')).toMatchObject({ favorites_only: 'true' })
        );
        await waitFor(() =>
            expect(lastParamsOf('/documents/count')).toEqual({ favorites_only: 'true' })
        );
    });

    it('shows the filter-empty state with Clear filters, distinct from the library-empty state', async () => {
        routeRequests();
        renderWithoutBackend(<DocumentsScreen />, ['/documents?format=pdf']);

        expect(await screen.findByText('DOCUMENTS_FILTER_EMPTY_TITLE')).toBeInTheDocument();
        expect(screen.queryByText('DOCUMENTS_EMPTY_TITLE')).toBeNull();

        const emptyState = screen.getByRole('status');
        fireEvent.click(
            within(emptyState).getByRole('button', { name: 'DOCUMENTS_FILTERS_CLEAR' })
        );

        expect(await screen.findByText('DOCUMENTS_EMPTY_TITLE')).toBeInTheDocument();
        expect(screen.queryByText('DOCUMENTS_FILTER_EMPTY_TITLE')).toBeNull();
        expect(lastParamsOf('/documents')).not.toHaveProperty('format');
    });

    it('shows the backend 400 message verbatim', async () => {
        routeRequests(() =>
            Promise.reject({
                response: { status: 400, data: { error: 'Filtro de formato inválido' } },
            })
        );

        renderWithoutBackend(<DocumentsScreen />, ['/documents?format=pdf']);

        expect(await screen.findByText('Filtro de formato inválido')).toBeInTheDocument();
    });
});

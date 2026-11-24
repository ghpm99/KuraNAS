import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsScreen from './documentsScreenInFileProvider';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const emptyPage = {
    data: { items: [], pagination: { page: 1, page_size: 50, has_next: false, has_prev: false } },
};
const treePage = {
    data: {
        items: [{ id: 11, name: 'Contratos', type: 1 }],
        pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
    },
};

const requestsTo = (url: string) =>
    mockGet.mock.calls.filter(([requestedUrl]) => requestedUrl === url);

const lastParamsOf = (url: string) => {
    const calls = requestsTo(url);
    return calls[calls.length - 1]?.[1]?.params;
};

const routeRequests = (
    listResponse: (params: Record<string, unknown>) => Promise<unknown> = () =>
        Promise.resolve(emptyPage),
    libraries: unknown = [{ category: 'documents', path: '/data/docs' }]
) => {
    mockGet.mockImplementation((url: string, config?: { params?: Record<string, unknown> }) => {
        if (url === '/libraries') return Promise.resolve({ data: libraries });
        if (url === '/files/by-disk-path') return Promise.resolve({ data: { id: 10 } });
        if (url === '/files/tree') return Promise.resolve(treePage);
        if (url === '/documents/count') return Promise.resolve({ data: { total: 0 } });
        if (url === '/documents') return listResponse(config?.params ?? {});
        return Promise.reject(new Error('no backend'));
    });
};

describe('DocumentsScreen folder panel', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
    });

    it('shows the folder panel and lists without folder_id by default', async () => {
        routeRequests();

        renderWithoutBackend(<DocumentsScreen />);

        expect(await screen.findByText('Contratos')).toBeInTheDocument();
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));
        expect(lastParamsOf('/documents')).not.toHaveProperty('folder_id');
    });

    it('applies folder_id to the list and the count, keeping other filters', async () => {
        routeRequests();

        renderWithoutBackend(<DocumentsScreen />, ['/documents?format=pdf']);
        fireEvent.click(await screen.findByRole('button', { name: 'Contratos' }));

        await waitFor(() => expect(lastParamsOf('/documents')).toMatchObject({ folder_id: 11 }));
        expect(lastParamsOf('/documents')).toMatchObject({ format: 'pdf' });
        expect(lastParamsOf('/documents/count')).toMatchObject({ folder_id: 11, format: 'pdf' });
        expect(screen.getByRole('button', { name: 'Contratos' })).toHaveAttribute(
            'aria-current',
            'true'
        );
    });

    it('clears folder_id from the all-documents entry', async () => {
        routeRequests();

        renderWithoutBackend(<DocumentsScreen />, ['/documents?folder_id=11']);
        await waitFor(() => expect(lastParamsOf('/documents')).toMatchObject({ folder_id: 11 }));
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_ALL' }));

        await waitFor(() => expect(lastParamsOf('/documents')).not.toHaveProperty('folder_id'));
        expect(lastParamsOf('/documents/count')).not.toHaveProperty('folder_id');
    });

    it('shows the backend message verbatim for an unknown folder and offers to clear it', async () => {
        routeRequests((params) =>
            params.folder_id
                ? Promise.reject({
                      response: { status: 404, data: { error: 'Pasta não encontrada' } },
                  })
                : Promise.resolve(emptyPage)
        );

        renderWithoutBackend(<DocumentsScreen />, ['/documents?folder_id=999']);

        expect(await screen.findByText('Pasta não encontrada')).toBeInTheDocument();
        const clearButtons = screen.getAllByRole('button', { name: 'DOCUMENTS_FOLDERS_ALL' });
        fireEvent.click(clearButtons[clearButtons.length - 1] as HTMLElement);

        await waitFor(() => expect(screen.queryByText('Pasta não encontrada')).toBeNull());
        expect(lastParamsOf('/documents')).not.toHaveProperty('folder_id');
    });

    it('keeps the list working and explains the setup when the library is not configured', async () => {
        routeRequests(undefined, [{ category: 'documents', path: '' }]);

        renderWithoutBackend(<DocumentsScreen />);

        expect(await screen.findByText('DOCUMENTS_FOLDERS_NOT_CONFIGURED')).toBeInTheDocument();
        await waitFor(() => expect(requestsTo('/documents')).toHaveLength(1));
        expect(await screen.findByText('DOCUMENTS_EMPTY_TITLE')).toBeInTheDocument();
    });

    it('moves the panel into a drawer on narrow screens', async () => {
        Object.defineProperty(window, 'matchMedia', {
            configurable: true,
            writable: true,
            value: (query: string) => ({
                matches: true,
                media: query,
                onchange: null,
                addListener: jest.fn(),
                removeListener: jest.fn(),
                addEventListener: jest.fn(),
                removeEventListener: jest.fn(),
                dispatchEvent: jest.fn(),
            }),
        });
        routeRequests();

        try {
            renderWithoutBackend(<DocumentsScreen />);
            expect(screen.queryByRole('navigation')).toBeNull();

            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_TITLE' }));
            fireEvent.click(await screen.findByRole('button', { name: 'Contratos' }));

            await waitFor(() =>
                expect(lastParamsOf('/documents')).toMatchObject({ folder_id: 11 })
            );
        } finally {
            Reflect.deleteProperty(window, 'matchMedia');
        }
    });
});

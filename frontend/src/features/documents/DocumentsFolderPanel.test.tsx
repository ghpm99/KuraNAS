import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsFolderPanel from './DocumentsFolderPanel';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const folder = (id: number, name: string) => ({ id, name, type: 1 });
const file = (id: number, name: string) => ({ id, name, type: 2 });
const page = (items: unknown[], page: number, hasNext = false) => ({
    data: { items, pagination: { page, page_size: 50, has_next: hasNext, has_prev: page > 1 } },
});

const requestsTo = (url: string) =>
    mockGet.mock.calls.filter(([requestedUrl]) => requestedUrl === url);

const routeRequests = (
    treeByParent: Record<number, (requestedPage: number) => unknown>,
    libraries: unknown = [{ category: 'documents', path: '/data/docs' }]
) => {
    mockGet.mockImplementation((url: string, config?: { params?: Record<string, number> }) => {
        if (url === '/libraries') return Promise.resolve({ data: libraries });
        if (url === '/files/by-disk-path') return Promise.resolve({ data: { id: 10 } });
        if (url === '/files/tree') {
            const respond = treeByParent[config?.params?.file_parent ?? -1];
            return respond
                ? Promise.resolve(respond(config?.params?.page ?? 1))
                : Promise.reject(new Error('no parent'));
        }
        return Promise.reject(new Error('no backend'));
    });
};

describe('DocumentsFolderPanel', () => {
    beforeEach(() => jest.clearAllMocks());

    it('renders without any backend and keeps the all-documents entry', async () => {
        mockGet.mockRejectedValue(new Error('no backend'));

        expect(() => renderWithoutBackend(<DocumentsFolderPanel />)).not.toThrow();

        expect(screen.getByText('DOCUMENTS_FOLDERS_ALL')).toBeInTheDocument();
        expect(await screen.findByText('DOCUMENTS_FOLDERS_LOAD_ERROR')).toBeInTheDocument();
    });

    it('survives partial payloads', async () => {
        mockGet.mockImplementation((url: string) => {
            if (url === '/libraries') return Promise.resolve({ data: null });
            return Promise.reject(new Error('no backend'));
        });

        renderWithoutBackend(<DocumentsFolderPanel />);

        expect(await screen.findByText('DOCUMENTS_FOLDERS_NOT_CONFIGURED')).toBeInTheDocument();
    });

    it('resolves the root through libraries, by-disk-path and the tree with file_parent', async () => {
        routeRequests({ 10: () => page([folder(11, 'Contratos'), file(12, 'a.pdf')], 1) });

        renderWithoutBackend(<DocumentsFolderPanel />);

        expect(await screen.findByText('Contratos')).toBeInTheDocument();
        expect(screen.queryByText('a.pdf')).not.toBeInTheDocument();
        expect(requestsTo('/files/by-disk-path')[0][1]).toEqual({ params: { path: '/data/docs' } });
        expect(requestsTo('/files/tree')[0][1].params).toMatchObject({
            file_parent: 10,
            category: 'all',
            page: 1,
            page_size: 50,
        });
    });

    it('expands a node lazily, loading its children only then', async () => {
        routeRequests({
            10: () => page([folder(11, 'Contratos')], 1),
            11: () => page([folder(13, '2026')], 1),
        });

        renderWithoutBackend(<DocumentsFolderPanel />);
        await screen.findByText('Contratos');
        expect(requestsTo('/files/tree')).toHaveLength(1);

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_EXPAND' }));

        expect(await screen.findByText('2026')).toBeInTheDocument();
        expect(requestsTo('/files/tree')[1][1].params.file_parent).toBe(11);
    });

    it('shows no-subfolders for an expanded node without folders', async () => {
        routeRequests({
            10: () => page([folder(11, 'Contratos')], 1),
            11: () => page([file(12, 'a.pdf')], 1),
        });

        renderWithoutBackend(<DocumentsFolderPanel />);
        await screen.findByText('Contratos');
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_EXPAND' }));

        expect(await screen.findByText('DOCUMENTS_FOLDERS_NO_SUBFOLDERS')).toBeInTheDocument();
    });

    it('loads more folders when the node has another page', async () => {
        routeRequests({
            10: (requestedPage) =>
                requestedPage === 1
                    ? page([folder(11, 'Contratos')], 1, true)
                    : page([folder(14, 'Recibos')], 2),
        });

        renderWithoutBackend(<DocumentsFolderPanel />);
        await screen.findByText('Contratos');
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_LOAD_MORE' }));

        expect(await screen.findByText('Recibos')).toBeInTheDocument();
        expect(requestsTo('/files/tree')[1][1].params.page).toBe(2);
        expect(screen.queryByRole('button', { name: 'DOCUMENTS_FOLDERS_LOAD_MORE' })).toBeNull();
    });

    it('reports the clicked folder and highlights the selected one', async () => {
        routeRequests({ 10: () => page([folder(11, 'Contratos')], 1) });
        const onSelect = jest.fn();

        renderWithoutBackend(<DocumentsFolderPanel selectedFolderId={11} onSelect={onSelect} />);
        const folderButton = await screen.findByRole('button', { name: 'Contratos' });

        expect(folderButton).toHaveAttribute('aria-current', 'true');
        fireEvent.click(folderButton);
        expect(onSelect).toHaveBeenCalledWith(11);
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_ALL' }));
        expect(onSelect).toHaveBeenLastCalledWith(undefined);
    });

    it('marks the all-documents entry as current without a selected folder', async () => {
        routeRequests({ 10: () => page([], 1) });

        renderWithoutBackend(<DocumentsFolderPanel />);

        expect(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_ALL' })).toHaveAttribute(
            'aria-current',
            'true'
        );
        await waitFor(() =>
            expect(screen.getByText('DOCUMENTS_FOLDERS_NO_SUBFOLDERS')).toBeVisible()
        );
    });

    it('explains how to configure when the documents library has no path', async () => {
        routeRequests({}, [{ category: 'documents', path: '' }]);

        renderWithoutBackend(<DocumentsFolderPanel />);

        expect(await screen.findByText('DOCUMENTS_FOLDERS_NOT_CONFIGURED')).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: 'DOCUMENTS_FOLDERS_OPEN_SETTINGS' })
        ).toHaveAttribute('href', '/settings');
        expect(requestsTo('/files/by-disk-path')).toHaveLength(0);
    });

    it('offers retry when the root cannot be resolved and recovers', async () => {
        let isBackendUp = false;
        mockGet.mockImplementation((url: string) => {
            if (!isBackendUp) return Promise.reject(new Error('down'));
            if (url === '/libraries')
                return Promise.resolve({ data: [{ category: 'documents', path: '/d' }] });
            if (url === '/files/by-disk-path') return Promise.resolve({ data: { id: 10 } });
            return Promise.resolve(page([folder(11, 'Contratos')], 1));
        });

        renderWithoutBackend(<DocumentsFolderPanel />);
        await screen.findByText('DOCUMENTS_FOLDERS_LOAD_ERROR');
        isBackendUp = true;
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_RETRY' }));

        expect(await screen.findByText('Contratos')).toBeInTheDocument();
    });

    it('offers retry when a node fails to load its children', async () => {
        let shouldFailChildren = true;
        mockGet.mockImplementation((url: string) => {
            if (url === '/libraries')
                return Promise.resolve({ data: [{ category: 'documents', path: '/d' }] });
            if (url === '/files/by-disk-path') return Promise.resolve({ data: { id: 10 } });
            if (shouldFailChildren) return Promise.reject(new Error('down'));
            return Promise.resolve(page([folder(11, 'Contratos')], 1));
        });

        renderWithoutBackend(<DocumentsFolderPanel />);
        await screen.findByText('DOCUMENTS_FOLDERS_LOAD_ERROR');
        shouldFailChildren = false;
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FOLDERS_RETRY' }));

        expect(await screen.findByText('Contratos')).toBeInTheDocument();
    });
});

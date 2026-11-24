import { readFileSync } from 'fs';
import path from 'path';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsScreen from './DocumentsScreen';
import { documentsViewModeStorageKey } from './documentsViewModePreference';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const goldenDirectory = path.resolve(
    __dirname,
    '../../../../backend/internal/api/v1/documenttext/testdata'
);

const readGolden = (fileName: string) => ({
    data: JSON.parse(readFileSync(path.join(goldenDirectory, fileName), 'utf8')),
});

const routeRequests = (listHandler: (page: number) => unknown, total = 3) => {
    mockGet.mockImplementation((url: string, config?: { params?: { page?: number } }) => {
        if (url === '/documents/count') return Promise.resolve({ data: { total } });
        if (url === '/documents') return listHandler(config?.params?.page ?? 1);
        return Promise.reject(new Error('no backend'));
    });
};

describe('DocumentsScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
    });

    it('renders without any backend', async () => {
        mockGet.mockRejectedValue(new Error('no backend'));

        expect(() => renderWithoutBackend(<DocumentsScreen />)).not.toThrow();

        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('NAV_DOCUMENTS');
        expect(await screen.findByRole('alert')).toBeInTheDocument();
    });

    it('survives a payload without items or pagination', async () => {
        routeRequests(() => Promise.resolve({ data: {} }), 0);

        renderWithoutBackend(<DocumentsScreen />);

        expect(await screen.findByText('DOCUMENTS_EMPTY_TITLE')).toBeInTheDocument();
    });

    it('shows the empty state when there are no documents', async () => {
        routeRequests(
            () =>
                Promise.resolve({
                    data: {
                        items: [],
                        pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
                    },
                }),
            0
        );

        renderWithoutBackend(<DocumentsScreen />);

        expect(await screen.findByText('DOCUMENTS_EMPTY_TITLE')).toBeInTheDocument();
        expect(screen.getByText('DOCUMENTS_EMPTY_DESCRIPTION')).toBeInTheDocument();
    });

    it('renders the real first page and appends the second, with the total at the end', async () => {
        routeRequests((page) =>
            Promise.resolve(
                readGolden(
                    page === 1 ? 'documents_page_with_next.json' : 'documents_last_page.json'
                )
            )
        );

        renderWithoutBackend(<DocumentsScreen />);

        expect(await screen.findAllByText('documento.pdf')).toHaveLength(2);
        expect(await screen.findByText('DOCUMENTS_TOTAL')).toBeInTheDocument();

        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        await waitFor(() => expect(screen.getAllByText('documento.pdf')).toHaveLength(3));
        expect(screen.queryByRole('button', { name: 'LOAD_MORE' })).toBeNull();
        expect(screen.getAllByText('DOCUMENTS_TOTAL').length).toBeGreaterThan(0);
    });

    it('ignores a repeated document in the next page', async () => {
        routeRequests(() => Promise.resolve(readGolden('documents_page_with_next.json')));

        renderWithoutBackend(<DocumentsScreen />);
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        await waitFor(() =>
            expect(mockGet.mock.calls.filter(([url]) => url === '/documents')).toHaveLength(2)
        );
        expect(screen.getAllByText('documento.pdf')).toHaveLength(2);
    });

    it('keeps the first page and offers retry when the second page fails', async () => {
        let isSecondPageFailing = true;
        routeRequests((page) => {
            if (page === 1) return Promise.resolve(readGolden('documents_page_with_next.json'));
            return isSecondPageFailing
                ? Promise.reject(new Error('boom'))
                : Promise.resolve(readGolden('documents_last_page.json'));
        });

        renderWithoutBackend(<DocumentsScreen />);
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.getAllByText('documento.pdf')).toHaveLength(2);

        isSecondPageFailing = false;
        fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' }));

        await waitFor(() => expect(screen.getAllByText('documento.pdf')).toHaveLength(3));
        expect(screen.queryByRole('alert')).toBeNull();
    });

    it('reloads list and count', async () => {
        routeRequests(() => Promise.resolve(readGolden('documents_last_page.json')), 1);

        renderWithoutBackend(<DocumentsScreen />);
        await screen.findByText('documento.pdf');
        const callsBefore = mockGet.mock.calls.length;

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_RELOAD' }));

        await waitFor(() =>
            expect(mockGet.mock.calls.length).toBeGreaterThanOrEqual(callsBefore + 2)
        );
    });

    it('starts in list view and switches to grid, remembering the choice', async () => {
        routeRequests(() => Promise.resolve(readGolden('documents_last_page.json')), 1);

        const { unmount } = renderWithoutBackend(<DocumentsScreen />);
        await screen.findByText('documento.pdf');
        expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEW_LIST' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEW_GRID' }));

        expect(window.localStorage.getItem(documentsViewModeStorageKey)).toBe('grid');
        expect(screen.getByText('documento.pdf')).toBeInTheDocument();
        unmount();

        renderWithoutBackend(<DocumentsScreen />);
        await screen.findByText('documento.pdf');
        expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEW_GRID' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
    });

    it('renders the grid with the same pagination and total behavior', async () => {
        window.localStorage.setItem(documentsViewModeStorageKey, 'grid');
        routeRequests((page) =>
            Promise.resolve(
                readGolden(
                    page === 1 ? 'documents_page_with_next.json' : 'documents_last_page.json'
                )
            )
        );

        renderWithoutBackend(<DocumentsScreen />);

        expect(await screen.findAllByText('documento.pdf')).toHaveLength(2);
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        await waitFor(() => expect(screen.getAllByText('documento.pdf')).toHaveLength(3));
        expect(screen.getAllByText('DOCUMENTS_TOTAL').length).toBeGreaterThan(0);
    });
});

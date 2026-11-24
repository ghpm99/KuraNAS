import { fireEvent, screen, waitFor } from '@testing-library/react';
import { Route, Routes, useLocation } from 'react-router-dom';
import FileProvider from '@/features/files/providers/fileProvider';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentViewerScreen from './DocumentViewerScreen';

const mockGet = jest.fn();
const mockPost = jest.fn();

jest.mock('@/service', () => ({
    apiBase: {
        get: (...args: unknown[]) => mockGet(...args),
        post: (...args: unknown[]) => mockPost(...args),
    },
}));

const LocationProbe = () => {
    const location = useLocation();
    return <output aria-label="location">{`${location.pathname}${location.search}`}</output>;
};

const renderViewer = (entry: string) =>
    renderWithoutBackend(
        <>
            <Routes>
                <Route
                    path="/documents/view/:fileId"
                    element={
                        <FileProvider>
                            <DocumentViewerScreen />
                        </FileProvider>
                    }
                />
                <Route path="*" element={null} />
            </Routes>
            <LocationProbe />
        </>,
        [entry]
    );

const currentLocation = () => screen.getByLabelText('location').textContent;

const pagination = (hasNext = false) => ({
    page: 1,
    page_size: 50,
    has_next: hasNext,
    has_prev: false,
});

const buildListItem = (id: number, name: string, format = '.pdf', starred = false) => ({
    id,
    name,
    format,
    parent_path: '/Casa',
    size: 1024,
    modified_at: '2026-01-01T00:00:00Z',
    created_at: '2026-01-01T00:00:00Z',
    starred,
    index_state: 'indexed',
});

const indexedState = {
    file_id: 0,
    state: 'indexed',
    indexed_bytes: 2048,
    reason: '',
    error: '',
};

type Routing = {
    items?: ReturnType<typeof buildListItem>[];
    secondPageItems?: ReturnType<typeof buildListItem>[];
    searchItems?: Record<string, unknown>[];
    indexState?: Record<string, unknown>;
    text?: Record<string, unknown>;
    textPreview?: string;
};

const routeRequests = ({
    items = [],
    secondPageItems,
    searchItems = [],
    indexState = indexedState,
    text = { text: '', is_partial: false },
    textPreview = '',
}: Routing = {}) => {
    mockGet.mockImplementation((url: string, config?: { params?: { page?: number } }) => {
        if (url === '/documents/count') return Promise.resolve({ data: { total: items.length } });
        if (url === '/documents') {
            const isSecondPage = config?.params?.page === 2;
            return Promise.resolve({
                data: {
                    items: isSecondPage ? secondPageItems : items,
                    pagination: pagination(Boolean(secondPageItems) && !isSecondPage),
                },
            });
        }
        if (url === '/documents/search') {
            return Promise.resolve({ data: { items: searchItems, pagination: pagination() } });
        }
        if (url.endsWith('/index-state')) return Promise.resolve({ data: indexState });
        if (url.endsWith('/text')) return Promise.resolve({ data: text });
        if (url.startsWith('/files/blob/')) return Promise.resolve({ data: textPreview });
        return Promise.reject(new Error('no backend'));
    });
};

describe('DocumentViewerScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
    });

    it('renders without any backend, showing the error instead of crashing', async () => {
        mockGet.mockRejectedValue(new Error('no backend'));

        expect(() => renderViewer('/documents/view/5')).not.toThrow();

        expect(await screen.findByText('DOCUMENTS_VIEWER_LOAD_ERROR')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_BACK' })).toBeInTheDocument();
    });

    it('rejects an id that is not a number', async () => {
        routeRequests();

        renderViewer('/documents/view/abc');

        expect(await screen.findByText('DOCUMENTS_VIEWER_LOAD_ERROR')).toBeInTheDocument();
    });

    it('survives partial payloads from the list and the index state', async () => {
        mockGet.mockImplementation((url: string) =>
            url.startsWith('/documents')
                ? Promise.resolve({ data: {} })
                : Promise.reject(new Error('x'))
        );

        renderViewer('/documents/view/5');

        expect(await screen.findByText('DOCUMENTS_VIEWER_LOAD_ERROR')).toBeInTheDocument();
    });

    it('opens a PDF from a search result on the page of the snippet', async () => {
        routeRequests({
            searchItems: [
                {
                    file_id: 9,
                    name: 'contrato.pdf',
                    path: '/Casa/contrato.pdf',
                    parent_path: '/Casa',
                    format: '.pdf',
                    size: 2048,
                    updated_at: '2026-03-01T10:00:00Z',
                    snippet: 'x',
                    page: 7,
                },
            ],
        });

        renderViewer('/documents/view/9?from=q%3Dcontrato&page=7');

        const frame = await screen.findByTitle('contrato.pdf');
        expect(frame.tagName).toBe('IFRAME');
        expect(frame.getAttribute('src')).toMatch(/\/files\/blob\/9#page=7$/);
        expect(screen.getByRole('heading', { level: 1, name: 'contrato.pdf' })).toBeInTheDocument();
        expect(mockGet).toHaveBeenCalledWith(
            '/documents/search',
            expect.objectContaining({ params: expect.objectContaining({ q: 'contrato' }) })
        );
    });

    it('opens a PDF without a page when none was given', async () => {
        routeRequests({ items: [buildListItem(1, 'a.pdf')] });

        renderViewer('/documents/view/1');

        const frame = await screen.findByTitle('a.pdf');
        expect(frame.getAttribute('src')).toMatch(/\/files\/blob\/1$/);
    });

    it('shows that a protected PDF is protected instead of the frame', async () => {
        routeRequests({
            items: [buildListItem(1, 'a.pdf')],
            indexState: { ...indexedState, state: 'name_only', reason: 'encrypted' },
        });

        renderViewer('/documents/view/1');

        expect(await screen.findByText('DOCUMENTS_VIEWER_PROTECTED')).toBeInTheDocument();
        expect(screen.queryByTitle('a.pdf')).toBeNull();
        expect(screen.getByRole('link', { name: 'DOCUMENTS_VIEWER_DOWNLOAD' })).toHaveAttribute(
            'href',
            expect.stringMatching(/\/files\/download\/1$/)
        );
    });

    it('warns about a scanned PDF but still shows it', async () => {
        routeRequests({
            items: [buildListItem(1, 'a.pdf')],
            indexState: { ...indexedState, state: 'name_only', reason: 'no_text' },
        });

        renderViewer('/documents/view/1');

        expect(await screen.findByText('DOCUMENTS_VIEWER_SCANNED_PDF')).toBeInTheDocument();
        expect(screen.getByTitle('a.pdf')).toBeInTheDocument();
    });

    it('shows the extracted text with the simplified notice and the partial notice', async () => {
        routeRequests({
            items: [buildListItem(3, 'planilha.xlsx', '.xlsx')],
            text: { text: 'texto extraido da planilha', is_partial: true },
        });

        renderViewer('/documents/view/3');

        expect(await screen.findByText('texto extraido da planilha')).toBeInTheDocument();
        expect(screen.getByText('DOCUMENTS_VIEWER_SIMPLIFIED')).toBeInTheDocument();
        expect(screen.getByText('DOCUMENTS_VIEWER_PARTIAL')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'DOCUMENTS_VIEWER_DOWNLOAD' })).toBeInTheDocument();
        expect(mockGet).toHaveBeenCalledWith('/documents/3/text');
        expect(mockGet).toHaveBeenCalledWith('/documents/3/index-state');
    });

    it('does not show the partial notice for a fully indexed text', async () => {
        routeRequests({
            items: [buildListItem(3, 'a.docx', '.docx')],
            text: { text: 'completo', is_partial: false },
        });

        renderViewer('/documents/view/3');

        expect(await screen.findByText('completo')).toBeInTheDocument();
        expect(screen.queryByText('DOCUMENTS_VIEWER_PARTIAL')).toBeNull();
    });

    it('explains a protected or scanned non-PDF instead of asking for its text', async () => {
        routeRequests({
            items: [buildListItem(3, 'a.docx', '.docx')],
            indexState: { ...indexedState, state: 'name_only', reason: 'encrypted' },
        });

        renderViewer('/documents/view/3');

        expect(await screen.findByText('DOCUMENTS_VIEWER_PROTECTED')).toBeInTheDocument();
        expect(mockGet).not.toHaveBeenCalledWith('/documents/3/text');
    });

    it('explains a scanned document without text', async () => {
        routeRequests({
            items: [buildListItem(3, 'a.docx', '.docx')],
            indexState: { ...indexedState, state: 'name_only', reason: 'no_text' },
        });

        renderViewer('/documents/view/3');

        expect(await screen.findByText('DOCUMENTS_VIEWER_SCANNED')).toBeInTheDocument();
    });

    it('says there is no extracted text and shows an error when the text cannot load', async () => {
        routeRequests({ items: [buildListItem(3, 'a.docx', '.docx')] });

        renderViewer('/documents/view/3');

        expect(await screen.findByText('DOCUMENTS_VIEWER_NO_TEXT')).toBeInTheDocument();
    });

    it('shows an error when the extracted text fails', async () => {
        routeRequests({ items: [buildListItem(3, 'a.docx', '.docx')] });
        const baseImplementation = mockGet.getMockImplementation();
        mockGet.mockImplementation((url: string, config?: unknown) =>
            url === '/documents/3/text'
                ? Promise.reject(new Error('boom'))
                : baseImplementation?.(url, config)
        );

        renderViewer('/documents/view/3');

        expect(await screen.findByText('DOCUMENTS_VIEWER_TEXT_ERROR')).toBeInTheDocument();
    });

    it('renders a Markdown file as formatted content', async () => {
        routeRequests({
            items: [buildListItem(4, 'leia.md', '.md')],
            textPreview: '# Ola\n\n[bom](https://a.com) [mau](javascript:alert(1))',
        });

        renderViewer('/documents/view/4');

        expect(await screen.findByRole('heading', { level: 2, name: 'Ola' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'bom' })).toBeInTheDocument();
        expect(screen.queryByRole('link', { name: 'mau' })).toBeNull();
    });

    it('renders a CSV file as a table', async () => {
        routeRequests({
            items: [buildListItem(5, 'dados.csv', '.csv')],
            textPreview: 'nome,valor\nluz,10',
        });

        renderViewer('/documents/view/5');

        expect(await screen.findByRole('columnheader', { name: 'nome' })).toBeInTheDocument();
        expect(screen.getByRole('cell', { name: 'luz' })).toBeInTheDocument();
    });

    it('renders a TXT file with the text viewer', async () => {
        routeRequests({
            items: [buildListItem(6, 'nota.txt', '.txt')],
            textPreview: 'conteudo simples',
        });

        renderViewer('/documents/view/6');

        expect(await screen.findByText('conteudo simples')).toBeInTheDocument();
    });

    describe('previous and next', () => {
        const items = [
            buildListItem(1, 'a.pdf'),
            buildListItem(2, 'b.pdf'),
            buildListItem(3, 'c.pdf'),
        ];

        it('walks the list order and keeps the list URL', async () => {
            routeRequests({ items });

            renderViewer('/documents/view/2?from=format%3Dpdf');

            await screen.findByTitle('b.pdf');
            await waitFor(() =>
                expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' })).toBeEnabled()
            );

            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' }));
            expect(await screen.findByTitle('c.pdf')).toBeInTheDocument();
            expect(currentLocation()).toBe('/documents/view/3?from=format%3Dpdf');
            expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' })).toBeDisabled();

            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_PREVIOUS' }));
            fireEvent.click(
                await screen.findByRole('button', { name: 'DOCUMENTS_VIEWER_PREVIOUS' })
            );
            expect(await screen.findByTitle('a.pdf')).toBeInTheDocument();
            expect(currentLocation()).toBe('/documents/view/1?from=format%3Dpdf');
            expect(
                screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_PREVIOUS' })
            ).toBeDisabled();
        });

        it('requests the same filters and sort carried in from', async () => {
            routeRequests({ items });

            renderViewer('/documents/view/2?from=format%3Dpdf%26sort%3Dname%26order%3Dasc');

            await screen.findByTitle('b.pdf');
            expect(mockGet).toHaveBeenCalledWith(
                '/documents',
                expect.objectContaining({
                    params: expect.objectContaining({ format: 'pdf', sort: 'name', order: 'asc' }),
                })
            );
        });

        it('fetches the next page when the document is at the edge of the loaded pages', async () => {
            routeRequests({ items, secondPageItems: [buildListItem(4, 'd.pdf')] });

            renderViewer('/documents/view/3');

            await screen.findByTitle('c.pdf');
            await waitFor(() =>
                expect(mockGet).toHaveBeenCalledWith(
                    '/documents',
                    expect.objectContaining({ params: expect.objectContaining({ page: 2 }) })
                )
            );
            await waitFor(() =>
                expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' })).toBeEnabled()
            );
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' }));
            expect(await screen.findByTitle('d.pdf')).toBeInTheDocument();
        });

        it('keeps looking through the pages for a document that is not on the first one', async () => {
            routeRequests({ items, secondPageItems: [buildListItem(4, 'd.pdf')] });
            const baseImplementation = mockGet.getMockImplementation();
            mockGet.mockImplementation((url: string, config?: unknown) =>
                url === '/files/location/4'
                    ? Promise.reject(new Error('not needed'))
                    : baseImplementation?.(url, config)
            );

            renderViewer('/documents/view/4');

            expect(await screen.findByTitle('d.pdf')).toBeInTheDocument();
            await waitFor(() =>
                expect(
                    screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_PREVIOUS' })
                ).toBeEnabled()
            );
        });

        it('moves between search results using each result page', async () => {
            const buildResult = (fileId: number, name: string, page: number | null) => ({
                file_id: fileId,
                name,
                path: `/Casa/${name}`,
                parent_path: '/Casa',
                format: '.pdf',
                size: 10,
                updated_at: '2026-03-01T10:00:00Z',
                snippet: '',
                page,
            });
            routeRequests({
                searchItems: [buildResult(1, 'a.pdf', 2), buildResult(2, 'b.pdf', 5)],
            });

            renderViewer('/documents/view/1?from=q%3Dcontrato&page=2');

            await screen.findByTitle('a.pdf');
            await waitFor(() =>
                expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' })).toBeEnabled()
            );
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' }));

            const frame = await screen.findByTitle('b.pdf');
            expect(frame.getAttribute('src')).toMatch(/#page=5$/);
            expect(currentLocation()).toBe('/documents/view/2?from=q%3Dcontrato&page=5');
        });
    });

    describe('document not on the list', () => {
        it('loads it through its location and path', async () => {
            routeRequests({ items: [] });
            const baseImplementation = mockGet.getMockImplementation();
            mockGet.mockImplementation((url: string, config?: unknown) => {
                if (url === '/files/location/8') {
                    return Promise.resolve({ data: { logical_path: '/Casa/solto.pdf' } });
                }
                if (url === '/files/path') {
                    return Promise.resolve({
                        data: {
                            items: [
                                {
                                    id: 8,
                                    name: 'solto.pdf',
                                    path: '/Casa/solto.pdf',
                                    parent_path: '/Casa',
                                    format: '.pdf',
                                    size: 1,
                                    starred: false,
                                },
                            ],
                        },
                    });
                }
                return baseImplementation?.(url, config);
            });

            renderViewer('/documents/view/8');

            expect(await screen.findByTitle('solto.pdf')).toBeInTheDocument();
            expect(mockGet).toHaveBeenCalledWith('/files/path', {
                params: { path: '/Casa/solto.pdf' },
            });
            expect(
                screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_PREVIOUS' })
            ).toBeDisabled();
            expect(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_NEXT' })).toBeDisabled();
        });

        it('shows an error when the path resolves to nothing', async () => {
            routeRequests({ items: [] });
            const baseImplementation = mockGet.getMockImplementation();
            mockGet.mockImplementation((url: string, config?: unknown) => {
                if (url === '/files/location/8') {
                    return Promise.resolve({ data: { logical_path: '/Casa/solto.pdf' } });
                }
                if (url === '/files/path') return Promise.resolve({ data: { items: [] } });
                return baseImplementation?.(url, config);
            });

            renderViewer('/documents/view/8');

            expect(await screen.findByText('DOCUMENTS_VIEWER_LOAD_ERROR')).toBeInTheDocument();
        });
    });

    describe('toolbar', () => {
        beforeEach(() => {
            routeRequests({ items: [buildListItem(1, 'a.pdf')] });
        });

        it('goes back to the same list the user came from', async () => {
            routeRequests({
                searchItems: [
                    {
                        file_id: 1,
                        name: 'a.pdf',
                        path: '/Casa/a.pdf',
                        parent_path: '/Casa',
                        format: '.pdf',
                        size: 1,
                        updated_at: '2026-03-01T10:00:00Z',
                        snippet: '',
                        page: 3,
                    },
                ],
            });
            renderViewer('/documents/view/1?from=format%3Dpdf%26q%3Dcontrato&page=3');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_BACK' }));

            expect(currentLocation()).toBe('/documents?format=pdf&q=contrato');
        });

        it('goes back to the plain list when there is no list URL', async () => {
            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_BACK' }));

            expect(currentLocation()).toBe('/documents');
        });

        it('opens the parent folder in Files', async () => {
            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_OPEN_FOLDER' }));

            expect(currentLocation()).toBe('/files/Casa');
        });

        it('favorites the document through the batch endpoint', async () => {
            mockPost.mockResolvedValue({ data: { updated_ids: [1], not_found_ids: [] } });

            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'FILES_FAVORITE_ITEM' }));

            await waitFor(() =>
                expect(mockPost).toHaveBeenCalledWith('/files/starred', {
                    ids: [1],
                    starred: true,
                })
            );
        });

        it('unfavorites a starred document', async () => {
            routeRequests({ items: [buildListItem(1, 'a.pdf', '.pdf', true)] });
            mockPost.mockResolvedValue({ data: { updated_ids: [1], not_found_ids: [] } });

            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'FILES_UNFAVORITE_ITEM' }));

            await waitFor(() =>
                expect(mockPost).toHaveBeenCalledWith('/files/starred', {
                    ids: [1],
                    starred: false,
                })
            );
        });
    });

    describe('details and reindex', () => {
        it('shows the index situation with state, indexed bytes and reason', async () => {
            routeRequests({
                items: [buildListItem(1, 'a.pdf')],
                indexState: {
                    file_id: 1,
                    state: 'failed',
                    indexed_bytes: 0,
                    reason: 'timeout',
                    error: 'Falha ao ler o arquivo',
                },
            });

            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_DETAILS' }));

            expect(await screen.findByText('DOCUMENTS_INDEX_STATE_FAILED')).toBeInTheDocument();
            expect(screen.getByText('DOCUMENTS_INDEX_REASON_TIMEOUT')).toBeInTheDocument();
            expect(screen.getByText('Falha ao ler o arquivo')).toBeInTheDocument();
        });

        it('shows an error when the index situation cannot be loaded', async () => {
            routeRequests({ items: [buildListItem(3, 'a.docx', '.docx')] });
            const baseImplementation = mockGet.getMockImplementation();
            mockGet.mockImplementation((url: string, config?: unknown) =>
                url.endsWith('/index-state')
                    ? Promise.reject(new Error('boom'))
                    : baseImplementation?.(url, config)
            );

            renderViewer('/documents/view/3');

            await waitFor(() => expect(mockGet).toHaveBeenCalledWith('/documents/3/index-state'));
            fireEvent.click(
                await screen.findByRole('button', { name: 'DOCUMENTS_VIEWER_DETAILS' })
            );

            expect(await screen.findByText('DOCUMENTS_INDEX_LOAD_ERROR')).toBeInTheDocument();
        });

        it('reindexes, tells the user and reloads the index situation', async () => {
            routeRequests({ items: [buildListItem(1, 'a.pdf')] });
            mockPost.mockResolvedValue({ data: { job_id: 12 } });

            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_DETAILS' }));
            const callsBefore = mockGet.mock.calls.filter(
                ([url]) => url === '/documents/1/index-state'
            ).length;
            fireEvent.click(await screen.findByRole('button', { name: 'DOCUMENTS_REINDEX' }));

            await waitFor(() => expect(mockPost).toHaveBeenCalledWith('/documents/1/reindex'));
            expect(await screen.findByText('DOCUMENTS_REINDEX_SUCCESS')).toBeInTheDocument();
            await waitFor(() =>
                expect(
                    mockGet.mock.calls.filter(([url]) => url === '/documents/1/index-state').length
                ).toBeGreaterThan(callsBefore)
            );
        });

        it('shows the backend message when reindexing fails', async () => {
            routeRequests({ items: [buildListItem(1, 'a.pdf')] });
            mockPost.mockRejectedValue({
                response: { data: { error: 'Reindexação indisponível' } },
            });

            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_DETAILS' }));
            fireEvent.click(await screen.findByRole('button', { name: 'DOCUMENTS_REINDEX' }));

            expect(await screen.findByText('Reindexação indisponível')).toBeInTheDocument();
        });

        it('falls back to the generic message when the failure has none', async () => {
            routeRequests({ items: [buildListItem(1, 'a.pdf')] });
            mockPost.mockRejectedValue(new Error('boom'));

            renderViewer('/documents/view/1');

            await screen.findByTitle('a.pdf');
            fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_VIEWER_DETAILS' }));
            fireEvent.click(await screen.findByRole('button', { name: 'DOCUMENTS_REINDEX' }));

            expect(await screen.findByText('DOCUMENTS_REINDEX_ERROR')).toBeInTheDocument();
        });
    });
});

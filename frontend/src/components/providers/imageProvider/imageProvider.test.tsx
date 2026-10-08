import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { buildImageLibraryItem } from '@/components/imageContent/imageLibraryTestFixtures';
import { apiBase } from '@/service';
import { ImageProvider, useImage } from './imageProvider';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

const mockedApiGet = apiBase.get as jest.Mock;

const libraryCalls = () =>
    mockedApiGet.mock.calls
        .filter(([url]) => url === '/image/library')
        .map(([, config]) => config.params);

const callsTo = (url: string) => mockedApiGet.mock.calls.filter(([calledUrl]) => calledUrl === url);

function Consumer() {
    const { items, status, total, timeline, hasNextPage, fetchNextPage, isFetchNextPageError } =
        useImage();
    return (
        <div>
            <span data-testid="ids">{items.map((item) => item.file_id).join(',')}</span>
            <span data-testid="status">{status}</span>
            <span data-testid="total">{String(total)}</span>
            <span data-testid="timeline">{timeline.length}</span>
            <span data-testid="has-next">{String(hasNextPage)}</span>
            <span data-testid="next-page-error">{String(isFetchNextPageError)}</span>
            <button type="button" onClick={() => fetchNextPage()}>
                next
            </button>
        </div>
    );
}

const renderProvider = (route: string) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <MemoryRouter initialEntries={[route]}>
                <ImageProvider>
                    <Consumer />
                </ImageProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );

const routeResponses = (pages: unknown[]) => {
    let pageIndex = 0;
    mockedApiGet.mockImplementation((url: string) => {
        if (url === '/image/library') {
            return Promise.resolve({ data: pages[Math.min(pageIndex++, pages.length - 1)] });
        }
        if (url === '/image/library/count') {
            return Promise.resolve({ data: { total: 3 } });
        }
        if (url === '/image/library/timeline') {
            return Promise.resolve({ data: [{ year: 2026, month: 3, count: 3 }] });
        }
        return Promise.reject(new Error(`unexpected ${url}`));
    });
};

describe('providers/imageProvider', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('loads the first library page, the total and the timeline', async () => {
        routeResponses([
            {
                items: [buildImageLibraryItem({ file_id: 1 })],
                next_cursor: '',
                has_next: false,
                page_size: 60,
            },
        ]);

        renderProvider('/images');

        await waitFor(() => expect(screen.getByTestId('ids')).toHaveTextContent('1'));
        await waitFor(() => expect(screen.getByTestId('total')).toHaveTextContent('3'));
        await waitFor(() => expect(screen.getByTestId('timeline')).toHaveTextContent('1'));
        expect(libraryCalls()[0]).toEqual(
            expect.objectContaining({
                sort: 'taken_at',
                order: 'desc',
                page_size: 60,
                cursor: undefined,
                page: undefined,
                taken_before: undefined,
            })
        );
        expect(screen.getByTestId('has-next')).toHaveTextContent('false');
    });

    it('continues with the server cursor and never repeats taken_before after the first page', async () => {
        routeResponses([
            {
                items: [buildImageLibraryItem({ file_id: 1 })],
                next_cursor: 'cursor-1',
                has_next: true,
                page_size: 60,
            },
            {
                items: [buildImageLibraryItem({ file_id: 2 })],
                next_cursor: '',
                has_next: false,
                page_size: 60,
            },
        ]);

        renderProvider('/images?before=2026-04-01');

        await waitFor(() => expect(screen.getByTestId('ids')).toHaveTextContent('1'));
        expect(libraryCalls()[0]).toEqual(expect.objectContaining({ taken_before: '2026-04-01' }));

        await userEvent.click(screen.getByRole('button', { name: 'next' }));

        await waitFor(() => expect(screen.getByTestId('ids')).toHaveTextContent('1,2'));
        expect(libraryCalls()[1]).toEqual(
            expect.objectContaining({
                cursor: 'cursor-1',
                taken_before: undefined,
                page: undefined,
            })
        );
        expect(screen.getByTestId('has-next')).toHaveTextContent('false');
    });

    it('stops paging when the server reports more pages but no cursor', async () => {
        routeResponses([
            {
                items: [buildImageLibraryItem({ file_id: 1 })],
                next_cursor: '',
                has_next: true,
                page_size: 60,
            },
        ]);

        renderProvider('/images');

        await waitFor(() => expect(screen.getByTestId('ids')).toHaveTextContent('1'));
        expect(screen.getByTestId('has-next')).toHaveTextContent('false');
    });

    it('pages by number and skips the timeline when sorting by name', async () => {
        routeResponses([
            {
                items: [buildImageLibraryItem({ file_id: 1 })],
                next_cursor: '',
                has_next: true,
                page: 1,
                page_size: 60,
            },
            {
                items: [buildImageLibraryItem({ file_id: 2 })],
                next_cursor: '',
                has_next: false,
                page: 2,
                page_size: 60,
            },
        ]);

        renderProvider('/images?sort=name');

        await waitFor(() => expect(screen.getByTestId('ids')).toHaveTextContent('1'));
        expect(libraryCalls()[0]).toEqual(
            expect.objectContaining({ sort: 'name', order: 'asc', page: 1, cursor: undefined })
        );

        await userEvent.click(screen.getByRole('button', { name: 'next' }));

        await waitFor(() => expect(screen.getByTestId('ids')).toHaveTextContent('1,2'));
        expect(libraryCalls()[1]).toEqual(expect.objectContaining({ page: 2, cursor: undefined }));
        expect(callsTo('/image/library/timeline')).toHaveLength(0);
    });

    it('sends every active filter to the server', async () => {
        routeResponses([{ items: [], next_cursor: '', has_next: false, page_size: 60 }]);

        renderProvider(
            '/images/captures?q=beach&from=2026-01-01&to=2026-02-01&format=jpg&format=png'
        );

        await waitFor(() => expect(libraryCalls()).toHaveLength(1));
        expect(libraryCalls()[0]).toEqual(
            expect.objectContaining({
                q: 'beach',
                category: ['capture', 'screenshot_app'],
                format: ['jpg', 'png'],
                taken_from: '2026-01-01',
                taken_to: '2026-02-01',
            })
        );
        await waitFor(() => expect(callsTo('/image/library/count')).toHaveLength(1));
        expect(callsTo('/image/library/count')[0]![1].params).toEqual(
            expect.objectContaining({ q: 'beach', category: ['capture', 'screenshot_app'] })
        );
    });

    it('asks only for starred images on the favorites section', async () => {
        routeResponses([{ items: [], next_cursor: '', has_next: false, page_size: 60 }]);

        renderProvider('/images/favorites');

        await waitFor(() => expect(libraryCalls()).toHaveLength(1));
        expect(libraryCalls()[0]).toEqual(expect.objectContaining({ starred: true }));
    });

    it('filters by folder when a folder is selected', async () => {
        routeResponses([{ items: [], next_cursor: '', has_next: false, page_size: 60 }]);

        renderProvider('/images/folders?folder=%2Fphotos%2Ftrip');

        await waitFor(() => expect(libraryCalls()).toHaveLength(1));
        expect(libraryCalls()[0]).toEqual(expect.objectContaining({ folder: '/photos/trip' }));
    });

    it('does not query the gallery while the album picker is shown', async () => {
        routeResponses([{ items: [], next_cursor: '', has_next: false, page_size: 60 }]);

        renderProvider('/images/albums');

        await act(async () => {
            await Promise.resolve();
        });
        expect(mockedApiGet).not.toHaveBeenCalled();
        expect(screen.getByTestId('status')).toHaveTextContent('pending');
    });

    it('queries the album categories once an album is selected', async () => {
        routeResponses([{ items: [], next_cursor: '', has_next: false, page_size: 60 }]);

        renderProvider('/images/albums?album=documents');

        await waitFor(() => expect(libraryCalls()).toHaveLength(1));
        expect(libraryCalls()[0]).toEqual(
            expect.objectContaining({ category: ['document', 'receipt'] })
        );
    });

    it('exposes an error status and keeps loaded items when the next page fails', async () => {
        let callCount = 0;
        mockedApiGet.mockImplementation((url: string) => {
            if (url === '/image/library') {
                callCount += 1;
                return callCount === 1
                    ? Promise.resolve({
                          data: {
                              items: [buildImageLibraryItem({ file_id: 1 })],
                              next_cursor: 'cursor-1',
                              has_next: true,
                              page_size: 60,
                          },
                      })
                    : Promise.reject(new Error('boom'));
            }
            return Promise.resolve({ data: url.endsWith('count') ? { total: 1 } : [] });
        });

        renderProvider('/images');

        await waitFor(() => expect(screen.getByTestId('ids')).toHaveTextContent('1'));
        await userEvent.click(screen.getByRole('button', { name: 'next' }));

        await waitFor(() => expect(screen.getByTestId('status')).toHaveTextContent('error'));
        expect(screen.getByTestId('next-page-error')).toHaveTextContent('true');
        expect(screen.getByTestId('ids')).toHaveTextContent('1');
    });

    it('throws when used outside the provider', () => {
        const consoleError = jest.spyOn(console, 'error').mockImplementation(() => undefined);
        expect(() => render(<Consumer />)).toThrow('useImage must be used within');
        consoleError.mockRestore();
    });
});

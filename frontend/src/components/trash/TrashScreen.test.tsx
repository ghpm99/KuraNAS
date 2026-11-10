import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import TrashScreen from './TrashScreen';
import {
    deleteTrashItem,
    emptyTrash,
    getTrashItems,
    getTrashRetention,
    getTrashSummary,
    restoreTrashItem,
} from '@/service/trash';
import type { TrashPage } from '@/types/trash';

jest.mock('@/service/trash', () => ({
    getTrashItems: jest.fn(),
    getTrashRetention: jest.fn(),
    getTrashSummary: jest.fn(),
    restoreTrashItem: jest.fn(),
    deleteTrashItem: jest.fn(),
    emptyTrash: jest.fn(),
    updateTrashRetention: jest.fn(),
}));

const mockedGetTrashItems = getTrashItems as jest.Mock;
const mockedGetRetention = getTrashRetention as jest.Mock;
const mockedGetSummary = getTrashSummary as jest.Mock;
const mockedRestore = restoreTrashItem as jest.Mock;
const mockedDelete = deleteTrashItem as jest.Mock;
const mockedEmpty = emptyTrash as jest.Mock;

const onePage = (hasNext = false): TrashPage => ({
    items: [
        {
            id: 1,
            original_path: '/data/docs/relatorio.pdf',
            size: 2048,
            deleted_at: '2026-06-11T10:00:00Z',
        },
    ],
    pagination: { page: 1, page_size: 15, has_next: hasNext, has_prev: false },
});

const renderScreen = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <MemoryRouter>
                    <TrashScreen />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('TrashScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetRetention.mockResolvedValue({ days: 30 });
        mockedGetSummary.mockResolvedValue({ items: 1, size_bytes: 2048 });
    });

    it('shows the loading state while fetching', () => {
        mockedGetTrashItems.mockReturnValue(new Promise(() => {}));
        renderScreen();
        expect(screen.getByText('TRASH_LOADING')).toBeInTheDocument();
    });

    it('shows the error state when the request fails', async () => {
        mockedGetTrashItems.mockRejectedValue(new Error('boom'));
        renderScreen();
        await waitFor(() => expect(screen.getByText('TRASH_LOAD_ERROR')).toBeInTheDocument());
    });

    it('shows the empty state when the trash has no items', async () => {
        mockedGetTrashItems.mockResolvedValue({
            items: [],
            pagination: { page: 1, page_size: 15, has_next: false, has_prev: false },
        });
        renderScreen();
        await waitFor(() => expect(screen.getByText('TRASH_EMPTY_STATE')).toBeInTheDocument());
    });

    it('lists items and restores one on click', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedRestore.mockResolvedValue(undefined);
        renderScreen();

        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));
        await waitFor(() => expect(mockedRestore.mock.calls[0]?.[0]).toBe(1));
    });

    it('shows the backend error verbatim when a restore conflicts', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedRestore.mockRejectedValue({
            response: { data: { error: 'mensagem do servidor' } },
        });
        renderScreen();

        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));
        await waitFor(() => expect(screen.getByText('mensagem do servidor')).toBeInTheDocument());
    });

    it('shows the action fallback, not the loading error, when a restore fails without a server message', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedRestore.mockRejectedValue(new Error('network down'));
        renderScreen();
        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));

        await waitFor(() => expect(screen.getByText('TRASH_ACTION_ERROR')).toBeInTheDocument());
        expect(screen.queryByText('TRASH_LOAD_ERROR')).not.toBeInTheDocument();
    });

    const confirmInDialog = (confirmLabel: RegExp) =>
        fireEvent.click(
            within(screen.getByRole('dialog')).getByRole('button', { name: confirmLabel })
        );

    it('asks for confirmation in a dialog before deleting forever', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedDelete.mockResolvedValue(undefined);
        renderScreen();
        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ }));
        expect(screen.getByText('TRASH_DELETE_CONFIRM')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /ACTION_CANCEL/ }));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        expect(mockedDelete).not.toHaveBeenCalled();

        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ }));
        confirmInDialog(/TRASH_DELETE_FOREVER_BUTTON/);
        await waitFor(() => expect(mockedDelete.mock.calls[0]?.[0]).toBe(1));
    });

    it('empties the trash after confirming in the dialog', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedEmpty.mockResolvedValue(undefined);
        renderScreen();
        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: /TRASH_EMPTY_BUTTON/ }));
        expect(screen.getByText('TRASH_EMPTY_CONFIRM')).toBeInTheDocument();
        confirmInDialog(/TRASH_EMPTY_BUTTON/);

        await waitFor(() => expect(mockedEmpty).toHaveBeenCalled());
    });

    it('navigates between pages', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage(true));
        renderScreen();

        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: /TRASH_NEXT_PAGE/ }));
        await waitFor(() =>
            expect(mockedGetTrashItems).toHaveBeenCalledWith(2, 15, {
                sortKey: 'deleted_at',
                order: 'desc',
                searchText: '',
            })
        );
    });

    it('goes back to the previous page after restoring the last item of the current one', async () => {
        const lastPage: TrashPage = {
            items: [
                {
                    id: 2,
                    original_path: '/data/ultimo.txt',
                    size: 1,
                    deleted_at: '2026-06-11T10:00:00Z',
                },
            ],
            pagination: { page: 2, page_size: 15, has_next: false, has_prev: true },
        };
        const emptiedLastPage: TrashPage = { ...lastPage, items: [] };
        let isLastItemRestored = false;
        mockedGetTrashItems.mockImplementation((page: number) => {
            if (page === 1) {
                return Promise.resolve(onePage(!isLastItemRestored));
            }
            return Promise.resolve(isLastItemRestored ? emptiedLastPage : lastPage);
        });
        mockedRestore.mockImplementation(() => {
            isLastItemRestored = true;
            return Promise.resolve();
        });
        renderScreen();
        await waitFor(() => expect(screen.getByText('TRASH_NEXT_PAGE')).toBeEnabled());
        fireEvent.click(screen.getByText('TRASH_NEXT_PAGE'));
        await waitFor(() => expect(screen.getByText('ultimo.txt')).toBeInTheDocument());

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));

        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());
        expect(screen.queryByText('TRASH_EMPTY_STATE')).not.toBeInTheDocument();
    });

    it('asks the server for the chosen order and goes back to the first page', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage(true));
        renderScreen();
        await waitFor(() => expect(screen.getByText('TRASH_NEXT_PAGE')).toBeEnabled());
        fireEvent.click(screen.getByText('TRASH_NEXT_PAGE'));

        fireEvent.click(screen.getByRole('button', { name: 'TRASH_SORT_ORDER_DESCENDING' }));

        await waitFor(() =>
            expect(mockedGetTrashItems).toHaveBeenLastCalledWith(1, 15, {
                sortKey: 'deleted_at',
                order: 'asc',
                searchText: '',
            })
        );
    });

    it('searches on the server and tells when nothing matches', async () => {
        mockedGetTrashItems.mockImplementation(
            (_page: number, _pageSize: number, listing: { searchText: string }) =>
                Promise.resolve(
                    listing.searchText === ''
                        ? onePage()
                        : {
                              items: [],
                              pagination: {
                                  page: 1,
                                  page_size: 15,
                                  has_next: false,
                                  has_prev: false,
                              },
                          }
                )
        );
        renderScreen();
        await waitFor(() => expect(screen.getByText('relatorio.pdf')).toBeInTheDocument());

        fireEvent.change(screen.getByLabelText('TRASH_SEARCH_LABEL'), {
            target: { value: ' inexistente ' },
        });

        await waitFor(() =>
            expect(screen.getByText('TRASH_SEARCH_NO_RESULTS')).toBeInTheDocument()
        );
        expect(mockedGetTrashItems).toHaveBeenLastCalledWith(1, 15, {
            sortKey: 'deleted_at',
            order: 'desc',
            searchText: 'inexistente',
        });
        expect(screen.queryByText('TRASH_EMPTY_STATE')).not.toBeInTheDocument();
    });
});

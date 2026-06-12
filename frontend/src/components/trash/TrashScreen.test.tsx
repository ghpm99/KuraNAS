import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import TrashScreen from './TrashScreen';
import {
    deleteTrashItem,
    emptyTrash,
    getTrashItems,
    getTrashRetention,
    restoreTrashItem,
    updateTrashRetention,
} from '@/service/trash';
import type { TrashPage } from '@/types/trash';

jest.mock('@/service/trash', () => ({
    getTrashItems: jest.fn(),
    getTrashRetention: jest.fn(),
    restoreTrashItem: jest.fn(),
    deleteTrashItem: jest.fn(),
    emptyTrash: jest.fn(),
    updateTrashRetention: jest.fn(),
}));

const mockedGetTrashItems = getTrashItems as jest.Mock;
const mockedGetRetention = getTrashRetention as jest.Mock;
const mockedRestore = restoreTrashItem as jest.Mock;
const mockedDelete = deleteTrashItem as jest.Mock;
const mockedEmpty = emptyTrash as jest.Mock;
const mockedUpdateRetention = updateTrashRetention as jest.Mock;

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

        await waitFor(() =>
            expect(screen.getByText('/data/docs/relatorio.pdf')).toBeInTheDocument()
        );

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));
        await waitFor(() => expect(mockedRestore.mock.calls[0]?.[0]).toBe(1));
    });

    it('shows the backend error verbatim when a restore conflicts', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedRestore.mockRejectedValue({
            response: { data: { error: 'mensagem do servidor' } },
        });
        renderScreen();

        await waitFor(() =>
            expect(screen.getByText('/data/docs/relatorio.pdf')).toBeInTheDocument()
        );

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));
        await waitFor(() => expect(screen.getByText('mensagem do servidor')).toBeInTheDocument());
    });

    it('asks for confirmation before deleting forever', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedDelete.mockResolvedValue(undefined);
        const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(false);
        renderScreen();

        await waitFor(() =>
            expect(screen.getByText('/data/docs/relatorio.pdf')).toBeInTheDocument()
        );

        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ }));
        expect(mockedDelete).not.toHaveBeenCalled();

        confirmSpy.mockReturnValue(true);
        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ }));
        await waitFor(() => expect(mockedDelete.mock.calls[0]?.[0]).toBe(1));

        confirmSpy.mockRestore();
    });

    it('empties the trash after confirmation', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedEmpty.mockResolvedValue(undefined);
        const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(true);
        renderScreen();

        await waitFor(() =>
            expect(screen.getByText('/data/docs/relatorio.pdf')).toBeInTheDocument()
        );

        fireEvent.click(screen.getByRole('button', { name: /TRASH_EMPTY_BUTTON/ }));
        await waitFor(() => expect(mockedEmpty).toHaveBeenCalled());

        confirmSpy.mockRestore();
    });

    it('updates the retention policy on blur', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        mockedUpdateRetention.mockResolvedValue({ days: 7 });
        renderScreen();

        const input = await screen.findByLabelText('TRASH_RETENTION_LABEL');
        fireEvent.change(input, { target: { value: '7' } });
        fireEvent.blur(input);

        await waitFor(() => expect(mockedUpdateRetention.mock.calls[0]?.[0]).toBe(7));
    });

    it('does not save a non-positive retention', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage());
        renderScreen();

        const input = await screen.findByLabelText('TRASH_RETENTION_LABEL');
        fireEvent.change(input, { target: { value: '0' } });
        fireEvent.blur(input);

        await waitFor(() => expect(mockedUpdateRetention).not.toHaveBeenCalled());
    });

    it('navigates between pages', async () => {
        mockedGetTrashItems.mockResolvedValue(onePage(true));
        renderScreen();

        await waitFor(() =>
            expect(screen.getByText('/data/docs/relatorio.pdf')).toBeInTheDocument()
        );

        fireEvent.click(screen.getByRole('button', { name: /TRASH_NEXT_PAGE/ }));
        await waitFor(() => expect(mockedGetTrashItems).toHaveBeenCalledWith(2, 15));
    });
});

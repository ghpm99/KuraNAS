import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import TrashScreen from './TrashScreen';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
        delete: jest.fn(),
        put: jest.fn(),
    },
}));

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
    delete: jest.Mock;
    put: jest.Mock;
};

const onePage = {
    items: [
        {
            id: 1,
            original_path: '/data/docs/relatorio.pdf',
            size: 2048,
            deleted_at: '2026-06-11T10:00:00Z',
        },
    ],
    pagination: { page: 1, page_size: 15, has_next: false, has_prev: false },
};

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

describe('TrashScreen (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/trash') return Promise.resolve({ data: onePage });
            if (url === '/trash/retention') return Promise.resolve({ data: { days: 30 } });
            if (url === '/trash/summary')
                return Promise.resolve({ data: { items: 1, size_bytes: 2048 } });
            return Promise.reject(new Error(`unexpected GET ${url}`));
        });
        mockedApi.post.mockResolvedValue({ data: undefined });
        mockedApi.delete.mockResolvedValue({ data: undefined });
        mockedApi.put.mockResolvedValue({ data: { days: 7 } });
    });

    const confirmInDialog = (confirmLabel: RegExp) =>
        fireEvent.click(
            within(screen.getByRole('dialog')).getByRole('button', { name: confirmLabel })
        );

    it('restores an item via POST /trash/:id/restore', async () => {
        renderScreen();
        await screen.findByText('relatorio.pdf');

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/trash/1/restore'));
    });

    it('permanently deletes an item via DELETE /trash/:id', async () => {
        renderScreen();
        await screen.findByText('relatorio.pdf');

        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ }));
        confirmInDialog(/TRASH_DELETE_FOREVER_BUTTON/);

        await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/trash/1'));
    });

    it('empties the trash via DELETE /trash', async () => {
        renderScreen();
        await screen.findByText('relatorio.pdf');

        fireEvent.click(screen.getByRole('button', { name: /TRASH_EMPTY_BUTTON/ }));
        confirmInDialog(/TRASH_EMPTY_BUTTON/);

        await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/trash'));
    });

    it('saves the retention policy on blur via PUT /trash/retention', async () => {
        renderScreen();
        const input = await screen.findByLabelText('TRASH_RETENTION_LABEL');
        await waitFor(() => expect(input).toHaveValue(30));

        fireEvent.change(input, { target: { value: '7' } });
        fireEvent.blur(input);

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/trash/retention', { days: 7 })
        );
    });

    const selectOnlyItem = async () => {
        await screen.findByText('relatorio.pdf');
        fireEvent.click(screen.getByRole('checkbox', { name: 'TRASH_SELECT_ITEM' }));
    };

    it('restores the selection via POST /trash/restore', async () => {
        mockedApi.post.mockResolvedValue({ data: { restored_ids: [1], failures: [] } });
        renderScreen();
        await selectOnlyItem();

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_SELECTED_BUTTON/ }));

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/trash/restore', { ids: [1] })
        );
        expect(await screen.findByText('TRASH_RESTORED_TOAST')).toBeInTheDocument();
    });

    it('permanently deletes the selection via POST /trash/delete after confirming', async () => {
        mockedApi.post.mockResolvedValue({ data: { deleted_ids: [1], failures: [] } });
        renderScreen();
        await selectOnlyItem();

        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_SELECTED_BUTTON/ }));
        expect(mockedApi.post).not.toHaveBeenCalled();
        expect(screen.getByText('TRASH_DELETE_SELECTED_CONFIRM')).toBeInTheDocument();
        confirmInDialog(/TRASH_DELETE_SELECTED_BUTTON/);

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/trash/delete', { ids: [1] })
        );
    });

    it('shows the server reason and keeps the item selected when the batch leaves it behind', async () => {
        mockedApi.post.mockResolvedValue({
            data: { restored_ids: [], failures: [{ id: 1, error: 'local ocupado' }] },
        });
        renderScreen();
        await selectOnlyItem();

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_SELECTED_BUTTON/ }));

        expect(await screen.findByText('local ocupado')).toBeInTheDocument();
        expect(screen.getByRole('checkbox', { name: 'TRASH_SELECT_ITEM' })).toBeChecked();
    });

    it('falls back to the trash action error when the batch request itself fails', async () => {
        mockedApi.post.mockRejectedValue(new Error('network down'));
        renderScreen();
        await selectOnlyItem();

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_SELECTED_BUTTON/ }));

        expect(await screen.findByText('TRASH_ACTION_ERROR')).toBeInTheDocument();
    });

    it('survives a batch answer without the result lists', async () => {
        mockedApi.post.mockResolvedValue({ data: {} });
        renderScreen();
        await selectOnlyItem();

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_SELECTED_BUTTON/ }));

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/trash/restore', { ids: [1] })
        );
        expect(screen.getByText('relatorio.pdf')).toBeInTheDocument();
    });

    it('selects the whole page and clears the selection', async () => {
        renderScreen();
        await screen.findByText('relatorio.pdf');

        fireEvent.click(screen.getByRole('button', { name: 'TRASH_SELECT_ALL' }));
        expect(screen.getByRole('checkbox', { name: 'TRASH_SELECT_ITEM' })).toBeChecked();

        fireEvent.click(screen.getByRole('button', { name: 'TRASH_CLEAR_SELECTION' }));
        expect(screen.getByRole('checkbox', { name: 'TRASH_SELECT_ITEM' })).not.toBeChecked();
    });
});

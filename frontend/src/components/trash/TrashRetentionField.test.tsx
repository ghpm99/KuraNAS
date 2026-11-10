import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import TrashRetentionField from './TrashRetentionField';
import { apiBase } from '@/service';
import { trashQueryKey } from '@/shared/queryKeys/fileQueryKeys';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';

const renderField = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>
                <TrashRetentionField />
            </SnackbarProvider>
        </QueryClientProvider>
    );
    return queryClient;
};

const findRetentionInput = () => screen.findByLabelText('TRASH_RETENTION_LABEL');

const typeAndLeave = (retentionInput: HTMLElement, typedDays: string) => {
    fireEvent.change(retentionInput, { target: { value: typedDays } });
    fireEvent.blur(retentionInput);
};

describe('TrashRetentionField', () => {
    afterEach(() => {
        jest.restoreAllMocks();
    });

    it('renders an editable field without backend', async () => {
        renderWithoutBackend(<TrashRetentionField />);

        const retentionInput = await screen.findByRole('spinbutton');
        await waitFor(() => expect(retentionInput).toBeEnabled());
        expect(retentionInput).toHaveValue(null);
    });

    it('shows the saved retention', async () => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { days: 30 } });
        renderField();

        await waitFor(async () => expect(await findRetentionInput()).toHaveValue(30));
    });

    it('tells the retention is unknown when it cannot be read and still saves a new one', async () => {
        jest.spyOn(apiBase, 'get').mockRejectedValue(new Error('offline'));
        const putRetention = jest.spyOn(apiBase, 'put').mockResolvedValue({ data: { days: 7 } });
        renderField();
        expect(await screen.findByText('TRASH_RETENTION_UNKNOWN')).toBeInTheDocument();

        typeAndLeave(await findRetentionInput(), '7');

        await waitFor(() =>
            expect(putRetention).toHaveBeenCalledWith('/trash/retention', { days: 7 })
        );
        await waitFor(() =>
            expect(screen.queryByText('TRASH_RETENTION_UNKNOWN')).not.toBeInTheDocument()
        );
        expect(await findRetentionInput()).toHaveValue(7);
    });

    it.each(['0', '-3', '1.5'])('flags %s as invalid and does not save it', async (typedDays) => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { days: 30 } });
        const putRetention = jest.spyOn(apiBase, 'put');
        renderField();
        const retentionInput = await findRetentionInput();
        await waitFor(() => expect(retentionInput).toHaveValue(30));

        typeAndLeave(retentionInput, typedDays);

        expect(await screen.findByText('TRASH_RETENTION_INVALID')).toBeInTheDocument();
        expect(putRetention).not.toHaveBeenCalled();
    });

    it('restores the saved retention when the field is left blank', async () => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { days: 30 } });
        const putRetention = jest.spyOn(apiBase, 'put');
        renderField();
        const retentionInput = await findRetentionInput();
        await waitFor(() => expect(retentionInput).toHaveValue(30));

        typeAndLeave(retentionInput, '');

        expect(retentionInput).toHaveValue(30);
        expect(putRetention).not.toHaveBeenCalled();
    });

    it('does not save an unchanged retention', async () => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { days: 30 } });
        const putRetention = jest.spyOn(apiBase, 'put');
        renderField();
        const retentionInput = await findRetentionInput();
        await waitFor(() => expect(retentionInput).toHaveValue(30));

        fireEvent.blur(retentionInput);

        expect(putRetention).not.toHaveBeenCalled();
    });

    it('follows a retention changed elsewhere', async () => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { days: 30 } });
        const queryClient = renderField();
        const retentionInput = await findRetentionInput();
        await waitFor(() => expect(retentionInput).toHaveValue(30));

        queryClient.setQueryData([trashQueryKey, 'retention'], { days: 90 });

        await waitFor(() => expect(retentionInput).toHaveValue(90));
    });

    it('shows the server message and keeps the typed value when saving fails', async () => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { days: 30 } });
        jest.spyOn(apiBase, 'put').mockRejectedValue({
            response: { data: { error: 'mensagem do servidor' } },
        });
        renderField();
        const retentionInput = await findRetentionInput();
        await waitFor(() => expect(retentionInput).toHaveValue(30));

        typeAndLeave(retentionInput, '7');

        expect(await screen.findByText('mensagem do servidor')).toBeInTheDocument();
        expect(retentionInput).toHaveValue(7);
    });

    it('falls back to the trash action error when saving fails without a server message', async () => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { days: 30 } });
        jest.spyOn(apiBase, 'put').mockRejectedValue(new Error('network down'));
        renderField();
        const retentionInput = await findRetentionInput();
        await waitFor(() => expect(retentionInput).toHaveValue(30));

        typeAndLeave(retentionInput, '7');

        expect(await screen.findByText('TRASH_ACTION_ERROR')).toBeInTheDocument();
    });
});

import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { apiBase } from '@/service';
import FinishedIngestDownloads from './FinishedIngestDownloads';
import { activeIngestDownloadsQueryKey } from './ingestQueryKeys';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const finishedDownload = (overrides: Record<string, unknown>) => ({
    id: 1,
    job_id: 3,
    url: 'https://example.com/video',
    title: 'First video',
    preset: 'best',
    target_root_id: 1,
    target_root_name: 'Videos',
    subfolder: 'Clips',
    state: 'completed',
    progress: 100,
    speed_bytes_per_second: null,
    eta_seconds: null,
    file_id: 10,
    error: '',
    created_at: '2026-11-25T10:00:00Z',
    ended_at: '2026-11-25T10:05:00Z',
    ...overrides,
});

const pageOf = (items: unknown[], page: number, hasNext: boolean) => ({
    items,
    pagination: { page, page_size: 20, has_next: hasNext, has_prev: page > 1 },
});

const renderHistory = () => {
    const client = new QueryClient({
        defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    const invalidateQueries = jest.spyOn(client, 'invalidateQueries');
    const view = render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <MemoryRouter>
                    <FinishedIngestDownloads />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
    return { ...view, invalidateQueries };
};

describe('FinishedIngestDownloads', () => {
    afterEach(() => jest.clearAllMocks());

    it('renders without a backend', () => {
        mockedApi.get.mockRejectedValue(new Error('offline'));
        expect(() => renderWithoutBackend(<FinishedIngestDownloads />)).not.toThrow();
    });

    it('requests finished downloads with an explicit page size', async () => {
        mockedApi.get.mockResolvedValue({ data: pageOf([finishedDownload({})], 1, false) });
        renderHistory();
        await screen.findByText('First video');
        expect(mockedApi.get).toHaveBeenCalledWith('/ingest/downloads', {
            params: { page: 1, page_size: 20, state: 'finished' },
        });
        expect(screen.getByText('INGEST_HISTORY_END')).toBeInTheDocument();
    });

    it('loads the second page and ignores duplicated ids', async () => {
        mockedApi.get
            .mockResolvedValueOnce({ data: pageOf([finishedDownload({})], 1, true) })
            .mockResolvedValueOnce({
                data: pageOf(
                    [finishedDownload({}), finishedDownload({ id: 2, title: 'Second video' })],
                    2,
                    false
                ),
            });
        renderHistory();
        await screen.findByText('First video');
        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));
        await screen.findByText('Second video');
        expect(screen.getAllByText('First video')).toHaveLength(1);
        expect(mockedApi.get).toHaveBeenLastCalledWith('/ingest/downloads', {
            params: { page: 2, page_size: 20, state: 'finished' },
        });
    });

    it('keeps the loaded items when the next page fails and retries it', async () => {
        mockedApi.get
            .mockResolvedValueOnce({ data: pageOf([finishedDownload({})], 1, true) })
            .mockRejectedValueOnce({ response: { data: { error: 'server broke' } } })
            .mockResolvedValueOnce({
                data: pageOf([finishedDownload({ id: 2, title: 'Second video' })], 2, false),
            });
        renderHistory();
        await screen.findByText('First video');
        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));
        await screen.findByText('server broke');
        expect(screen.getByText('First video')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));
        await screen.findByText('Second video');
    });

    it('shows the server message when the first load fails', async () => {
        mockedApi.get.mockRejectedValue({ response: { data: { error: 'history unavailable' } } });
        renderHistory();
        expect(await screen.findByText('history unavailable')).toBeInTheDocument();
    });

    it('shows the empty state', async () => {
        mockedApi.get.mockResolvedValue({ data: pageOf([], 1, false) });
        renderHistory();
        expect(await screen.findByText('INGEST_HISTORY_EMPTY')).toBeInTheDocument();
    });

    it('disables the open actions without a file id and shows the failure reason', async () => {
        mockedApi.get.mockResolvedValue({
            data: pageOf(
                [
                    finishedDownload({ id: 1, file_id: null }),
                    finishedDownload({
                        id: 2,
                        title: 'Broken video',
                        state: 'failed',
                        file_id: null,
                        error: 'Download failed badly',
                    }),
                ],
                1,
                false
            ),
        });
        renderHistory();
        await screen.findByText('Broken video');
        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'INGEST_HISTORY_OPEN' })).toBeDisabled()
        );
        expect(screen.getByRole('button', { name: 'INGEST_HISTORY_OPEN_FOLDER' })).toBeDisabled();
        expect(screen.getByText('Download failed badly')).toBeInTheDocument();
        expect(screen.getByText(/INGEST_HISTORY_STATE_FAILED/)).toBeInTheDocument();
    });

    describe('retry', () => {
        const failedDownload = finishedDownload({ id: 2, title: 'Failed video', state: 'failed' });
        const canceledDownload = finishedDownload({
            id: 3,
            title: 'Canceled video',
            state: 'canceled',
        });
        const retryButtons = () => screen.getAllByRole('button', { name: 'INGEST_HISTORY_RETRY' });
        const retryButtonAt = (index: number): HTMLElement => {
            const button = retryButtons()[index];
            if (!button) throw new Error(`retry button ${index} not found`);
            return button;
        };
        const finishedCalls = () =>
            mockedApi.get.mock.calls.filter(
                ([url, config]) => url === '/ingest/downloads' && config.params.state === 'finished'
            ).length;

        it('shows the button only on failed and canceled items', async () => {
            mockedApi.get.mockResolvedValue({
                data: pageOf([finishedDownload({}), failedDownload, canceledDownload], 1, false),
            });
            renderHistory();
            await screen.findByText('Failed video');
            expect(retryButtons()).toHaveLength(2);
        });

        it('posts the retry, disables only the clicked item and refetches the lists', async () => {
            let finishRetry: (value: unknown) => void = () => undefined;
            mockedApi.post.mockReturnValue(
                new Promise((resolve) => {
                    finishRetry = resolve;
                })
            );
            mockedApi.get.mockResolvedValue({
                data: pageOf([failedDownload, canceledDownload], 1, false),
            });
            const { invalidateQueries } = renderHistory();
            await screen.findByText('Failed video');
            const buttons = retryButtons();
            fireEvent.click(buttons[0] as HTMLElement);

            await waitFor(() => expect(buttons[0]).toBeDisabled());
            expect(buttons[1]).not.toBeDisabled();
            expect(mockedApi.post).toHaveBeenCalledWith('/ingest/downloads/2/retry');

            const finishedBefore = finishedCalls();
            finishRetry({ data: { ...failedDownload, id: 9, state: 'queued' } });
            await waitFor(() => expect(finishedCalls()).toBeGreaterThan(finishedBefore));
            await waitFor(() => expect(buttons[0]).not.toBeDisabled());
            expect(invalidateQueries).toHaveBeenCalledWith({
                queryKey: activeIngestDownloadsQueryKey,
            });
        });

        it('shows the server message when the retry fails', async () => {
            mockedApi.post.mockRejectedValue({ response: { data: { error: 'cannot retry now' } } });
            mockedApi.get.mockResolvedValue({ data: pageOf([failedDownload], 1, false) });
            renderHistory();
            await screen.findByText('Failed video');
            fireEvent.click(retryButtonAt(0));
            expect(await screen.findByText('cannot retry now')).toBeInTheDocument();
        });

        it('falls back to the translated message without a server message', async () => {
            mockedApi.post.mockRejectedValue(new Error('offline'));
            mockedApi.get.mockResolvedValue({ data: pageOf([failedDownload], 1, false) });
            renderHistory();
            await screen.findByText('Failed video');
            fireEvent.click(retryButtonAt(0));
            expect(await screen.findByText('INGEST_HISTORY_RETRY_ERROR')).toBeInTheDocument();
        });
    });
});

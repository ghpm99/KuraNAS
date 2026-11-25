import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { apiBase } from '@/service';
import ActiveIngestDownloads from './ActiveIngestDownloads';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const activeDownload = {
    id: 7,
    job_id: 3,
    url: 'https://example.com/video',
    title: null,
    preset: 'best',
    target_root_id: 1,
    target_root_name: 'Videos',
    subfolder: 'Clips',
    state: 'downloading',
    progress: 42,
    speed_bytes_per_second: 2048,
    eta_seconds: 125,
    file_id: null,
    error: '',
    created_at: '2026-11-25T10:00:00Z',
    ended_at: null,
};

const pageOf = (items: unknown[]) => ({
    items,
    pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
});

const mockApi = (downloadsResponse: () => unknown) => {
    mockedApi.get.mockImplementation((path: string) =>
        path === '/ingest/downloads'
            ? Promise.resolve({ data: downloadsResponse() })
            : Promise.resolve({ data: [{ key: 'best', label: 'Best quality' }] })
    );
};

const downloadsCalls = () =>
    mockedApi.get.mock.calls.filter(([path]: [string]) => path === '/ingest/downloads').length;

const renderBlock = () => {
    const client = new QueryClient({
        defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    return render(
        <QueryClientProvider client={client}>
            <ActiveIngestDownloads />
        </QueryClientProvider>
    );
};

describe('ActiveIngestDownloads', () => {
    afterEach(() => {
        jest.useRealTimers();
        jest.clearAllMocks();
    });

    it('renders without a backend', () => {
        mockedApi.get.mockRejectedValue(new Error('offline'));
        expect(() => renderWithoutBackend(<ActiveIngestDownloads />)).not.toThrow();
    });

    it('shows destination, format, progress, speed and remaining time', async () => {
        mockApi(() => pageOf([activeDownload]));
        renderBlock();

        expect(await screen.findByText('https://example.com/video')).toBeDefined();
        expect(screen.getByText(/INGEST_ACTIVE_DESTINATION/)).toBeDefined();
        expect(screen.getByText('42%')).toBeDefined();
        expect(screen.getByText(/INGEST_ACTIVE_SPEED/)).toBeDefined();
        expect(screen.getByText(/INGEST_ACTIVE_ETA/)).toBeDefined();
        expect(mockedApi.get).toHaveBeenCalledWith('/ingest/downloads', {
            params: { page: 1, page_size: 50, state: 'active' },
        });
    });

    it('shows the server message when loading fails', async () => {
        mockedApi.get.mockRejectedValue({ response: { data: { error: 'server says no' } } });
        renderBlock();

        expect(await screen.findByText('server says no')).toBeDefined();
    });

    it('polls every 2s while active and stops once the list is empty', async () => {
        jest.useFakeTimers();
        let items: unknown[] = [activeDownload];
        mockApi(() => pageOf(items));
        renderBlock();

        await waitFor(() => expect(downloadsCalls()).toBe(1));
        await jest.advanceTimersByTimeAsync(2000);
        await waitFor(() => expect(downloadsCalls()).toBe(2));

        items = [];
        await jest.advanceTimersByTimeAsync(2000);
        await waitFor(() => expect(downloadsCalls()).toBe(3));

        await jest.advanceTimersByTimeAsync(10000);
        expect(downloadsCalls()).toBe(3);
    });
});

import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import IngestForm from './IngestForm';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const targets = [{ label: 'Videos', path: '/data/videos' }];
const presets = [{ key: 'best', label: 'Best quality' }];

const mockBackend = (ytDlpStatusResponse: () => Promise<unknown>) => {
    mockedApi.get.mockImplementation((path: string) => {
        if (path === '/ingest/ytdlp/status') return ytDlpStatusResponse();
        return Promise.resolve({ data: path === '/ingest/targets' ? targets : presets });
    });
};

const renderForm = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <IngestForm />
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const submitButton = () =>
    screen.getByRole('button', { name: 'INGEST_FORM_SUBMIT' }) as HTMLButtonElement;

const waitForTargetsLoaded = () =>
    waitFor(() => expect(mockedApi.get).toHaveBeenCalledWith('/ingest/ytdlp/status'));

describe('IngestForm yt-dlp status', () => {
    beforeEach(() => jest.clearAllMocks());

    it('warns, links to the yt-dlp settings and disables submit when yt-dlp is not installed', async () => {
        mockBackend(() => Promise.resolve({ data: { installed: false } }));
        renderForm();

        const link = await screen.findByRole('link', { name: 'INGEST_FORM_YTDLP_MISSING_LINK' });
        expect(link.getAttribute('href')).toBe('/settings#yt-dlp');
        expect(screen.getByRole('alert').textContent).toContain('INGEST_FORM_YTDLP_MISSING');
        expect(submitButton().disabled).toBe(true);
    });

    it('shows no warning and allows submit when yt-dlp is installed', async () => {
        mockBackend(() => Promise.resolve({ data: { installed: true } }));
        renderForm();

        await waitFor(() => expect(submitButton().disabled).toBe(false));
        expect(screen.queryByRole('link', { name: 'INGEST_FORM_YTDLP_MISSING_LINK' })).toBeNull();
    });

    it('does not block submit when the status request fails', async () => {
        mockBackend(() => Promise.reject(new Error('503')));
        renderForm();
        await waitForTargetsLoaded();

        await waitFor(() => expect(submitButton().disabled).toBe(false));
        expect(screen.queryByRole('link', { name: 'INGEST_FORM_YTDLP_MISSING_LINK' })).toBeNull();
    });

    it('does not block submit while the status is still loading', async () => {
        mockBackend(() => new Promise(() => undefined));
        renderForm();

        await waitFor(() => expect(submitButton().disabled).toBe(false));
    });
});

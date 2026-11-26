import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import IngestForm from './IngestForm';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const targets = [
    { label: 'Videos', path: '/data/videos' },
    { label: 'Music', path: '/data/music' },
];
const presets = [
    { key: 'best', label: 'Best quality' },
    { key: 'audio', label: 'Audio only' },
];

const renderForm = () => {
    const client = new QueryClient({
        defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <IngestForm />
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const urlInput = () => screen.getByLabelText('INGEST_FORM_URL_LABEL') as HTMLInputElement;

const typeUrl = (value: string) => fireEvent.change(urlInput(), { target: { value } });

const submitButton = () => screen.getByRole('button', { name: 'INGEST_FORM_SUBMIT' });

const waitForSubmitEnabled = () =>
    waitFor(() => expect((submitButton() as HTMLButtonElement).disabled).toBe(false));

describe('IngestForm (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((path: string) =>
            Promise.resolve({ data: path === '/ingest/targets' ? targets : presets })
        );
        mockedApi.post.mockResolvedValue({ data: { job_id: 1, download_id: 2 } });
    });

    it('submitting posts url, preset, target_root and subfolder to /ingest/fetch and clears the url', async () => {
        renderForm();
        await waitForSubmitEnabled();

        typeUrl('  https://example.com/watch?v=1  ');
        fireEvent.click(submitButton());

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/ingest/fetch', {
                url: 'https://example.com/watch?v=1',
                preset: 'best',
                target_root: '/data/videos',
                subfolder: '',
            })
        );
        await waitFor(() => expect(urlInput().value).toBe(''));
    });

    it('rejects a non http(s) scheme inline without calling the server', async () => {
        renderForm();
        await waitForSubmitEnabled();

        typeUrl('ftp://example.com/file');
        fireEvent.click(submitButton());

        expect(await screen.findByText('INGEST_FORM_URL_INVALID')).toBeDefined();
        expect(mockedApi.post).not.toHaveBeenCalled();
    });

    it('rejects an empty url inline', async () => {
        renderForm();
        await waitForSubmitEnabled();

        fireEvent.click(submitButton());

        expect(await screen.findByText('INGEST_FORM_URL_REQUIRED')).toBeDefined();
        expect(mockedApi.post).not.toHaveBeenCalled();
    });

    it('rejects a url longer than 2048 characters inline', async () => {
        renderForm();
        await waitForSubmitEnabled();

        typeUrl(`https://example.com/${'a'.repeat(2048)}`);
        fireEvent.click(submitButton());

        expect(await screen.findByText('INGEST_FORM_URL_TOO_LONG')).toBeDefined();
        expect(mockedApi.post).not.toHaveBeenCalled();
    });

    it('shows the server error verbatim and keeps the url', async () => {
        mockedApi.post.mockRejectedValue({ response: { data: { error: 'yt-dlp não instalado' } } });
        renderForm();
        await waitForSubmitEnabled();

        typeUrl('https://example.com/video');
        fireEvent.click(submitButton());

        expect(await screen.findByText('yt-dlp não instalado')).toBeDefined();
        expect(urlInput().value).toBe('https://example.com/video');
    });

    it('keeps destination and preset after a successful submit', async () => {
        renderForm();
        await waitForSubmitEnabled();

        fireEvent.mouseDown(screen.getByLabelText('INGEST_FORM_PRESET_LABEL'));
        fireEvent.click(await screen.findByRole('option', { name: 'Audio only' }));
        typeUrl('https://example.com/one');
        fireEvent.click(submitButton());
        await waitFor(() => expect(urlInput().value).toBe(''));

        typeUrl('https://example.com/two');
        fireEvent.click(submitButton());

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenLastCalledWith(
                '/ingest/fetch',
                expect.objectContaining({ preset: 'audio', target_root: '/data/videos' })
            )
        );
    });
});

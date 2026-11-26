import { MemoryRouter } from 'react-router-dom';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import CapturesScreen from './CapturesScreen';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const buildCapture = (id: number, status: string) => ({
    id,
    name: `recording_${id}`,
    file_name: `recording_${id}.mp4`,
    file_path: `/x/recording_${id}.mp4`,
    media_type: 'video',
    mime_type: 'video/mp4',
    size: 1024,
    episode_key: '',
    created_at: '2026-06-01T00:00:00Z',
    status,
});

const buildPage = (statuses: string[]) => ({
    items: statuses.map((status, index) => buildCapture(index + 1, status)),
    pagination: { page: 1, page_size: 20, has_next: false, has_prev: false },
});

const renderScreen = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <MemoryRouter>
                    <CapturesScreen />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('CapturesScreen retry promotion (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: buildPage(['failed', 'promoted', 'failed']) });
        mockedApi.post.mockResolvedValue({ data: { job_id: 'job-1' }, status: 202 });
    });

    it('shows the retry button only on failed captures', async () => {
        renderScreen();
        await screen.findByText('recording_1');
        expect(screen.getAllByRole('button', { name: 'CAPTURES_RETRY_PROMOTION' })).toHaveLength(2);
    });

    it('posts to /captures/:id/promote, announces success and reloads the list', async () => {
        renderScreen();
        await screen.findByText('recording_1');
        const listLoadsBefore = mockedApi.get.mock.calls.length;

        fireEvent.click(screen.getAllByRole('button', { name: 'CAPTURES_RETRY_PROMOTION' })[0]);

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/captures/1/promote'));
        expect(await screen.findByText('CAPTURE_RETRY_PROMOTION_STARTED')).toBeInTheDocument();
        await waitFor(() =>
            expect(mockedApi.get.mock.calls.length).toBeGreaterThan(listLoadsBefore)
        );
    });

    it('disables only the button of the capture being retried', async () => {
        mockedApi.post.mockReturnValue(new Promise(() => undefined));
        renderScreen();
        await screen.findByText('recording_1');

        fireEvent.click(screen.getAllByRole('button', { name: 'CAPTURES_RETRY_PROMOTION' })[0]);

        await waitFor(() => {
            const [firstButton, secondButton] = screen.getAllByRole('button', {
                name: 'CAPTURES_RETRY_PROMOTION',
            });
            expect(firstButton).toBeDisabled();
            expect(secondButton).toBeEnabled();
        });
    });

    it('shows the server message when the promotion is refused with 409', async () => {
        mockedApi.post.mockRejectedValue({
            isAxiosError: true,
            response: { status: 409, data: { error: 'A captura não está com falha.' } },
        });
        renderScreen();
        await screen.findByText('recording_1');

        fireEvent.click(screen.getAllByRole('button', { name: 'CAPTURES_RETRY_PROMOTION' })[0]);

        expect(await screen.findByText('A captura não está com falha.')).toBeInTheDocument();
    });
});

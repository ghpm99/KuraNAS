import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import CapturesScreen from './CapturesScreen';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const buildCapture = (id: number) => ({
    id,
    name: `recording_${id}`,
    file_name: `recording_${id}.mp4`,
    file_path: `/x/recording_${id}.mp4`,
    media_type: 'video',
    mime_type: 'video/mp4',
    size: 1024,
    episode_key: '',
    created_at: '2026-06-01T00:00:00Z',
    status: 'uploaded',
});

const respondWith = (captures: ReturnType<typeof buildCapture>[]) =>
    mockedApi.get.mockResolvedValue({
        data: {
            items: captures,
            pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
        },
    });

const renderScreen = (initialUrl = '/captures') => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <MemoryRouter initialEntries={[initialUrl]}>
                    <CapturesScreen />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

const lastCapturesParams = () => {
    const calls = mockedApi.get.mock.calls.filter(([url]) => url === '/captures');
    return calls[calls.length - 1][1].params;
};

const capturesCallCount = () =>
    mockedApi.get.mock.calls.filter(([url]) => url === '/captures').length;

describe('CapturesScreen filters (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        respondWith([buildCapture(1)]);
    });

    it('applies the filters from the URL and fills the fields', async () => {
        renderScreen('/captures?q=naruto&status=failed');
        await screen.findByText('recording_1');

        expect(lastCapturesParams()).toMatchObject({ name: 'naruto', status: 'failed' });
        expect(screen.getByRole('searchbox')).toHaveValue('naruto');
        expect(screen.getByRole('combobox')).toHaveTextContent('CAPTURES_STATUS_FAILED');
    });

    it('only requests the new name after 300ms, from page 1', async () => {
        renderScreen();
        await screen.findByText('recording_1');
        jest.useFakeTimers();
        try {
            const callsBeforeTyping = capturesCallCount();
            fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'bleach' } });
            act(() => {
                jest.advanceTimersByTime(250);
            });
            expect(capturesCallCount()).toBe(callsBeforeTyping);
            act(() => {
                jest.advanceTimersByTime(100);
            });
        } finally {
            jest.useRealTimers();
        }

        await waitFor(() =>
            expect(lastCapturesParams()).toMatchObject({ name: 'bleach', page: 1 })
        );
    });

    it('requests immediately when the status changes', async () => {
        renderScreen();
        await screen.findByText('recording_1');

        await userEvent.click(screen.getByRole('combobox'));
        await userEvent.click(
            await screen.findByRole('option', { name: 'CAPTURES_STATUS_PROMOTED' })
        );

        await waitFor(() => expect(lastCapturesParams()).toMatchObject({ status: 'promoted' }));
    });

    it('shows the filtered empty state and clears the filters', async () => {
        respondWith([]);
        renderScreen('/captures?q=zzz');
        expect(await screen.findByText('CAPTURES_FILTER_EMPTY')).toBeInTheDocument();
        expect(screen.queryByText('CAPTURES_EMPTY')).not.toBeInTheDocument();

        respondWith([buildCapture(1)]);
        await userEvent.click(screen.getByRole('button', { name: 'CAPTURES_FILTER_CLEAR' }));

        await screen.findByText('recording_1');
        expect(lastCapturesParams().name).toBeUndefined();
        expect(screen.getByRole('searchbox')).toHaveValue('');
    });
});

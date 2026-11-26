import { MemoryRouter } from 'react-router-dom';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import CapturesScreen from './CapturesScreen';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const backendCapture = (id: number) => ({
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

const backendPage = (ids: number[], page: number, hasNext: boolean) => ({
    data: {
        items: ids.map(backendCapture),
        pagination: { page, page_size: 50, has_next: hasNext, has_prev: page > 1 },
    },
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

describe('CapturesScreen infinite scroll', () => {
    beforeEach(() => jest.clearAllMocks());

    it('sends page_size 50, appends the next page, ignores a duplicate id and marks the end', async () => {
        mockedApi.get
            .mockResolvedValueOnce(backendPage([1, 2], 1, true))
            .mockResolvedValueOnce(backendPage([2, 3], 2, false));
        renderScreen();

        expect(await screen.findByText('recording_1')).toBeInTheDocument();
        expect(mockedApi.get.mock.calls[0][1].params).toMatchObject({ page: 1, page_size: 50 });

        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByText('recording_3')).toBeInTheDocument();
        expect(mockedApi.get.mock.calls[1][1].params).toMatchObject({ page: 2, page_size: 50 });
        expect(screen.getAllByText('recording_2')).toHaveLength(1);
        expect(screen.getByText('CAPTURES_END')).toBeInTheDocument();
    });

    it('keeps the loaded items and offers retry when the next page fails', async () => {
        mockedApi.get
            .mockResolvedValueOnce(backendPage([1], 1, true))
            .mockRejectedValueOnce({ response: { data: { error: 'falha na página' } } })
            .mockResolvedValueOnce(backendPage([2], 2, false));
        renderScreen();

        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByText('falha na página')).toBeInTheDocument();
        expect(screen.getByText('recording_1')).toBeInTheDocument();
        expect(screen.queryByText('CAPTURES_END')).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        expect(await screen.findByText('recording_2')).toBeInTheDocument();
        await waitFor(() => expect(screen.queryByText('falha na página')).not.toBeInTheDocument());
        expect(screen.getByText('recording_1')).toBeInTheDocument();
    });
});

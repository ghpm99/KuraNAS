import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import CapturesScreen from './CapturesScreen';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const buildCapture = (id: number, status: string, fileId?: number) => ({
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
    ...(fileId === undefined ? {} : { file_id: fileId }),
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

const mockCaptures = (captures: ReturnType<typeof buildCapture>[]) => {
    mockedApi.get.mockResolvedValue({
        data: {
            items: captures,
            pagination: { page: 1, page_size: 20, has_next: false, has_prev: false },
        },
    });
};

describe('CapturesScreen watch action (seam)', () => {
    beforeEach(() => jest.clearAllMocks());

    it('links a promoted capture with file_id to the video player', async () => {
        mockCaptures([buildCapture(1, 'promoted', 42)]);
        renderScreen();
        await screen.findByText('recording_1');
        expect(screen.getByRole('link', { name: 'CAPTURES_WATCH' })).toHaveAttribute(
            'href',
            '/video/42'
        );
    });

    it('hides the watch action when there is no playable file', async () => {
        mockCaptures([
            buildCapture(1, 'promoted'),
            buildCapture(2, 'failed'),
            buildCapture(3, 'uploaded'),
            buildCapture(4, 'promoting'),
        ]);
        renderScreen();
        await screen.findByText('recording_1');
        expect(screen.queryByText('CAPTURES_WATCH')).not.toBeInTheDocument();
    });
});

import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import CapturesScreen from './CapturesScreen';
import { getCaptures } from '@/service/captures';
import type { Capture } from '@/types/captures';

jest.mock('@/service/captures', () => ({
    getCaptures: jest.fn(),
    deleteCapture: jest.fn(),
    captureThumbnailHref: () => '',
}));

const mockedGetCaptures = getCaptures as jest.Mock;

const capture = (overrides: Partial<Capture>): Capture => ({
    id: 1,
    name: 'recording',
    file_name: 'recording.mp4',
    file_path: '/x/recording.mp4',
    media_type: 'video',
    mime_type: 'video/mp4',
    size: 1024,
    episode_key: '',
    created_at: '2026-06-01T00:00:00Z',
    status: 'promoted',
    ...overrides,
});

const renderScreen = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <CapturesScreen />
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('CapturesScreen reload', () => {
    const scrollIntoView = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
        window.HTMLElement.prototype.scrollIntoView = scrollIntoView;
    });

    it('scrolls to the top, refetches page 1 and shows an item created elsewhere', async () => {
        mockedGetCaptures.mockResolvedValueOnce({
            items: [capture({ id: 1, title: 'Antiga' })],
            pagination: {},
        });
        renderScreen();
        expect(await screen.findByText('Antiga')).toBeInTheDocument();

        mockedGetCaptures.mockResolvedValueOnce({
            items: [capture({ id: 2, title: 'Nova' }), capture({ id: 1, title: 'Antiga' })],
            pagination: {},
        });
        fireEvent.click(screen.getByRole('button', { name: /CAPTURES_RELOAD/ }));

        expect(scrollIntoView).toHaveBeenCalledWith({ block: 'start' });
        expect(await screen.findByText('Nova')).toBeInTheDocument();
        expect(mockedGetCaptures).toHaveBeenLastCalledWith({ page: 1, pageSize: 50 });
    });

    it('keeps the items and shows the server message when the reload fails', async () => {
        mockedGetCaptures.mockResolvedValueOnce({
            items: [capture({ id: 1, title: 'Antiga' })],
            pagination: {},
        });
        renderScreen();
        expect(await screen.findByText('Antiga')).toBeInTheDocument();

        mockedGetCaptures.mockRejectedValueOnce({
            response: { data: { error: 'banco indisponível' } },
        });
        fireEvent.click(screen.getByRole('button', { name: /CAPTURES_RELOAD/ }));

        expect(await screen.findByText('banco indisponível')).toBeInTheDocument();
        await waitFor(() => expect(screen.getByText('Antiga')).toBeInTheDocument());
    });
});

import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CapturesScreen from './CapturesScreen';
import { getCaptures, deleteCapture } from '@/service/captures';
import type { Capture } from '@/types/captures';

jest.mock('@/service/captures', () => ({
    getCaptures: jest.fn(),
    deleteCapture: jest.fn(),
    captureThumbnailHref: (capture: Capture) =>
        capture.status === 'promoted' && capture.file_id ? `/thumb/${capture.file_id}` : '',
}));

const mockedGetCaptures = getCaptures as jest.Mock;
const mockedDeleteCapture = deleteCapture as jest.Mock;

const capture = (overrides: Partial<Capture> = {}): Capture => ({
    id: 1,
    name: 'recording_1',
    file_name: 'recording_1.mp4',
    file_path: '/x/recording_1.mp4',
    media_type: 'video',
    mime_type: 'video/mp4',
    size: 5 * 1024 * 1024,
    episode_key: '',
    created_at: '2026-06-01T00:00:00Z',
    status: 'uploaded',
    ...overrides,
});

const renderScreen = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <CapturesScreen />
        </QueryClientProvider>
    );
};

describe('CapturesScreen', () => {
    beforeEach(() => jest.clearAllMocks());

    it('renders without crashing when the backend is absent (rejected query)', async () => {
        mockedGetCaptures.mockRejectedValue(new Error('network down'));
        renderScreen();
        await waitFor(() => expect(screen.getByText('CAPTURES_ERROR')).toBeInTheDocument());
        expect(screen.getByText('CAPTURES_PAGE_TITLE')).toBeInTheDocument();
    });

    it('survives a partial payload with no items array', async () => {
        mockedGetCaptures.mockResolvedValue({ pagination: {} });
        renderScreen();
        await waitFor(() => expect(screen.getByText('CAPTURES_EMPTY')).toBeInTheDocument());
    });

    it('shows the loading state while fetching', () => {
        mockedGetCaptures.mockReturnValue(new Promise(() => {}));
        renderScreen();
        expect(screen.getByText('CAPTURES_LOADING')).toBeInTheDocument();
    });

    it('renders capture rows with status, platform, episode and source link', async () => {
        mockedGetCaptures.mockResolvedValue({
            items: [
                capture({
                    id: 2,
                    title: 'Frieren',
                    episode_title: 'The Journey',
                    season: 1,
                    episode: 3,
                    platform: 'Crunchyroll',
                    status: 'promoted',
                    file_id: 42,
                    source_url: 'https://crunchyroll.com/x',
                }),
            ],
            pagination: {},
        });
        renderScreen();

        await waitFor(() => expect(screen.getByText('Frieren')).toBeInTheDocument());
        expect(screen.getByText('CAPTURES_STATUS_PROMOTED')).toBeInTheDocument();
        expect(screen.getByText('Crunchyroll')).toBeInTheDocument();
        expect(screen.getByText(/S1E3/)).toBeInTheDocument();
        expect(screen.getByText('CAPTURES_SOURCE_LINK').closest('a')).toHaveAttribute(
            'href',
            'https://crunchyroll.com/x'
        );
        // Promoted with file_id -> falls back to the video thumbnail image.
        expect(screen.getByRole('img', { name: 'Frieren' })).toHaveAttribute('src', '/thumb/42');
    });

    it('falls back to the capture name and an icon when no title/thumbnail', async () => {
        mockedGetCaptures.mockResolvedValue({ items: [capture()], pagination: {} });
        renderScreen();
        await waitFor(() => expect(screen.getByText('recording_1')).toBeInTheDocument());
        expect(screen.queryByRole('img')).not.toBeInTheDocument();
    });

    it('renders promoting and failed statuses and an episode without a season', async () => {
        mockedGetCaptures.mockResolvedValue({
            items: [
                capture({ id: 10, status: 'promoting', episode: 5 }),
                capture({ id: 11, name: 'rec_b', status: 'failed' }),
            ],
            pagination: {},
        });
        renderScreen();

        await waitFor(() =>
            expect(screen.getByText('CAPTURES_STATUS_PROMOTING')).toBeInTheDocument()
        );
        expect(screen.getByText('CAPTURES_STATUS_FAILED')).toBeInTheDocument();
        expect(screen.getByText('E5')).toBeInTheDocument();
    });

    it('deletes a capture when the delete button is clicked', async () => {
        mockedGetCaptures.mockResolvedValue({ items: [capture({ id: 5 })], pagination: {} });
        mockedDeleteCapture.mockResolvedValue(undefined);
        renderScreen();

        await waitFor(() => expect(screen.getByText('recording_1')).toBeInTheDocument());
        fireEvent.click(screen.getByLabelText('CAPTURES_DELETE'));
        await waitFor(() => expect(mockedDeleteCapture).toHaveBeenCalledWith(5));
    });
});

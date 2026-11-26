import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import CapturesScreen from './CapturesScreen';
import { apiBase } from '@/service';
import { buildFilesUrl } from '@/app/routes';

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

const CurrentLocation = () => {
    const location = useLocation();
    return <div data-testid="current-location">{location.pathname + location.search}</div>;
};

const renderScreen = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <MemoryRouter initialEntries={['/captures']}>
                    <Routes>
                        <Route path="/captures" element={<CapturesScreen />} />
                        <Route path="*" element={<CurrentLocation />} />
                    </Routes>
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

const mockCaptures = (
    captures: ReturnType<typeof buildCapture>[],
    locationResponse: () => Promise<unknown>
) => {
    mockedApi.get.mockImplementation((url: string) => {
        if (url.startsWith('/files/location/')) return locationResponse();
        return Promise.resolve({
            data: {
                items: captures,
                pagination: { page: 1, page_size: 20, has_next: false, has_prev: false },
            },
        });
    });
};

describe('CapturesScreen open folder action (seam)', () => {
    beforeEach(() => jest.clearAllMocks());

    it('opens the folder of the promoted video in Files', async () => {
        mockCaptures([buildCapture(1, 'promoted', 42)], () =>
            Promise.resolve({ data: { logical_path: '/Videos/Anime/ep1.mp4' } })
        );
        renderScreen();
        await screen.findByText('recording_1');

        await userEvent.click(screen.getByRole('button', { name: 'CAPTURES_OPEN_FOLDER' }));

        expect(mockedApi.get).toHaveBeenCalledWith('/files/location/42');
        const expectedUrl = buildFilesUrl('/Videos/Anime');
        await waitFor(() =>
            expect(screen.getByTestId('current-location')).toHaveTextContent(expectedUrl)
        );
    });

    it('shows the error only on the item whose location failed', async () => {
        mockCaptures([buildCapture(1, 'promoted', 42), buildCapture(2, 'promoted', 43)], () =>
            Promise.reject(new Error('boom'))
        );
        renderScreen();
        await screen.findByText('recording_1');

        await userEvent.click(screen.getAllByRole('button', { name: 'CAPTURES_OPEN_FOLDER' })[0]);

        expect(await screen.findAllByText('CAPTURES_OPEN_FOLDER_ERROR')).toHaveLength(1);
    });

    it('hides the open folder action without a promoted file', async () => {
        mockCaptures(
            [
                buildCapture(1, 'promoted'),
                buildCapture(2, 'failed', 5),
                buildCapture(3, 'uploaded'),
            ],
            () => Promise.reject(new Error('unused'))
        );
        renderScreen();
        await screen.findByText('recording_1');
        expect(screen.queryByText('CAPTURES_OPEN_FOLDER')).not.toBeInTheDocument();
    });
});

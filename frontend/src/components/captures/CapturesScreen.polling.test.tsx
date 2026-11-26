import { MemoryRouter } from 'react-router-dom';
import { act, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import CapturesScreen from './CapturesScreen';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

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

const buildPage = (captures: ReturnType<typeof buildCapture>[]) => ({
    data: {
        items: captures,
        pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
    },
});

const renderScreen = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>
                <MemoryRouter>
                    <CapturesScreen />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

const advanceTwoSeconds = async () => {
    await act(async () => {
        await jest.advanceTimersByTimeAsync(2050);
    });
};

describe('CapturesScreen processing polling', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        jest.clearAllMocks();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('refetches every 2s while a capture is promoting and updates it in place without the loading state', async () => {
        mockedApi.get
            .mockResolvedValueOnce(
                buildPage([buildCapture(1, 'promoting'), buildCapture(2, 'promoted')])
            )
            .mockResolvedValue(
                buildPage([buildCapture(1, 'promoted'), buildCapture(2, 'promoted')])
            );
        renderScreen();
        await act(async () => {
            await jest.advanceTimersByTimeAsync(0);
        });
        expect(screen.getByText('CAPTURES_STATUS_PROMOTING')).toBeTruthy();
        expect(mockedApi.get).toHaveBeenCalledTimes(1);

        await advanceTwoSeconds();

        expect(mockedApi.get).toHaveBeenCalledTimes(2);
        expect(screen.queryByText('CAPTURES_LOADING')).toBeNull();
        expect(screen.queryByText('CAPTURES_STATUS_PROMOTING')).toBeNull();
        expect(screen.getAllByText('CAPTURES_STATUS_PROMOTED')).toHaveLength(2);
        expect(screen.getAllByText(/^recording_\d$/)).toHaveLength(2);
    });

    it('never shows the loading state while refetching', async () => {
        let resolveRefetch: (value: unknown) => void = () => undefined;
        mockedApi.get
            .mockResolvedValueOnce(buildPage([buildCapture(1, 'uploaded')]))
            .mockImplementationOnce(() => new Promise((resolve) => (resolveRefetch = resolve)));
        renderScreen();
        await act(async () => {
            await jest.advanceTimersByTimeAsync(0);
        });

        await advanceTwoSeconds();

        expect(mockedApi.get).toHaveBeenCalledTimes(2);
        expect(screen.queryByText('CAPTURES_LOADING')).toBeNull();
        expect(screen.getByText('recording_1')).toBeTruthy();
        await act(async () => {
            resolveRefetch(buildPage([buildCapture(1, 'promoted')]));
        });
    });

    it('stops polling once nothing is uploaded or promoting', async () => {
        mockedApi.get.mockResolvedValue(
            buildPage([buildCapture(1, 'promoted'), buildCapture(2, 'failed')])
        );
        renderScreen();
        await act(async () => {
            await jest.advanceTimersByTimeAsync(0);
        });

        await advanceTwoSeconds();
        await advanceTwoSeconds();

        expect(mockedApi.get).toHaveBeenCalledTimes(1);
    });
});

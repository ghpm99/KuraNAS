import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';
import type { IMusicData } from '@/types/music';
import TrackStarButton from './TrackStarButton';
import { clearAllStarredOverrides } from './trackStarOverrides';

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params?.name ? `${key}:${params.name}` : key,
    }),
}));

const mockedApi = apiBase as unknown as { post: jest.Mock };

let serverStarred = false;
const fetchTrackFromServer = jest.fn(async () => ({
    id: 9,
    name: 'nine.mp3',
    starred: serverStarred,
}));

const TrackFromServer = () => {
    const { data: track } = useQuery({
        queryKey: ['music-reconcile'],
        queryFn: fetchTrackFromServer,
    });
    return track ? <TrackStarButton track={track as IMusicData} /> : null;
};

const renderWithClient = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
        <QueryClientProvider client={queryClient}>
            <TrackFromServer />
        </QueryClientProvider>
    );
    return queryClient;
};

const starButton = () => screen.findByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' });
const unstarButton = () => screen.findByRole('button', { name: 'MUSIC_TRACK_UNSTAR:nine.mp3' });

describe('TrackStarButton (server reconciliation)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        clearAllStarredOverrides();
        serverStarred = false;
        mockedApi.post.mockResolvedValue({ data: {} });
    });

    it('shows the server value after a toggle once the queries are refetched', async () => {
        const queryClient = renderWithClient();
        fireEvent.click(await starButton());
        serverStarred = true;
        await unstarButton();
        await waitFor(() => expect(fetchTrackFromServer).toHaveBeenCalledTimes(2));

        serverStarred = false;
        await queryClient.refetchQueries({ queryKey: ['music-reconcile'] });

        expect(await starButton()).toHaveAttribute('aria-pressed', 'false');
    });

    it('flips immediately and keeps the value while the refetch settles', async () => {
        renderWithClient();
        fireEvent.click(await starButton());

        expect(screen.getByRole('button', { name: 'MUSIC_TRACK_UNSTAR:nine.mp3' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
    });

    it('rolls back to the server value when the request fails', async () => {
        mockedApi.post.mockRejectedValue(new Error('offline'));
        renderWithClient();
        fireEvent.click(await starButton());

        await waitFor(async () =>
            expect(await starButton()).toHaveAttribute('aria-pressed', 'false')
        );
    });
});

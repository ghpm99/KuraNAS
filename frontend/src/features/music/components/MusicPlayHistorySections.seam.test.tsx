import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { apiBase } from '@/service';
import MusicPlayHistorySections from './MusicPlayHistorySections';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

const mockedReplaceQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useOptionalGlobalMusic: () => ({ replaceQueue: mockedReplaceQueue }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) => {
            if (params?.count) return `${key}:${params.count}`;
            if (params?.name) return `${key}:${params.name}`;
            return key;
        },
    }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const playedTrack = (id: number, title: string, playCount: number) => ({
    track: {
        id,
        name: `${title}.mp3`,
        path: `/music/${title}.mp3`,
        metadata: { title, artist: 'The Band', length: 200 },
    },
    play_count: playCount,
    last_played_at: '2026-04-01T10:00:00Z',
});

const renderSections = () =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <MusicPlayHistorySections />
        </QueryClientProvider>
    );

describe('MusicPlayHistorySections (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/music/library/most-played') {
                return Promise.resolve({
                    data: {
                        items: [playedTrack(1, 'Hit', 12), playedTrack(2, 'Second', 5)],
                        pagination: { page: 1, page_size: 6, has_next: false, has_prev: false },
                    },
                });
            }
            return Promise.resolve({
                data: {
                    items: [playedTrack(3, 'Fresh', 1)],
                    pagination: { page: 1, page_size: 6, has_next: false, has_prev: false },
                },
            });
        });
    });

    it('requests the most played and recent plays endpoints with their params', async () => {
        renderSections();

        await screen.findByText('Hit');
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/most-played', {
            params: { page: 1, page_size: 6, period: 'all' },
        });
        expect(mockedApi.get).toHaveBeenCalledWith('/music/library/recent-plays', {
            params: { page: 1, page_size: 6 },
        });
        expect(screen.getByText('MUSIC_HOME_PLAY_COUNT:12')).toBeInTheDocument();
        expect(screen.getByText('Fresh')).toBeInTheDocument();
    });

    it('plays the row starting at the chosen card', async () => {
        renderSections();

        await screen.findByText('Second');
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_HOME_PLAY_TRACK:Second' }));

        await waitFor(() => expect(mockedReplaceQueue).toHaveBeenCalledTimes(1));
        const [queuedTracks, startIndex] = mockedReplaceQueue.mock.calls[0]!;
        expect(queuedTracks.map((track: { id: number }) => track.id)).toEqual([1, 2]);
        expect(startIndex).toBe(1);
    });

    it('plays the recently played row from its own tracks', async () => {
        renderSections();

        await screen.findByText('Fresh');
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_HOME_PLAY_TRACK:Fresh' }));

        const [queuedTracks, startIndex] = mockedReplaceQueue.mock.calls[0]!;
        expect(queuedTracks.map((track: { id: number }) => track.id)).toEqual([3]);
        expect(startIndex).toBe(0);
    });
});

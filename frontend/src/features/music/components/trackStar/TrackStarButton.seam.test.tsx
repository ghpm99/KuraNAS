import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
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

const track = { id: 9, name: 'nine.mp3', starred: false } as IMusicData;

describe('TrackStarButton (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        clearAllStarredOverrides();
        mockedApi.post.mockResolvedValue({ data: {} });
    });

    it('posts to /files/starred/:id and flips the star optimistically', async () => {
        render(<TrackStarButton track={track} />);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' }));

        expect(screen.getByRole('button', { name: 'MUSIC_TRACK_UNSTAR:nine.mp3' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/files/starred/9'));
        expect(mockedApi.post).toHaveBeenCalledTimes(1);
    });

    it('toggles back off on a second click', async () => {
        render(<TrackStarButton track={track} />);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' }));
        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(1));
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_TRACK_UNSTAR:nine.mp3' }));

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(2));
        expect(screen.getByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' })).toHaveAttribute(
            'aria-pressed',
            'false'
        );
    });

    it('reverts the optimistic star when the request fails', async () => {
        mockedApi.post.mockRejectedValue(new Error('offline'));
        render(<TrackStarButton track={track} />);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' }));

        await waitFor(() =>
            expect(
                screen.getByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' })
            ).toHaveAttribute('aria-pressed', 'false')
        );
    });

    it('shares the starred state between two buttons of the same track', async () => {
        render(
            <>
                <TrackStarButton track={track} />
                <TrackStarButton track={{ ...track }} />
            </>
        );

        fireEvent.click(screen.getAllByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' })[0]!);

        expect(screen.getAllByRole('button', { name: 'MUSIC_TRACK_UNSTAR:nine.mp3' })).toHaveLength(
            2
        );
        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(1));
    });

    it('invalidates music lists and the favorites playlist after the toggle', async () => {
        const queryClient = new QueryClient();
        const invalidateQueries = jest.spyOn(queryClient, 'invalidateQueries');
        render(
            <QueryClientProvider client={queryClient}>
                <TrackStarButton track={track} />
            </QueryClientProvider>
        );

        await act(async () => {
            fireEvent.click(screen.getByRole('button', { name: 'MUSIC_TRACK_STAR:nine.mp3' }));
        });

        await waitFor(() => expect(invalidateQueries).toHaveBeenCalledTimes(1));
        const { predicate } = invalidateQueries.mock.calls[0]![0]!;
        const matches = (rootKey: string) => predicate!({ queryKey: [rootKey] } as never);
        expect(matches('music-home')).toBe(true);
        expect(matches('music-by-album')).toBe(true);
        expect(matches('automatic-playlists')).toBe(true);
        expect(matches('playlist-tracks')).toBe(true);
        expect(matches('files')).toBe(false);
    });
});

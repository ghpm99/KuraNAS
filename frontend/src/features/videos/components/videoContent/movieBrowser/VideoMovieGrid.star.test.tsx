import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { SnackbarProvider } from 'notistack';
import { apiBase } from '@/service';
import type { VideoFileDto } from '@/service/videoPlayback';
import VideoMovieGrid from './VideoMovieGrid';

jest.mock('@/service', () => ({
    apiBase: { put: jest.fn() },
}));

const movieQueryKey = ['videos', 'library-movies', 'name'];

const initialMovies: VideoFileDto[] = [
    {
        id: 4,
        name: 'Alien.mkv',
        path: '/m/Alien.mkv',
        parent_path: '/m',
        format: 'mkv',
        size: 1,
    },
];

function CachedMovieGrid({ onPlayVideo }: { onPlayVideo: jest.Mock }) {
    const { data: movies = [] } = useQuery<VideoFileDto[]>({
        queryKey: movieQueryKey,
        queryFn: () => initialMovies,
        initialData: initialMovies,
    });
    return (
        <VideoMovieGrid
            movies={movies}
            hasNextPage={false}
            isFetchingNextPage={false}
            fetchNextPage={jest.fn()}
            onPlayVideo={onPlayVideo}
        />
    );
}

describe('VideoMovieGrid star', () => {
    it('toggles aria-pressed from the video cache without starting playback', async () => {
        (apiBase.put as jest.Mock).mockResolvedValue({ data: { id: 4, starred: true } });
        const onPlayVideo = jest.fn();
        render(
            <QueryClientProvider client={new QueryClient()}>
                <SnackbarProvider>
                    <CachedMovieGrid onPlayVideo={onPlayVideo} />
                </SnackbarProvider>
            </QueryClientProvider>
        );

        const starButton = screen.getByRole('button', {
            name: 'FILES_FAVORITE_ITEM',
            pressed: false,
        });
        fireEvent.click(starButton);

        await waitFor(() =>
            expect(
                screen.getByRole('button', { name: 'FILES_UNFAVORITE_ITEM', pressed: true })
            ).toBeTruthy()
        );
        expect(apiBase.put).toHaveBeenCalledWith('/files/starred/4', { starred: true });
        expect(onPlayVideo).not.toHaveBeenCalled();
    });
});

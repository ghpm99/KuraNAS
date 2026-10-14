import { render, screen } from '@testing-library/react';
import type { VideoFileDto } from '@/service/videoPlayback';
import VideoMovieGrid from './VideoMovieGrid';

const buildMovie = (overrides: Partial<VideoFileDto> = {}): VideoFileDto => ({
    id: 1,
    name: 'Movie',
    path: '/movies/movie.mkv',
    parent_path: '/movies',
    format: '.mkv',
    size: 10,
    ...overrides,
});

const renderGrid = (movies: VideoFileDto[]) =>
    render(
        <VideoMovieGrid
            movies={movies}
            hasNextPage={false}
            isFetchingNextPage={false}
            fetchNextPage={jest.fn()}
            onPlayVideo={jest.fn()}
        />
    );

describe('VideoMovieGrid', () => {
    it('renders a movie without technical fields and shows no badges', () => {
        renderGrid([buildMovie()]);

        expect(screen.getByText('Movie')).toBeInTheDocument();
        expect(screen.queryByText(/VIDEO_BADGE_RESOLUTION/)).not.toBeInTheDocument();
    });

    it('shows duration and resolution badges', () => {
        renderGrid([buildMovie({ duration_seconds: 7200, height: 2160 })]);

        expect(screen.getByText('2:00:00')).toBeInTheDocument();
        expect(screen.getByText('VIDEO_BADGE_RESOLUTION_UHD')).toBeInTheDocument();
    });
});

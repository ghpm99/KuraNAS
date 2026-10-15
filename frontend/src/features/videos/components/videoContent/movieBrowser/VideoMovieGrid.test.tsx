import { fireEvent, render, screen } from '@testing-library/react';
import type { VideoFileDto } from '@/service/videoPlayback';
import VideoMovieGrid from './VideoMovieGrid';
import { resetFailedPosterUrls } from '../useLocalPoster';

jest.mock('@/features/videos/hooks/useVideoStarToggle/useVideoStarToggle', () => ({
    useVideoStarToggle: () => ({ toggleStar: jest.fn(), isStarTogglePending: false }),
}));

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
    beforeEach(() => resetFailedPosterUrls());

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

    it('requests the local poster first and falls back to the thumbnail once on error', () => {
        renderGrid([buildMovie({ id: 5 })]);

        const image = screen.getByRole('img', { name: 'Movie' });
        expect(image.getAttribute('src')).toMatch(
            /\/video\/movies\/5\/poster\?width=400&height=600$/
        );

        fireEvent.error(image);
        expect(image.getAttribute('src')).toMatch(
            /\/files\/video-thumbnail\/5\?width=320&height=180$/
        );

        fireEvent.error(image);
        expect(image.getAttribute('src')).toMatch(/\/files\/video-thumbnail\/5/);
    });
});

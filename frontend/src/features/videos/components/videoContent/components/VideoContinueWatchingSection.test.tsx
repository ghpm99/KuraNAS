import { fireEvent, render, screen } from '@testing-library/react';
import VideoContinueWatchingSection from './VideoContinueWatchingSection';
import type { VideoContinueItemDto } from '@/service/videoPlayback';

const buildContinueItem = (
    videoId: number,
    positionSeconds: number,
    durationSeconds: number
): VideoContinueItemDto => ({
    video: {
        id: videoId,
        name: `Episode ${videoId}`,
        path: `/shows/episode-${videoId}.mp4`,
        parent_path: '/shows',
        format: '.mp4',
        size: 100,
    },
    position_seconds: positionSeconds,
    duration_seconds: durationSeconds,
    updated_at: '2026-03-01T00:00:00Z',
});

describe('VideoContinueWatchingSection', () => {
    it('renders without any provider or service mock and with partial data', () => {
        const partialItem = {
            video: { id: 9, name: 'Partial' },
            position_seconds: 5,
        } as unknown as VideoContinueItemDto;

        render(<VideoContinueWatchingSection items={[partialItem]} onPlayVideo={jest.fn()} />);

        expect(screen.getByText('Partial')).toBeInTheDocument();
        expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0');
    });

    it('shows the empty state when nothing is in progress', () => {
        render(<VideoContinueWatchingSection items={[]} onPlayVideo={jest.fn()} />);

        expect(screen.getByText('VIDEO_NO_CONTINUE_WATCHING')).toBeInTheDocument();
    });

    it('renders one card per video with its progress bar', () => {
        render(
            <VideoContinueWatchingSection
                items={[buildContinueItem(1, 25, 100), buildContinueItem(2, 90, 60)]}
                onPlayVideo={jest.fn()}
            />
        );

        const progressBars = screen.getAllByRole('progressbar');
        expect(progressBars).toHaveLength(2);
        expect(progressBars[0]).toHaveAttribute('aria-valuenow', '25');
        expect(progressBars[1]).toHaveAttribute('aria-valuenow', '100');
    });

    it('plays exactly the video of the clicked card', () => {
        const onPlayVideo = jest.fn();
        render(
            <VideoContinueWatchingSection
                items={[buildContinueItem(1, 25, 100), buildContinueItem(2, 10, 100)]}
                onPlayVideo={onPlayVideo}
            />
        );

        fireEvent.click(screen.getAllByRole('button', { name: /VIDEO_PLAY/ })[1]!);

        expect(onPlayVideo).toHaveBeenCalledTimes(1);
        expect(onPlayVideo).toHaveBeenCalledWith(2, null);
    });

    it('marks a continue watching card as watched', () => {
        const onSetWatched = jest.fn();
        render(
            <VideoContinueWatchingSection
                items={[buildContinueItem(1, 25, 100), buildContinueItem(2, 90, 60)]}
                onPlayVideo={jest.fn()}
                onSetWatched={onSetWatched}
            />
        );

        fireEvent.click(screen.getAllByRole('button', { name: 'VIDEO_MARK_WATCHED' })[1]!);

        expect(onSetWatched).toHaveBeenCalledWith(2, true);
    });
});

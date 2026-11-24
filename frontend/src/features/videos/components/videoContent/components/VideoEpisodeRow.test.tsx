import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { VideoPlaylistItemDto } from '@/service/videoPlayback';
import VideoEpisodeRow from './VideoEpisodeRow';

const buildEpisode = (overrides: Partial<VideoPlaylistItemDto> = {}): VideoPlaylistItemDto => ({
    id: 1,
    order_index: 0,
    source_kind: 'auto',
    status: 'not_started',
    progress_pct: 0,
    season: 2,
    episode: 3,
    video: {
        id: 90,
        name: 'Show.S02E03.mkv',
        path: '/show/Show.S02E03.mkv',
        parent_path: '/show',
        format: 'mkv',
        size: 1,
        duration_seconds: 1500,
    },
    ...overrides,
});

describe('VideoEpisodeRow', () => {
    it('renders a bare episode without a backend, season, episode or duration', () => {
        renderWithoutBackend(
            <VideoEpisodeRow
                episode={
                    {
                        id: 1,
                        status: 'not_started',
                        progress_pct: 0,
                        video: { id: 5, name: 'Loose.mkv' },
                    } as VideoPlaylistItemDto
                }
                onOpenVideo={jest.fn()}
            />
        );

        expect(screen.getByText('Loose')).toBeInTheDocument();
    });

    it('shows number, title, duration and opens the video', () => {
        const onOpenVideo = jest.fn();
        renderWithoutBackend(
            <VideoEpisodeRow episode={buildEpisode()} onOpenVideo={onOpenVideo} />
        );

        expect(screen.getByText(/VIDEO_SERIES_EPISODE_NUMBER/)).toBeInTheDocument();
        expect(screen.getByText('25:00')).toBeInTheDocument();
        expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: /Show/ }));
        expect(onOpenVideo).toHaveBeenCalledWith(90);
    });

    it('shows the progress bar while in progress and the watched badge when completed', () => {
        const { unmount } = renderWithoutBackend(
            <VideoEpisodeRow
                episode={buildEpisode({ status: 'in_progress', progress_pct: 40 })}
                onOpenVideo={jest.fn()}
            />
        );
        expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '40');
        expect(screen.queryByText('VIDEO_STATUS_COMPLETED')).not.toBeInTheDocument();
        unmount();

        renderWithoutBackend(
            <VideoEpisodeRow
                episode={buildEpisode({ status: 'completed', progress_pct: 100 })}
                onOpenVideo={jest.fn()}
            />
        );
        expect(screen.getByText(/VIDEO_STATUS_COMPLETED/)).toBeInTheDocument();
    });
});

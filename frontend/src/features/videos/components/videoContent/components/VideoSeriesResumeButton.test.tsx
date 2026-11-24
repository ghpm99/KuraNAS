import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoSeriesResumeButton from './VideoSeriesResumeButton';

describe('VideoSeriesResumeButton', () => {
    it('renders disabled without a resume target and without a backend', () => {
        renderWithoutBackend(<VideoSeriesResumeButton resume={null} onOpenVideo={jest.fn()} />);

        expect(screen.getByRole('button')).toBeDisabled();
    });

    it.each([
        ['continue', 'VIDEO_SERIES_CONTINUE'],
        ['start', 'VIDEO_SERIES_WATCH'],
        ['rewatch', 'VIDEO_SERIES_REWATCH'],
    ] as const)('labels the %s mode and opens the resume video', (mode, expectedKey) => {
        const onOpenVideo = jest.fn();
        renderWithoutBackend(
            <VideoSeriesResumeButton
                resume={{
                    video_id: 77,
                    season: 1,
                    episode: 5,
                    position_seconds: 12,
                    mode,
                }}
                onOpenVideo={onOpenVideo}
            />
        );

        const button = screen.getByRole('button', { name: new RegExp(expectedKey) });
        fireEvent.click(button);

        expect(onOpenVideo).toHaveBeenCalledWith(77);
    });

    it('falls back to the generic resume label when the episode number is unknown', () => {
        renderWithoutBackend(
            <VideoSeriesResumeButton
                resume={{
                    video_id: 8,
                    season: null,
                    episode: null,
                    position_seconds: 0,
                    mode: 'continue',
                }}
                onOpenVideo={jest.fn()}
            />
        );

        expect(screen.getByRole('button', { name: /VIDEO_DETAIL_RESUME_ACTION/ })).toBeEnabled();
    });
});

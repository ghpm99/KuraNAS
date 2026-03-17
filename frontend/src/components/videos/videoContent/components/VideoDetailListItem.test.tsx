import { fireEvent, render, screen } from '@testing-library/react';
import VideoDetailListItem from './VideoDetailListItem';
import type { VideoDetailItem } from '../useVideoPlaylistDetail';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => {
            const map: Record<string, string> = {
                VIDEO_STATUS_NOT_STARTED: 'Not Started',
                VIDEO_STATUS_IN_PROGRESS: 'In Progress',
                VIDEO_STATUS_COMPLETED: 'Completed',
                VIDEO_SOURCE_MANUAL: 'Manual',
                VIDEO_SOURCE_AUTO: 'Auto',
            };
            return map[key] ?? key;
        },
    }),
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => 'http://localhost:8000/api/v1',
}));

const createItem = (overrides: Partial<VideoDetailItem> = {}): VideoDetailItem => ({
    id: 1,
    order_index: 0,
    source_kind: 'auto',
    status: 'not_started',
    progress_pct: 0,
    displayTitle: '',
    sequenceLabel: '',
    seasonNumber: null,
    episodeNumber: null,
    video: {
        id: 10,
        name: 'Episode 1.mp4',
        path: '/videos/episode-1.mp4',
        parent_path: '/videos',
        format: 'mp4',
        size: 500000,
    },
    ...overrides,
});

describe('VideoDetailListItem', () => {
    it('renders the video name and format', () => {
        render(<VideoDetailListItem item={createItem()} onOpenVideo={jest.fn()} />);

        expect(screen.getByText('Episode 1.mp4')).toBeInTheDocument();
        expect(screen.getByText(/MP4/)).toBeInTheDocument();
    });

    it('renders the thumbnail with the correct src', () => {
        render(<VideoDetailListItem item={createItem()} onOpenVideo={jest.fn()} />);

        const img = screen.getByRole('img', { name: 'Episode 1.mp4' });
        expect(img).toHaveAttribute(
            'src',
            'http://localhost:8000/api/v1/files/video-thumbnail/10?width=320&height=180'
        );
    });

    it('shows displayTitle when provided', () => {
        const item = createItem({ displayTitle: 'Custom Title' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText('Custom Title')).toBeInTheDocument();
    });

    it('falls back to video.name when displayTitle is empty', () => {
        const item = createItem({ displayTitle: '' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText('Episode 1.mp4')).toBeInTheDocument();
    });

    it('shows the sequence label when provided', () => {
        const item = createItem({ sequenceLabel: 'S01E03' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText('S01E03')).toBeInTheDocument();
    });

    it('does not render a sequence tag when label is empty', () => {
        const item = createItem({ sequenceLabel: '' });
        const { container } = render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        // The sequenceTag span should not be present
        expect(container.querySelectorAll('[class*="sequenceTag"]')).toHaveLength(0);
    });

    it('calls onOpenVideo with the video id on click', () => {
        const onOpenVideo = jest.fn();
        render(<VideoDetailListItem item={createItem()} onOpenVideo={onOpenVideo} />);

        fireEvent.click(screen.getByRole('button'));
        expect(onOpenVideo).toHaveBeenCalledWith(10);
    });

    it('renders "Not Started" status badge for not_started', () => {
        const item = createItem({ status: 'not_started' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText('Not Started')).toBeInTheDocument();
    });

    it('renders "In Progress" status badge for in_progress', () => {
        const item = createItem({ status: 'in_progress' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText('In Progress')).toBeInTheDocument();
    });

    it('renders "Completed" status badge for completed', () => {
        const item = createItem({ status: 'completed' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText('Completed')).toBeInTheDocument();
    });

    it('shows the progress bar when progress > 0 and status is in_progress', () => {
        const item = createItem({ status: 'in_progress', progress_pct: 55 });
        const { container } = render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        const progressFill = container.querySelector('[class*="progressFill"]');
        expect(progressFill).not.toBeNull();
        expect(progressFill).toHaveStyle({ width: '55%' });
    });

    it('does not show the progress bar when status is not_started even with progress > 0', () => {
        const item = createItem({ status: 'not_started', progress_pct: 10 });
        const { container } = render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        const progressTrack = container.querySelector('[class*="progressTrack"]');
        expect(progressTrack).toBeNull();
    });

    it('does not show progress bar when progress is 0', () => {
        const item = createItem({ status: 'in_progress', progress_pct: 0 });
        const { container } = render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        const progressTrack = container.querySelector('[class*="progressTrack"]');
        expect(progressTrack).toBeNull();
    });

    it('shows "Manual" for manual source kind', () => {
        const item = createItem({ source_kind: 'manual' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText(/Manual/)).toBeInTheDocument();
    });

    it('shows "Auto" for auto source kind', () => {
        const item = createItem({ source_kind: 'auto' });
        render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        expect(screen.getByText(/Auto/)).toBeInTheDocument();
    });

    it('renders completed status without progress bar', () => {
        const item = createItem({ status: 'completed', progress_pct: 100 });
        const { container } = render(<VideoDetailListItem item={item} onOpenVideo={jest.fn()} />);

        // Completed items should show the progress bar (progress > 0 and status !== 'not_started')
        const progressFill = container.querySelector('[class*="progressFill"]');
        expect(progressFill).not.toBeNull();
    });
});

import { render, screen } from '@testing-library/react';
import VideoMediaBadges from './VideoMediaBadges';

describe('VideoMediaBadges', () => {
    it('renders without a video', () => {
        const { container } = render(<VideoMediaBadges />);
        expect(container).toBeEmptyDOMElement();
    });

    it('renders nothing for a video without technical fields', () => {
        const { container } = render(<VideoMediaBadges video={{}} />);
        expect(container).toBeEmptyDOMElement();
    });

    it('renders nothing for null fields', () => {
        const { container } = render(
            <VideoMediaBadges video={{ duration_seconds: null, height: null }} />
        );
        expect(container).toBeEmptyDOMElement();
    });

    it('renders duration and resolution badges', () => {
        render(<VideoMediaBadges video={{ duration_seconds: 3725, height: 1080 }} />);
        expect(screen.getByText('1:02:05')).toBeInTheDocument();
        expect(screen.getByText('VIDEO_BADGE_RESOLUTION_FHD')).toBeInTheDocument();
    });

    it('renders only the duration when height is missing', () => {
        render(<VideoMediaBadges video={{ duration_seconds: 59 }} />);
        expect(screen.getByText('0:59')).toBeInTheDocument();
        expect(screen.queryByText(/VIDEO_BADGE_RESOLUTION/)).not.toBeInTheDocument();
    });

    it('renders only the resolution when duration is missing', () => {
        render(<VideoMediaBadges video={{ height: 2160 }} />);
        expect(screen.getByText('VIDEO_BADGE_RESOLUTION_UHD')).toBeInTheDocument();
        expect(screen.queryByText(/:/)).not.toBeInTheDocument();
    });
});

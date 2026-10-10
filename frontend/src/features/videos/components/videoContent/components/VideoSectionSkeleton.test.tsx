import { render, screen } from '@testing-library/react';
import VideoSectionSkeleton from './VideoSectionSkeleton';

describe('VideoSectionSkeleton', () => {
    it('renders with default props', () => {
        render(<VideoSectionSkeleton />);
        expect(screen.getByTestId('video-section-skeleton')).toBeInTheDocument();
    });

    it('renders the requested number of cards in the catalog rail layout', () => {
        const { container } = render(<VideoSectionSkeleton cardCount={2} layout="catalogRail" />);
        expect(container.querySelectorAll('.MuiSkeleton-rounded')).toHaveLength(2);
    });
});

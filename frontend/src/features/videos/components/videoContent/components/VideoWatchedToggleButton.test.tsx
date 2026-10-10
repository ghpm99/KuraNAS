import { fireEvent, render, screen } from '@testing-library/react';
import VideoWatchedToggleButton from './VideoWatchedToggleButton';

describe('VideoWatchedToggleButton', () => {
    it('renders without any provider or service mock', () => {
        render(<VideoWatchedToggleButton isWatched={false} onToggle={jest.fn()} />);

        expect(screen.getByRole('button', { name: 'VIDEO_MARK_WATCHED' })).toBeInTheDocument();
    });

    it('offers to mark as unwatched when already watched and toggles on click', () => {
        const onToggle = jest.fn();
        render(<VideoWatchedToggleButton isWatched onToggle={onToggle} />);

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_MARK_UNWATCHED' }));

        expect(onToggle).toHaveBeenCalledTimes(1);
    });
});

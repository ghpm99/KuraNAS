import { createRef } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import VideoPlayer from './videoPlayer';

const renderPlayer = (overrides: Record<string, unknown> = {}) =>
    render(
        <VideoPlayer
            currentVideo={{ id: 42, name: 'a.mkv' } as any}
            videoRef={createRef<HTMLVideoElement>()}
            setCurrentTime={jest.fn()}
            setDuration={jest.fn()}
            onBack={jest.fn()}
            onVideoEnded={jest.fn()}
            originBadgeLabel="Videos"
            contextDescription="From Videos"
            metadataLine=""
            {...overrides}
        />
    );

describe('videoPlayer details button', () => {
    it('renders no details button without a handler', () => {
        renderPlayer();

        expect(screen.queryByRole('button', { name: /FILES_DETAILS|Details/ })).toBeNull();
    });

    it('opens the details through the handler', () => {
        const onOpenDetails = jest.fn();
        renderPlayer({ onOpenDetails });

        fireEvent.click(screen.getByRole('button', { name: 'FILES_DETAILS' }));

        expect(onOpenDetails).toHaveBeenCalledTimes(1);
    });

    it('hides the details button when no video is playing', () => {
        renderPlayer({ onOpenDetails: jest.fn(), currentVideo: null });

        expect(screen.queryByRole('button', { name: 'FILES_DETAILS' })).toBeNull();
    });
});

describe('videoPlayer star button', () => {
    it('renders no star button without a handler', () => {
        renderPlayer();

        expect(screen.queryByRole('button', { name: /FILES_FAVORITE_ITEM/ })).toBeNull();
    });

    it('shows the unpressed state and toggles through the handler', () => {
        const onToggleStar = jest.fn();
        renderPlayer({ onToggleStar });

        const starButton = screen.getByRole('button', { name: 'FILES_FAVORITE_ITEM' });
        expect(starButton).toHaveAttribute('aria-pressed', 'false');
        fireEvent.click(starButton);

        expect(onToggleStar).toHaveBeenCalledTimes(1);
    });

    it('shows the pressed state for a starred video', () => {
        renderPlayer({
            onToggleStar: jest.fn(),
            currentVideo: { id: 42, name: 'a.mkv', starred: true },
        });

        expect(screen.getByRole('button', { name: 'FILES_UNFAVORITE_ITEM' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
    });

    it('hides the star button when no video is playing', () => {
        renderPlayer({ onToggleStar: jest.fn(), currentVideo: null });

        expect(screen.queryByRole('button', { name: /FAVORITE/ })).toBeNull();
    });
});

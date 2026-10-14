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

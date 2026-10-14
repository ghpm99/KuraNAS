import { fireEvent, render, screen } from '@testing-library/react';
import VideoControls from './videoControls';

const buildProps = () => ({
    isPlaying: false,
    currentTime: 0,
    duration: 100,
    volume: 0.5,
    playbackRate: 1,
    isFullscreen: false,
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    setPlaybackRate: jest.fn(),
    toggleFullscreen: jest.fn(),
    togglePlayPause: jest.fn(),
    nextVideo: jest.fn(),
    previousVideo: jest.fn(),
});

describe('features/videos/videoControls/videoControls', () => {
    it('renders without the shortcuts help button when no handler is given', () => {
        render(<VideoControls {...buildProps()} />);

        expect(screen.queryByRole('button', { name: 'VIDEO_SHORTCUTS_HELP_BUTTON' })).toBeNull();
    });

    it('opens the shortcuts help from the discreet button', () => {
        const onOpenShortcutsHelp = jest.fn();
        render(<VideoControls {...buildProps()} onOpenShortcutsHelp={onOpenShortcutsHelp} />);

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_SHORTCUTS_HELP_BUTTON' }));

        expect(onOpenShortcutsHelp).toHaveBeenCalledTimes(1);
    });

    it('hides the picture in picture button when unsupported', () => {
        render(<VideoControls {...buildProps()} />);

        expect(screen.queryByRole('button', { name: 'VIDEO_PICTURE_IN_PICTURE' })).toBeNull();
    });

    it('toggles picture in picture from the button with the matching label', () => {
        const togglePictureInPicture = jest.fn();
        const { rerender } = render(
            <VideoControls
                {...buildProps()}
                isPictureInPictureSupported
                togglePictureInPicture={togglePictureInPicture}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_PICTURE_IN_PICTURE' }));
        expect(togglePictureInPicture).toHaveBeenCalledTimes(1);

        rerender(
            <VideoControls
                {...buildProps()}
                isPictureInPictureSupported
                isPictureInPicture
                togglePictureInPicture={togglePictureInPicture}
            />
        );
        expect(
            screen.getByRole('button', { name: 'VIDEO_EXIT_PICTURE_IN_PICTURE' })
        ).toHaveAttribute('aria-pressed', 'true');
    });
});

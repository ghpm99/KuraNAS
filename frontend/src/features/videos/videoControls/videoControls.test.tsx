import { fireEvent, render, screen } from '@testing-library/react';
import VideoControls from './videoControls';

const buildProps = () => ({
    isPlaying: false,
    currentTime: 0,
    duration: 100,
    volume: 0.5,
    isMuted: false,
    playbackRate: 1,
    isFullscreen: false,
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    toggleMute: jest.fn(),
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

    it('offers to mute when sound is on and calls toggleMute', () => {
        const props = buildProps();
        render(<VideoControls {...props} />);

        const muteButton = screen.getByRole('button', { name: 'VIDEO_MUTE' });
        expect(muteButton).toHaveAttribute('aria-pressed', 'false');
        fireEvent.click(muteButton);

        expect(props.toggleMute).toHaveBeenCalledTimes(1);
        expect(props.setVolume).not.toHaveBeenCalled();
    });

    it('offers to unmute when muted while the slider keeps the real volume', () => {
        const { container } = render(<VideoControls {...buildProps()} isMuted />);

        expect(screen.getByRole('button', { name: 'VIDEO_UNMUTE' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
        expect(container.querySelector('.volume-slider input')).toHaveValue('0.5');
    });

    it('offers to unmute when the volume is zero', () => {
        render(<VideoControls {...buildProps()} volume={0} />);

        expect(screen.getByRole('button', { name: 'VIDEO_UNMUTE' })).toBeInTheDocument();
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

    describe('time slider scrubbing', () => {
        const getTimeSlider = (container: HTMLElement) =>
            container.querySelector('.progress-bar input') as HTMLInputElement;

        const mockSliderGeometry = (container: HTMLElement) => {
            const sliderRoot = container.querySelector('.progress-bar') as HTMLElement;
            jest.spyOn(sliderRoot, 'getBoundingClientRect').mockReturnValue({
                left: 0,
                right: 200,
                width: 200,
                top: 0,
                bottom: 10,
                height: 10,
                x: 0,
                y: 0,
                toJSON: () => ({}),
            });
            return sliderRoot;
        };

        it('does not seek while dragging and seeks once with the final value on release', () => {
            const props = buildProps();
            const { container } = render(<VideoControls {...props} />);
            const sliderRoot = mockSliderGeometry(container);

            fireEvent.mouseDown(sliderRoot, { clientX: 40, clientY: 5 });
            fireEvent.mouseMove(document, { clientX: 70, clientY: 5, buttons: 1 });
            fireEvent.mouseMove(document, { clientX: 120, clientY: 5, buttons: 1 });

            expect(props.seekTo).not.toHaveBeenCalled();
            expect(getTimeSlider(container)).toHaveValue('60');
            expect(screen.getAllByText('1:00').length).toBeGreaterThan(0);

            fireEvent.mouseUp(document, { clientX: 120, clientY: 5 });

            expect(props.seekTo).toHaveBeenCalledTimes(1);
            expect(props.seekTo).toHaveBeenCalledWith(60);
        });

        it('keeps the dragged position when currentTime updates during the drag', () => {
            const props = buildProps();
            const { container, rerender } = render(<VideoControls {...props} />);
            const sliderRoot = mockSliderGeometry(container);

            fireEvent.mouseDown(sliderRoot, { clientX: 120, clientY: 5 });
            rerender(<VideoControls {...props} currentTime={5} />);

            expect(getTimeSlider(container)).toHaveValue('60');

            fireEvent.mouseUp(document, { clientX: 120, clientY: 5 });
        });

        it('applies the seek once when a keyboard step is confirmed', () => {
            const props = buildProps();
            const { container } = render(<VideoControls {...props} currentTime={10} />);

            fireEvent.keyDown(getTimeSlider(container), { key: 'ArrowRight' });

            expect(props.seekTo).toHaveBeenCalledTimes(1);
            expect(props.seekTo).toHaveBeenCalledWith(11);
        });

        it('follows playback again after the seek is applied', () => {
            const props = buildProps();
            const { container, rerender } = render(<VideoControls {...props} currentTime={10} />);

            fireEvent.keyDown(getTimeSlider(container), { key: 'End' });
            rerender(<VideoControls {...props} currentTime={40} />);

            expect(getTimeSlider(container)).toHaveValue('40');
        });
    });
});

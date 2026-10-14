import { fireEvent, render, screen } from '@testing-library/react';
import UpNextAutoplayCard from './UpNextAutoplayCard';

const buildProps = (overrides: Partial<React.ComponentProps<typeof UpNextAutoplayCard>> = {}) => ({
    nextVideoId: 31,
    nextVideoTitle: 'Episode 2',
    isCounting: true,
    secondsRemaining: 5,
    isAutoplayEnabled: true,
    onPlayNow: jest.fn(),
    onCancel: jest.fn(),
    onAutoplayEnabledChange: jest.fn(),
    ...overrides,
});

describe('features/videos/upNext/UpNextAutoplayCard', () => {
    it('renders without any provider or backend', () => {
        render(<UpNextAutoplayCard {...buildProps()} />);

        expect(screen.getByText('Episode 2')).toBeInTheDocument();
        expect(screen.getByRole('group')).toBeInTheDocument();
    });

    it('announces the countdown in a polite live region and loads the thumbnail', () => {
        render(<UpNextAutoplayCard {...buildProps()} />);

        const countdown = screen.getByRole('status');
        expect(countdown).toHaveAttribute('aria-live', 'polite');
        expect(countdown).toHaveTextContent('VIDEO_UP_NEXT_COUNTDOWN');
        expect(document.querySelector('img')?.getAttribute('src')).toContain(
            '/files/video-thumbnail/31'
        );
    });

    it('shows no countdown text when it is not counting', () => {
        render(<UpNextAutoplayCard {...buildProps({ isCounting: false })} />);

        expect(screen.getByRole('status')).toBeEmptyDOMElement();
    });

    it('focuses the primary action on appearing', () => {
        render(<UpNextAutoplayCard {...buildProps()} />);

        expect(screen.getByText('VIDEO_UP_NEXT_PLAY_NOW')).toHaveFocus();
    });

    it('does not steal focus from a field the user is typing in', () => {
        const field = document.createElement('input');
        document.body.appendChild(field);
        field.focus();

        render(<UpNextAutoplayCard {...buildProps()} />);

        expect(field).toHaveFocus();
        field.remove();
    });

    it('plays now and cancels through the buttons and Escape', () => {
        const props = buildProps();
        render(<UpNextAutoplayCard {...props} />);

        fireEvent.click(screen.getByText('VIDEO_UP_NEXT_PLAY_NOW'));
        expect(props.onPlayNow).toHaveBeenCalledTimes(1);

        fireEvent.click(screen.getByText('VIDEO_UP_NEXT_CANCEL'));
        expect(props.onCancel).toHaveBeenCalledTimes(1);

        fireEvent.keyDown(document, { key: 'Escape' });
        expect(props.onCancel).toHaveBeenCalledTimes(2);

        fireEvent.keyDown(document, { key: 'a' });
        expect(props.onCancel).toHaveBeenCalledTimes(2);
    });

    it('stops listening for Escape after unmount', () => {
        const props = buildProps();
        const { unmount } = render(<UpNextAutoplayCard {...props} />);

        unmount();
        fireEvent.keyDown(document, { key: 'Escape' });

        expect(props.onCancel).not.toHaveBeenCalled();
    });

    it('toggles the autoplay preference', () => {
        const props = buildProps();
        render(<UpNextAutoplayCard {...props} />);

        const toggle = screen.getByRole('checkbox');
        expect(toggle).toBeChecked();
        fireEvent.click(toggle);

        expect(props.onAutoplayEnabledChange).toHaveBeenCalledWith(false);
    });
});

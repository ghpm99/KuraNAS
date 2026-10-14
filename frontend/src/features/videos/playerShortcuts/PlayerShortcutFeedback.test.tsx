import { render, screen } from '@testing-library/react';
import PlayerShortcutFeedback from './PlayerShortcutFeedback';

describe('features/videos/playerShortcuts/PlayerShortcutFeedback', () => {
    it('renders an empty polite live region without any props', () => {
        render(<PlayerShortcutFeedback />);

        const liveRegion = screen.getByRole('status');
        expect(liveRegion).toHaveAttribute('aria-live', 'polite');
        expect(liveRegion).toBeEmptyDOMElement();
    });

    it('renders the feedback message', () => {
        render(
            <PlayerShortcutFeedback
                feedback={{ id: 1, messageKey: 'VIDEO_SHORTCUT_FEEDBACK_MUTED' }}
            />
        );

        expect(screen.getByRole('status')).toHaveTextContent('VIDEO_SHORTCUT_FEEDBACK_MUTED');
    });
});

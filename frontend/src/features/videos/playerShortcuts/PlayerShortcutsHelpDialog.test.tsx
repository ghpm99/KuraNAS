import { fireEvent, render, screen } from '@testing-library/react';
import PlayerShortcutsHelpDialog from './PlayerShortcutsHelpDialog';
import { playerShortcutHelpEntries } from './playerShortcutHelpEntries';

describe('features/videos/playerShortcuts/PlayerShortcutsHelpDialog', () => {
    it('renders closed without any props', () => {
        render(<PlayerShortcutsHelpDialog />);

        expect(screen.queryByRole('dialog')).toBeNull();
    });

    it('lists every shortcut when open and closes on Escape', () => {
        const onClose = jest.fn();
        render(<PlayerShortcutsHelpDialog isOpen onClose={onClose} />);

        expect(screen.getByRole('dialog')).toBeInTheDocument();
        expect(screen.getByText('VIDEO_SHORTCUT_HELP_TITLE')).toBeInTheDocument();
        playerShortcutHelpEntries.forEach((entry) => {
            expect(screen.getByText(entry.descriptionKey)).toBeInTheDocument();
        });
        expect(screen.getByText('VIDEO_SHORTCUT_KEY_SPACE')).toBeInTheDocument();

        fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
        expect(onClose).toHaveBeenCalled();
    });
});

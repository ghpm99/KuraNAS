import { fireEvent, render, screen } from '@testing-library/react';
import KeyboardShortcutsDialog from './keyboardShortcutsDialog';
import { fileShortcutDefinitions } from './fileShortcutDefinitions';

describe('KeyboardShortcutsDialog', () => {
    it('renders nothing visible when closed, without any provider', () => {
        render(<KeyboardShortcutsDialog isOpen={false} onClose={jest.fn()} />);

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('lists every shortcut and closes on request', () => {
        const onClose = jest.fn();
        render(<KeyboardShortcutsDialog isOpen onClose={onClose} />);

        expect(screen.getByRole('dialog')).toBeInTheDocument();
        fileShortcutDefinitions.forEach((shortcut) => {
            expect(screen.getByText(shortcut.descriptionKey)).toBeInTheDocument();
        });
        fireEvent.click(screen.getByRole('button', { name: 'CLOSE' }));
        expect(onClose).toHaveBeenCalledTimes(1);
    });
});

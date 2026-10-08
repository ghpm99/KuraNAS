import { fireEvent, render, screen } from '@testing-library/react';
import ActionBarMoreMenu from './actionBarMoreMenu';

describe('ActionBarMoreMenu', () => {
    it('renders nothing without entries and without any provider or mock', () => {
        const { container } = render(<ActionBarMoreMenu entries={[]} />);

        expect(container).toBeEmptyDOMElement();
    });

    it('opens the menu, runs the chosen entry and closes', async () => {
        const onSelect = jest.fn();
        render(
            <ActionBarMoreMenu
                entries={[
                    { key: 'rescan', label: 'Rescan', icon: <i aria-hidden="true" />, onSelect },
                    { key: 'delete', label: 'Delete', icon: <i aria-hidden="true" />, onSelect: jest.fn(), isDestructive: true },
                ]}
            />
        );

        expect(screen.queryByRole('menuitem')).toBeNull();
        fireEvent.click(screen.getByRole('button', { name: 'FILES_MORE_ACTIONS' }));
        expect(screen.getAllByRole('menuitem')).toHaveLength(2);

        fireEvent.click(screen.getByRole('menuitem', { name: 'Rescan' }));

        expect(onSelect).toHaveBeenCalledTimes(1);
    });
});

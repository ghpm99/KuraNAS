import { fireEvent, render, screen } from '@testing-library/react';
import VacuumSwitchPreference from './VacuumSwitchPreference';

describe('VacuumSwitchPreference', () => {
    it('renders without any provider or service', () => {
        render(
            <VacuumSwitchPreference label="Carpet" isChecked={false} isDisabled={false} onToggle={jest.fn()} />
        );

        expect(screen.getByRole('switch', { name: 'Carpet' })).not.toBeChecked();
    });

    it('reports the new value when toggled', () => {
        const onToggle = jest.fn();
        render(
            <VacuumSwitchPreference label="Carpet" isChecked={false} isDisabled={false} onToggle={onToggle} />
        );

        fireEvent.click(screen.getByRole('switch', { name: 'Carpet' }));

        expect(onToggle).toHaveBeenCalledWith(true);
    });

    it('reports false when turned off', () => {
        const onToggle = jest.fn();
        render(<VacuumSwitchPreference label="Carpet" isChecked isDisabled={false} onToggle={onToggle} />);

        fireEvent.click(screen.getByRole('switch', { name: 'Carpet' }));

        expect(onToggle).toHaveBeenCalledWith(false);
    });

    it('is disabled while a change is being saved', () => {
        render(<VacuumSwitchPreference label="Carpet" isChecked isDisabled onToggle={jest.fn()} />);

        expect(screen.getByRole('switch', { name: 'Carpet' })).toBeDisabled();
    });
});

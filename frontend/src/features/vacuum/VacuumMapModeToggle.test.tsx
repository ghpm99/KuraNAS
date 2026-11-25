import { fireEvent, render, screen } from '@testing-library/react';
import VacuumMapModeToggle from './VacuumMapModeToggle';

describe('VacuumMapModeToggle', () => {
    it('renders without any provider or service', () => {
        expect(() => render(<VacuumMapModeToggle touchMode="select-rooms" onChange={jest.fn()} />)).not.toThrow();
    });

    it('marks the current mode as pressed', () => {
        render(<VacuumMapModeToggle touchMode="clean-spot" onChange={jest.fn()} />);

        expect(screen.getByRole('button', { name: 'VACUUM_MAP_MODE_CLEAN_SPOT' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
        expect(screen.getByRole('button', { name: 'VACUUM_MAP_MODE_SELECT_ROOMS' })).toHaveAttribute(
            'aria-pressed',
            'false'
        );
    });

    it('reports the other mode and ignores a click on the current one', () => {
        const onChange = jest.fn();
        render(<VacuumMapModeToggle touchMode="select-rooms" onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_MODE_SELECT_ROOMS' }));
        expect(onChange).not.toHaveBeenCalled();

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_MODE_CLEAN_SPOT' }));
        expect(onChange).toHaveBeenCalledWith('clean-spot');
    });
});

import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FilterChipPopover from './FilterChipPopover';

describe('FilterChipPopover', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<FilterChipPopover label="" />);
    });

    it('opens its content from the chip and closes on escape', () => {
        render(
            <FilterChipPopover label="Tipo">
                <span>conteudo</span>
            </FilterChipPopover>
        );

        expect(screen.queryByText('conteudo')).toBeNull();
        fireEvent.click(screen.getByRole('button', { name: 'Tipo' }));
        expect(screen.getByText('conteudo')).toBeInTheDocument();

        fireEvent.keyDown(screen.getByText('conteudo'), { key: 'Escape' });
        expect(screen.getByRole('button', { name: 'Tipo', hidden: true })).toHaveAttribute(
            'aria-expanded',
            'false'
        );
    });
});

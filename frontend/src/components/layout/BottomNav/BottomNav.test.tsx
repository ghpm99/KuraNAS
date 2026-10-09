import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { BottomNav } from './BottomNav';

const renderAt = (pathname: string, onOpenMenu = jest.fn()) => {
    render(
        <MemoryRouter initialEntries={[pathname]}>
            <BottomNav onOpenMenu={onOpenMenu} />
        </MemoryRouter>
    );
    return { onOpenMenu };
};

describe('layout/BottomNav', () => {
    it('renders labelled navigation without any provider or mock', () => {
        renderAt('/home');

        expect(screen.getByRole('navigation', { name: 'NAV_BOTTOM_LABEL' })).toBeInTheDocument();
        expect(screen.getAllByRole('link')).toHaveLength(4);
    });

    it('marks only the current section link with aria-current page', () => {
        renderAt('/files/some/folder');

        expect(screen.getByRole('link', { name: 'FILES' })).toHaveAttribute('aria-current', 'page');
        expect(screen.getByRole('link', { name: 'HOME' })).not.toHaveAttribute('aria-current');
        expect(screen.getByRole('link', { name: 'NAV_MUSIC' })).not.toHaveAttribute('aria-current');
    });

    it('keeps the menu entry as a button that opens the drawer', () => {
        const { onOpenMenu } = renderAt('/home');

        fireEvent.click(screen.getByRole('button', { name: 'MENU' }));

        expect(onOpenMenu).toHaveBeenCalledTimes(1);
    });
});

import { fireEvent, render, screen } from '@testing-library/react';
import Sidebar from './Sidebar';
import { MemoryRouter } from 'react-router-dom';
import { navigationItems } from '@/components/layout/navigationItems';

describe('layout/Sidebar', () => {
    it('renders without any provider or service mock', () => {
        render(
            <MemoryRouter initialEntries={['/files']}>
                <Sidebar />
            </MemoryRouter>
        );

        expect(screen.getByRole('navigation')).toBeInTheDocument();
        expect(screen.getAllByRole('link')).toHaveLength(navigationItems.length);
    });

    it('renders brand and navigation entries', () => {
        render(
            <MemoryRouter initialEntries={['/images']}>
                <Sidebar />
            </MemoryRouter>
        );

        expect(screen.getByText('APP_NAME')).toBeInTheDocument();
        expect(screen.getByText('HOME')).toBeInTheDocument();
        expect(screen.getByText('FILES')).toBeInTheDocument();
        expect(screen.getByText('NAV_IMAGES')).toBeInTheDocument();
    });

    it('shows the collapse control that reports the toggle', () => {
        const onToggleCollapsed = jest.fn();
        render(
            <MemoryRouter>
                <Sidebar isCollapsed={false} onToggleCollapsed={onToggleCollapsed} />
            </MemoryRouter>
        );

        const collapseButton = screen.getByRole('button', { name: 'SIDEBAR_COLLAPSE' });
        expect(collapseButton).toHaveAttribute('aria-expanded', 'true');
        fireEvent.click(collapseButton);

        expect(onToggleCollapsed).toHaveBeenCalledTimes(1);
        expect(screen.getByText('HOME')).toBeInTheDocument();
    });

    it('collapses to icons only, with every label exposed as aria-label', () => {
        render(
            <MemoryRouter>
                <Sidebar isCollapsed onToggleCollapsed={jest.fn()} />
            </MemoryRouter>
        );

        expect(screen.getByRole('button', { name: 'SIDEBAR_EXPAND' })).toHaveAttribute(
            'aria-expanded',
            'false'
        );
        expect(screen.queryByText('HOME')).not.toBeInTheDocument();
        expect(screen.getAllByRole('link')).toHaveLength(navigationItems.length);
        navigationItems.forEach((item) => {
            expect(screen.getByRole('link', { name: item.labelKey })).toBeInTheDocument();
        });
        expect(screen.getByRole('navigation')).toHaveAttribute('data-collapsed', 'true');
    });

    it('shows the label as a tooltip when hovering a collapsed item', async () => {
        render(
            <MemoryRouter>
                <Sidebar isCollapsed onToggleCollapsed={jest.fn()} />
            </MemoryRouter>
        );

        fireEvent.mouseOver(screen.getByRole('link', { name: 'HOME' }));

        expect(await screen.findByRole('tooltip')).toHaveTextContent('HOME');
    });

    it('ignores the collapsed state in the mobile drawer', () => {
        render(
            <MemoryRouter>
                <Sidebar mobile isCollapsed onToggleCollapsed={jest.fn()} />
            </MemoryRouter>
        );

        expect(screen.queryByRole('button', { name: /SIDEBAR_/ })).not.toBeInTheDocument();
        expect(screen.getByText('HOME')).toBeInTheDocument();
    });
});

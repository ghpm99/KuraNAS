import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import DomainNavTabs from './DomainNavTabs';

describe('layout/DomainNavTabs', () => {
    it('renders with an empty item list, without any provider or mock', () => {
        render(
            <MemoryRouter>
                <DomainNavTabs ariaLabel="domain" items={[]} />
            </MemoryRouter>
        );
        expect(screen.getByRole('navigation', { name: 'domain' })).toBeInTheDocument();
    });

    it('renders links and marks the active item', () => {
        render(
            <MemoryRouter>
                <DomainNavTabs
                    ariaLabel="domain"
                    items={[
                        { key: 'home', href: '/music', label: 'Inicio', isActive: true },
                        { key: 'albums', href: '/music/albums', label: 'Albuns', isActive: false },
                    ]}
                />
            </MemoryRouter>
        );

        expect(screen.getByRole('link', { name: 'Inicio' })).toHaveAttribute(
            'aria-current',
            'page'
        );
        expect(screen.getByRole('link', { name: 'Albuns' })).not.toHaveAttribute('aria-current');
    });
});

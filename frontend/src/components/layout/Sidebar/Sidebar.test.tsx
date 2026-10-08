import { render, screen } from '@testing-library/react';
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
});

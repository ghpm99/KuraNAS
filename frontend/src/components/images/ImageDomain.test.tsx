import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ImageDomainHeader from './ImageDomainHeader';
import ImageDomainNav from './ImageDomainNav';
import { getImageSectionFromPath } from './navigation';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

describe('components/images domain shell', () => {
    it('renders contextual header and active nav tab from route', () => {
        render(
            <MemoryRouter initialEntries={['/images/albums']}>
                <ImageDomainHeader />
                <ImageDomainNav />
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'IMAGES_SECTION_ALBUMS' })).toBeInTheDocument();
        expect(screen.getByText('IMAGES_SECTION_ALBUMS_DESCRIPTION')).toBeInTheDocument();
        const activeLink = screen.getByRole('link', { name: /IMAGES_SECTION_ALBUMS/i });
        expect(activeLink).toHaveAttribute('href', '/images/albums');
        expect(activeLink).toHaveAttribute('aria-current', 'page');
    });

    it('exposes the favorites tab and resolves its section from the route', () => {
        render(
            <MemoryRouter initialEntries={['/images/favorites']}>
                <ImageDomainHeader />
                <ImageDomainNav />
            </MemoryRouter>
        );

        expect(
            screen.getByRole('heading', { name: 'IMAGES_SECTION_FAVORITES' })
        ).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /IMAGES_SECTION_FAVORITES/i })).toHaveAttribute(
            'aria-current',
            'page'
        );
        expect(getImageSectionFromPath('/images/favorites')).toBe('favorites');
    });
});

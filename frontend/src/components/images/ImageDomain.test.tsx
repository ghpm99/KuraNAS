import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ImageDomainHeader from './ImageDomainHeader';
import ImageSidebar from './ImageSidebar';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

describe('components/images domain shell', () => {
    it('renders contextual header and active sidebar item from route', () => {
        render(
            <MemoryRouter initialEntries={['/images/albums']}>
                <ImageDomainHeader />
                <ImageSidebar />
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'IMAGES_SECTION_ALBUMS' })).toBeInTheDocument();
        expect(screen.getAllByText('IMAGES_SECTION_ALBUMS_DESCRIPTION')[0]).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /IMAGES_SECTION_ALBUMS/i })).toHaveAttribute(
            'href',
            '/images/albums'
        );
    });
});

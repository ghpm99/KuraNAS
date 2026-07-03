import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import VideoDomainHeader from './VideoDomainHeader';
import VideoDomainNav from './VideoDomainNav';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

describe('features/videos domain shell', () => {
    it('renders contextual header and active nav tab from route', () => {
        render(
            <MemoryRouter initialEntries={['/videos/folders/archive-home']}>
                <VideoDomainHeader />
                <VideoDomainNav />
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'VIDEO_SECTION_FOLDERS' })).toBeInTheDocument();
        expect(screen.getByText('VIDEO_SECTION_FOLDERS_DESCRIPTION')).toBeInTheDocument();
        const activeLink = screen.getByRole('link', { name: /VIDEO_SECTION_FOLDERS/i });
        expect(activeLink).toHaveAttribute('href', '/videos/folders');
        expect(activeLink).toHaveAttribute('aria-current', 'page');
    });
});

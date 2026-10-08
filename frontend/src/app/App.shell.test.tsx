import { render, screen } from '@testing-library/react';
import App from './App';
import { navigationItems } from '@/components/layout/navigationItems';

jest.mock('@/pages/home', () => () => <div>HomePagePlaceholder</div>);
jest.mock('@/pages/files', () => () => <div>FilesPagePlaceholder</div>);

const renderAppAt = (pathname: string) => {
    window.history.pushState({}, '', pathname);
    return render(<App />);
};

describe('App shell with real providers', () => {
    it.each([
        ['/home', 'HomePagePlaceholder'],
        ['/files', 'FilesPagePlaceholder'],
    ])('renders sidebar navigation and page placeholder at %s', async (pathname, placeholder) => {
        renderAppAt(pathname);

        expect(await screen.findByText(placeholder)).toBeInTheDocument();
        const sidebarLinks = screen
            .getAllByRole('navigation')
            .flatMap((navigation) => Array.from(navigation.querySelectorAll('a')));
        expect(sidebarLinks.length).toBeGreaterThanOrEqual(navigationItems.length);
        expect(screen.queryByText('SOMETHING_WENT_WRONG')).not.toBeInTheDocument();
    });
});

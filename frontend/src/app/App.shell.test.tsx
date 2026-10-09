import { act, fireEvent, render, screen } from '@testing-library/react';
import { onlineManager } from '@tanstack/react-query';
import App from './App';
import { navigationItems } from '@/components/layout/navigationItems';

jest.mock('@/pages/home', () => () => <div>HomePagePlaceholder</div>);
jest.mock('@/pages/files', () => () => <div>FilesPagePlaceholder</div>);
jest.mock('@/pages/settings', () => () => {
    throw new Error('settings-exploded');
});

const renderAppAt = (pathname: string) => {
    window.history.pushState({}, '', pathname);
    return render(<App />);
};

describe('App shell with real providers', () => {
    it('shows the route fallback inside the shell while a lazy page loads', async () => {
        renderAppAt('/home');

        expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
        expect(screen.getAllByRole('navigation').length).toBeGreaterThan(0);
        expect(await screen.findByText('HomePagePlaceholder')).toBeInTheDocument();
    });

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

    it('exposes the skip link, labelled landmarks and the document title inside the real shell', async () => {
        renderAppAt('/does/not/exist');
        await screen.findByRole('heading', { level: 1, name: 'Página não encontrada' });

        expect(screen.getByRole('link', { name: 'Pular para o conteúdo' })).toHaveAttribute(
            'href',
            '#main-content'
        );
        expect(screen.getByRole('main')).toHaveAttribute('id', 'main-content');
        expect(screen.getAllByRole('navigation', { name: 'Navegação principal' }).length).toBe(1);
        expect(screen.getByRole('navigation', { name: 'Navegação rápida' })).toBeInTheDocument();
        expect(document.title).toBe('Página não encontrada · KuraNAS');
        expect(document.documentElement.lang).not.toBe('');
    });

    it('renders the not found page inside the shell with the requested path', async () => {
        renderAppAt('/does/not/exist');

        expect(
            await screen.findByRole('heading', { level: 1, name: 'Página não encontrada' })
        ).toBeInTheDocument();
        expect(screen.getByText('Endereço solicitado: /does/not/exist')).toBeInTheDocument();
        expect(screen.getAllByRole('navigation').length).toBeGreaterThan(0);
    });

    it('keeps the shell visible when a page crashes and recovers on navigation', async () => {
        const consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => undefined);
        renderAppAt('/settings');

        expect(await screen.findByText('Algo deu errado')).toBeInTheDocument();
        expect(screen.getAllByRole('navigation').length).toBeGreaterThan(0);
        expect(screen.getByText('settings-exploded')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'Ir para o início' }));

        expect(await screen.findByText('HomePagePlaceholder')).toBeInTheDocument();
        consoleErrorSpy.mockRestore();
    });

    it('shows the connection banner with the embedded catalog when the backend is unreachable', async () => {
        renderAppAt('/home');
        await screen.findByText('HomePagePlaceholder');

        act(() => onlineManager.setOnline(false));
        expect(
            await screen.findByText('Servidor indisponível — tentando reconectar…')
        ).toBeInTheDocument();

        act(() => onlineManager.setOnline(true));
    });
});

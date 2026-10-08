import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, useNavigate } from 'react-router-dom';
import { AppShell } from './AppShell';

const mockUseAppShell = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

jest.mock('./useAppShell', () => ({
    useAppShell: () => mockUseAppShell(),
}));

jest.mock('../Header/Header', () => ({
    __esModule: true,
    default: () => <div data-testid="header">header</div>,
}));

jest.mock('../Sidebar/Sidebar', () => ({
    __esModule: true,
    default: () => <div data-testid="sidebar">sidebar</div>,
}));

describe('layout/AppShell', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders header, sidebar and body when queue is active', () => {
        mockUseAppShell.mockReturnValue({ hasQueue: true });

        render(
            <MemoryRouter>
                <AppShell>
                    <div>body</div>
                </AppShell>
            </MemoryRouter>
        );

        expect(screen.getByTestId('header')).toBeInTheDocument();
        expect(screen.getByTestId('sidebar')).toBeInTheDocument();
        expect(screen.getByText('body')).toBeInTheDocument();
    });

    it('marks the shell with data-has-player only when the queue is non-empty', () => {
        mockUseAppShell.mockReturnValue({ hasQueue: true });
        const { container, rerender } = render(
            <MemoryRouter>
                <AppShell>
                    <div>body</div>
                </AppShell>
            </MemoryRouter>
        );
        expect(container.firstElementChild).toHaveAttribute('data-has-player', 'true');

        mockUseAppShell.mockReturnValue({ hasQueue: false });
        rerender(
            <MemoryRouter>
                <AppShell>
                    <div>body</div>
                </AppShell>
            </MemoryRouter>
        );
        expect(container.firstElementChild).toHaveAttribute('data-has-player', 'false');
    });

    describe('navigation accessibility', () => {
        const NavigateButton = ({ to }: { to: string }) => {
            const navigate = useNavigate();
            return (
                <button type="button" onClick={() => navigate(to)}>
                    {`go ${to}`}
                </button>
            );
        };

        const renderShell = (children: React.ReactNode) => {
            mockUseAppShell.mockReturnValue({ hasQueue: false });
            return render(
                <MemoryRouter initialEntries={['/home']}>
                    <AppShell>{children}</AppShell>
                </MemoryRouter>
            );
        };

        it('renders the skip link first and moves focus to main when activated', () => {
            const { container } = renderShell(<div>body</div>);

            const skipLink = screen.getByRole('link', { name: 'SKIP_TO_CONTENT' });
            expect(container.querySelector('a')).toBe(skipLink);
            expect(skipLink).toHaveAttribute('href', '#main-content');

            fireEvent.click(skipLink);

            const mainElement = screen.getByRole('main');
            expect(mainElement).toHaveAttribute('id', 'main-content');
            expect(mainElement).toHaveAttribute('tabindex', '-1');
            expect(mainElement).toHaveFocus();
        });

        it('moves focus to the page heading and announces it when the pathname changes', () => {
            const { container } = renderShell(
                <>
                    <h1>Pagina</h1>
                    <NavigateButton to="/files" />
                </>
            );
            const heading = screen.getByRole('heading', { level: 1 });
            heading.setAttribute('tabindex', '-1');
            expect(heading).not.toHaveFocus();

            fireEvent.click(screen.getByRole('button', { name: 'go /files' }));

            expect(heading).toHaveFocus();
            expect(container.querySelector('[aria-live="polite"]')).toHaveTextContent('Pagina');
        });

        it('falls back to focusing main when the page has no heading', () => {
            renderShell(<NavigateButton to="/files" />);

            fireEvent.click(screen.getByRole('button', { name: 'go /files' }));

            expect(screen.getByRole('main')).toHaveFocus();
        });

        it('does not move focus on query-only navigation', () => {
            renderShell(
                <>
                    <h1>Pagina</h1>
                    <NavigateButton to="/home?tab=2" />
                </>
            );
            const heading = screen.getByRole('heading', { level: 1 });
            heading.setAttribute('tabindex', '-1');

            fireEvent.click(screen.getByRole('button', { name: 'go /home?tab=2' }));

            expect(heading).not.toHaveFocus();
            expect(screen.getByRole('main')).not.toHaveFocus();
        });
    });
});

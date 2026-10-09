import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import GlobalShortcutsProvider from './GlobalShortcutsProvider';
import { globalShortcutDefinitions } from './globalShortcutDefinitions';
import { useRegisterPageShortcuts } from './shortcutRegistry';
import { appRoutes } from '@/app/routes';
import { requestShortcutsHelp } from '@/components/layout/appCommandEvents';

const LocationProbe = () => <p data-testid="pathname">{useLocation().pathname}</p>;

const pageShortcutDefinitions = [{ keyLabels: ['F2'], descriptionKey: 'PAGE_RENAME' }];

const PageWithShortcuts = () => {
    useRegisterPageShortcuts(pageShortcutDefinitions);
    return null;
};

const renderProvider = (children: React.ReactNode = null) =>
    render(
        <MemoryRouter initialEntries={['/home']}>
            <GlobalShortcutsProvider>
                <LocationProbe />
                {children}
                <input aria-label="field" />
            </GlobalShortcutsProvider>
        </MemoryRouter>
    );

const press = (key: string, init: KeyboardEventInit = {}, target: Element | Document = document) => {
    act(() => {
        fireEvent.keyDown(target, { key, ...init });
    });
};

describe('GlobalShortcutsProvider', () => {
    afterEach(() => {
        jest.useRealTimers();
    });

    it('opens the help dialog when the shortcuts help command is requested', async () => {
        renderProvider();

        act(() => requestShortcutsHelp());

        expect(await screen.findByText('SHORTCUTS_DIALOG_TITLE')).toBeInTheDocument();
    });

    it('renders its children without a backend and keeps the dialog closed', () => {
        renderProvider();

        expect(screen.getByTestId('pathname')).toHaveTextContent('/home');
        expect(screen.queryByText('SHORTCUTS_DIALOG_TITLE')).not.toBeInTheDocument();
    });

    it('opens the global dialog with ? listing every global shortcut and no page section', async () => {
        renderProvider();

        press('?', { shiftKey: true });

        expect(await screen.findByText('SHORTCUTS_DIALOG_TITLE')).toBeInTheDocument();
        globalShortcutDefinitions.forEach((shortcut) => {
            expect(screen.getByText(shortcut.descriptionKey)).toBeInTheDocument();
        });
        expect(screen.queryByText('SHORTCUTS_SECTION_CURRENT_PAGE')).not.toBeInTheDocument();
    });

    it('adds a current page section when a page registers its shortcuts', async () => {
        renderProvider(<PageWithShortcuts />);

        press('?', { shiftKey: true });

        expect(await screen.findByText('SHORTCUTS_SECTION_CURRENT_PAGE')).toBeInTheDocument();
        expect(screen.getByText('PAGE_RENAME')).toBeInTheDocument();
    });

    it.each([
        ['h', appRoutes.home],
        ['f', appRoutes.files],
        ['i', appRoutes.images],
        ['m', appRoutes.music],
        ['v', appRoutes.videos],
        ['s', appRoutes.settings],
    ])('navigates with the g then %s sequence', (targetKey, expectedRoute) => {
        renderProvider();

        press('g');
        press(targetKey);

        expect(screen.getByTestId('pathname')).toHaveTextContent(expectedRoute);
    });

    it('cancels the sequence on an unknown second key or after the timeout', () => {
        jest.useFakeTimers();
        renderProvider();

        press('g');
        press('x');
        press('f');
        expect(screen.getByTestId('pathname')).toHaveTextContent('/home');

        press('g');
        act(() => {
            jest.advanceTimersByTime(2000);
        });
        press('f');
        expect(screen.getByTestId('pathname')).toHaveTextContent('/home');
    });

    it('ignores shortcuts while typing in an input', () => {
        renderProvider();
        const field = screen.getByLabelText('field');

        press('g', {}, field);
        press('f', {}, field);
        press('?', { shiftKey: true }, field);

        expect(screen.getByTestId('pathname')).toHaveTextContent('/home');
        expect(screen.queryByText('SHORTCUTS_DIALOG_TITLE')).not.toBeInTheDocument();
    });

    it('ignores shortcuts with modifier keys and while a dialog is open', async () => {
        renderProvider();
        press('g', { ctrlKey: true });
        press('f');
        expect(screen.getByTestId('pathname')).toHaveTextContent('/home');

        press('?', { shiftKey: true });
        await screen.findByText('SHORTCUTS_DIALOG_TITLE');
        press('g');
        press('f');
        expect(screen.getByTestId('pathname')).toHaveTextContent('/home');
    });

    it('closes the dialog with its close button', async () => {
        renderProvider();
        press('?', { shiftKey: true });

        fireEvent.click(await screen.findByRole('button', { name: 'CLOSE' }));

        await waitFor(() =>
            expect(screen.queryByText('SHORTCUTS_DIALOG_TITLE')).not.toBeInTheDocument()
        );
    });
});

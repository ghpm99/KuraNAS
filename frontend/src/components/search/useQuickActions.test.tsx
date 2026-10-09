import { act, renderHook } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { ColorSchemeContextProvider } from '@/components/providers/colorSchemeProvider/colorSchemeContext';
import type { ThemeMode } from '@/components/providers/colorSchemeProvider/colorScheme';
import {
    defaultSettingsConfiguration,
    SettingsContextProvider,
    type SettingsContextType,
} from '@/components/providers/settingsProvider/settingsContext';
import {
    subscribeToShortcutsHelp,
    subscribeToSidebarToggle,
} from '@/components/layout/appCommandEvents';
import { I18nContextProvider } from '@/components/i18n/provider/i18nContext';
import { quickActionDestinations } from './quickActionDestinations';
import { useQuickActions } from './useQuickActions';

const routerWrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter>{children}</MemoryRouter>
);

const mockNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useNavigate: () => mockNavigate,
}));

const buildThemeWrapper = (
    themeMode: ThemeMode,
    saveSettings: SettingsContextType['saveSettings']
) => {
    const settingsValue: SettingsContextType = {
        settings: defaultSettingsConfiguration,
        isLoading: false,
        isSaving: false,
        hasError: false,
        refresh: async () => undefined,
        saveSettings,
    };
    return ({ children }: { children: ReactNode }) => (
        <MemoryRouter>
            <SettingsContextProvider value={settingsValue}>
                <ColorSchemeContextProvider value={{ themeMode, colorScheme: 'dark' }}>
                    {children}
                </ColorSchemeContextProvider>
            </SettingsContextProvider>
        </MemoryRouter>
    );
};

const findAction = (actions: ReturnType<typeof useQuickActions>, actionId: string) => {
    const action = actions.find((candidate) => candidate.id === actionId);
    if (!action) {
        throw new Error(`missing action ${actionId}`);
    }
    return action;
};

describe('search/useQuickActions', () => {
    beforeEach(() => {
        mockNavigate.mockReset();
    });

    it('builds the actions without any provider or backend', () => {
        const { result } = renderHook(() => useQuickActions(), { wrapper: routerWrapper });

        expect(result.current.length).toBeGreaterThanOrEqual(quickActionDestinations.length);
        expect(result.current.every((action) => action.label !== '')).toBe(true);
        expect(result.current.some((action) => action.id === 'action-toggle-theme')).toBe(false);
    });

    it('creates one navigation action per destination', () => {
        const { result } = renderHook(() => useQuickActions(), { wrapper: routerWrapper });

        quickActionDestinations.forEach((destination) => {
            mockNavigate.mockClear();
            act(() =>
                findAction(result.current, `action-destination-${destination.route}`).onSelect()
            );
            expect(mockNavigate).toHaveBeenCalledWith(destination.route);
        });
    });

    it('labels domain sections with their domain', () => {
        const interpolatingTranslate = (key: string, options?: Record<string, string>) =>
            options ? `${key}|${Object.values(options).join('|')}` : key;
        const { result } = renderHook(() => useQuickActions(), {
            wrapper: ({ children }: { children: ReactNode }) => (
                <I18nContextProvider value={{ t: interpolatingTranslate }}>
                    <MemoryRouter>{children}</MemoryRouter>
                </I18nContextProvider>
            ),
        });

        const musicArtists = findAction(result.current, 'action-destination-/music/artists');
        expect(musicArtists.label).toBe(
            'GLOBAL_SEARCH_ACTION_SECTION_LABEL|NAV_MUSIC|MUSIC_ARTISTS'
        );
    });

    it('toggles the sidebar through the app command', () => {
        const onToggle = jest.fn();
        const unsubscribe = subscribeToSidebarToggle(onToggle);
        const { result } = renderHook(() => useQuickActions(), { wrapper: routerWrapper });

        act(() => findAction(result.current, 'action-toggle-sidebar').onSelect());

        expect(onToggle).toHaveBeenCalledTimes(1);
        unsubscribe();
    });

    it('opens the keyboard shortcuts help through the app command', () => {
        const onShowHelp = jest.fn();
        const unsubscribe = subscribeToShortcutsHelp(onShowHelp);
        const { result } = renderHook(() => useQuickActions(), { wrapper: routerWrapper });

        act(() => findAction(result.current, 'action-show-shortcuts').onSelect());

        expect(onShowHelp).toHaveBeenCalledTimes(1);
        unsubscribe();
    });

    it.each([
        ['dark', 'light'],
        ['light', 'system'],
        ['system', 'dark'],
    ] as const)('cycles the theme from %s to %s by saving the settings', async (current, next) => {
        const saveSettings = jest.fn().mockResolvedValue(defaultSettingsConfiguration);
        const { result } = renderHook(() => useQuickActions(), {
            wrapper: buildThemeWrapper(current, saveSettings),
        });

        await act(async () => findAction(result.current, 'action-toggle-theme').onSelect());

        expect(saveSettings).toHaveBeenCalledWith(
            expect.objectContaining({
                appearance: expect.objectContaining({ theme_mode: next }),
            })
        );
    });

    it('reports a failed theme save instead of throwing', async () => {
        const saveSettings = jest.fn().mockRejectedValue(new Error('boom'));
        const { result } = renderHook(() => useQuickActions(), {
            wrapper: buildThemeWrapper('dark', saveSettings),
        });

        await expect(
            act(async () => findAction(result.current, 'action-toggle-theme').onSelect())
        ).resolves.toBeUndefined();
    });
});

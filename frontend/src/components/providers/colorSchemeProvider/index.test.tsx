import { act, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SettingsProvider from '@/components/providers/settingsProvider';
import { getSettingsConfiguration } from '@/service/configuration';
import ColorSchemeProvider from './index';
import { useColorScheme } from './colorSchemeContext';
import { applyPersistedColorScheme, readStoredThemeMode, themeModeStorageKey } from './colorScheme';

jest.mock('@/service/configuration', () => ({
    getSettingsConfiguration: jest.fn(),
    updateSettingsConfiguration: jest.fn(),
}));

const mockedGetSettingsConfiguration = getSettingsConfiguration as jest.Mock;

type MediaQueryListener = (event: { matches: boolean }) => void;

const installMatchMedia = (initiallyMatchesDark: boolean) => {
    const listeners = new Set<MediaQueryListener>();
    let matchesDark = initiallyMatchesDark;
    Object.defineProperty(window, 'matchMedia', {
        configurable: true,
        writable: true,
        value: (query: string) => ({
            get matches() {
                return query.includes('prefers-color-scheme: dark') ? matchesDark : false;
            },
            media: query,
            addEventListener: (_: string, listener: MediaQueryListener) => listeners.add(listener),
            removeEventListener: (_: string, listener: MediaQueryListener) =>
                listeners.delete(listener),
            addListener: jest.fn(),
            removeListener: jest.fn(),
        }),
    });
    return (nextMatchesDark: boolean) => {
        matchesDark = nextMatchesDark;
        listeners.forEach((listener) => listener({ matches: nextMatchesDark }));
    };
};

const ColorSchemeProbe = () => {
    const { themeMode, colorScheme } = useColorScheme();
    return <p>{`${themeMode}:${colorScheme}`}</p>;
};

const renderProvider = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={queryClient}>
            <SettingsProvider>
                <ColorSchemeProvider>
                    <ColorSchemeProbe />
                </ColorSchemeProvider>
            </SettingsProvider>
        </QueryClientProvider>
    );
};

const settingsWithThemeMode = (themeMode?: string) => ({
    appearance: { accent_color: 'violet', reduce_motion: false, theme_mode: themeMode },
});

describe('components/providers/colorSchemeProvider', () => {
    const originalMatchMedia = window.matchMedia;

    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
        document.documentElement.removeAttribute('data-theme');
        document.documentElement.removeAttribute('style');
        document.head.querySelector('meta[name="theme-color"]')?.remove();
    });

    afterEach(() => {
        Object.defineProperty(window, 'matchMedia', {
            configurable: true,
            writable: true,
            value: originalMatchMedia,
        });
    });

    it('renders children and falls back to the dark scheme without any backend or matchMedia', async () => {
        mockedGetSettingsConfiguration.mockRejectedValue(new Error('offline'));

        renderProvider();

        expect(screen.getByText('dark:dark')).toBeInTheDocument();
        await waitFor(() => expect(document.documentElement.dataset.theme).toBe('dark'));
        expect(document.documentElement.style.colorScheme).toBe('dark');
    });

    it('keeps the persisted mode while the server value is unavailable', async () => {
        window.localStorage.setItem(themeModeStorageKey, 'light');
        mockedGetSettingsConfiguration.mockRejectedValue(new Error('offline'));

        renderProvider();

        await waitFor(() => expect(document.documentElement.dataset.theme).toBe('light'));
        expect(screen.getByText('light:light')).toBeInTheDocument();
    });

    it('syncs with the server value and persists it', async () => {
        window.localStorage.setItem(themeModeStorageKey, 'dark');
        mockedGetSettingsConfiguration.mockResolvedValue(settingsWithThemeMode('light'));

        renderProvider();

        await waitFor(() => expect(screen.getByText('light:light')).toBeInTheDocument());
        expect(readStoredThemeMode()).toBe('light');
        expect(document.documentElement.dataset.theme).toBe('light');
        expect(document.documentElement.style.colorScheme).toBe('light');
    });

    it('treats a payload without theme_mode as the dark default', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue({ indexing: {} });

        renderProvider();

        await waitFor(() => expect(readStoredThemeMode()).toBe('dark'));
        expect(screen.getByText('dark:dark')).toBeInTheDocument();
    });

    it('follows prefers-color-scheme changes live when the mode is system', async () => {
        const emitSystemPreference = installMatchMedia(true);
        mockedGetSettingsConfiguration.mockResolvedValue(settingsWithThemeMode('system'));

        renderProvider();

        await waitFor(() => expect(screen.getByText('system:dark')).toBeInTheDocument());

        act(() => emitSystemPreference(false));
        await waitFor(() => expect(screen.getByText('system:light')).toBeInTheDocument());
        expect(document.documentElement.dataset.theme).toBe('light');

        act(() => emitSystemPreference(true));
        await waitFor(() => expect(document.documentElement.dataset.theme).toBe('dark'));
    });

    it('ignores system preference changes when the mode is fixed', async () => {
        const emitSystemPreference = installMatchMedia(true);
        mockedGetSettingsConfiguration.mockResolvedValue(settingsWithThemeMode('dark'));

        renderProvider();
        await waitFor(() => expect(screen.getByText('dark:dark')).toBeInTheDocument());

        act(() => emitSystemPreference(false));

        expect(screen.getByText('dark:dark')).toBeInTheDocument();
    });

    it('updates the meta theme-color with the scheme', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue(settingsWithThemeMode('light'));

        renderProvider();

        await waitFor(() => {
            const themeColorMeta = document.head.querySelector('meta[name="theme-color"]');
            expect(themeColorMeta?.getAttribute('content')).toBe('#F3F5FA');
        });
    });
});

describe('colorScheme persisted pre-render value', () => {
    beforeEach(() => {
        window.localStorage.clear();
        document.documentElement.removeAttribute('data-theme');
        document.documentElement.removeAttribute('style');
    });

    it('applies the dark scheme when nothing was persisted', () => {
        applyPersistedColorScheme();

        expect(document.documentElement.dataset.theme).toBe('dark');
    });

    it('applies the persisted light scheme before React renders', () => {
        window.localStorage.setItem(themeModeStorageKey, 'light');

        applyPersistedColorScheme();

        expect(document.documentElement.dataset.theme).toBe('light');
        expect(document.documentElement.style.colorScheme).toBe('light');
    });

    it('resolves a persisted system mode through the current system preference', () => {
        window.localStorage.setItem(themeModeStorageKey, 'system');
        const originalMatchMedia = window.matchMedia;
        installMatchMedia(false);

        applyPersistedColorScheme();

        expect(document.documentElement.dataset.theme).toBe('light');
        Object.defineProperty(window, 'matchMedia', {
            configurable: true,
            writable: true,
            value: originalMatchMedia,
        });
    });

    it('ignores an invalid persisted value', () => {
        window.localStorage.setItem(themeModeStorageKey, 'sepia');

        applyPersistedColorScheme();

        expect(readStoredThemeMode()).toBeNull();
        expect(document.documentElement.dataset.theme).toBe('dark');
    });

    it('survives a storage that throws', () => {
        const getItemSpy = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('blocked');
        });

        expect(() => applyPersistedColorScheme()).not.toThrow();
        expect(document.documentElement.dataset.theme).toBe('dark');
        getItemSpy.mockRestore();
    });
});

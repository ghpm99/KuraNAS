import { colorSchemeTokens, type ColorScheme } from '@/theme/visualTokens';

export type ThemeMode = 'dark' | 'light' | 'system';

export const themeModeStorageKey = 'kuranas.themeMode';
export const defaultThemeMode: ThemeMode = 'dark';
export const systemDarkMediaQuery = '(prefers-color-scheme: dark)';

const themeModes: readonly ThemeMode[] = ['dark', 'light', 'system'];

export const isThemeMode = (candidate: unknown): candidate is ThemeMode =>
    themeModes.includes(candidate as ThemeMode);

export const readStoredThemeMode = (): ThemeMode | null => {
    try {
        const storedThemeMode = window.localStorage.getItem(themeModeStorageKey);
        return isThemeMode(storedThemeMode) ? storedThemeMode : null;
    } catch {
        return null;
    }
};

export const storeThemeMode = (themeMode: ThemeMode) => {
    try {
        window.localStorage.setItem(themeModeStorageKey, themeMode);
    } catch {
        return;
    }
};

export const readSystemPrefersDark = (): boolean => {
    if (typeof window.matchMedia !== 'function') {
        return true;
    }
    return window.matchMedia(systemDarkMediaQuery).matches;
};

export const resolveColorScheme = (
    themeMode: ThemeMode,
    systemPrefersDark: boolean
): ColorScheme => {
    if (themeMode === 'system') {
        return systemPrefersDark ? 'dark' : 'light';
    }
    return themeMode;
};

const findOrCreateThemeColorMeta = (): HTMLMetaElement => {
    const existingMeta = document.head.querySelector<HTMLMetaElement>('meta[name="theme-color"]');
    if (existingMeta) {
        return existingMeta;
    }
    const createdMeta = document.createElement('meta');
    createdMeta.name = 'theme-color';
    document.head.appendChild(createdMeta);
    return createdMeta;
};

export const applyColorScheme = (colorScheme: ColorScheme) => {
    const rootElement = document.documentElement;
    const backgroundColor = colorSchemeTokens[colorScheme].colors.backgroundRoot;
    rootElement.dataset.theme = colorScheme;
    rootElement.style.colorScheme = colorScheme;
    rootElement.style.backgroundColor = backgroundColor;
    findOrCreateThemeColorMeta().content = colorSchemeTokens[colorScheme].metaThemeColor;
};

export const applyPersistedColorScheme = () => {
    const persistedThemeMode = readStoredThemeMode() ?? defaultThemeMode;
    applyColorScheme(resolveColorScheme(persistedThemeMode, readSystemPrefersDark()));
};

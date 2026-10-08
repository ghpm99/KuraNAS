import { ThemeProvider } from '@mui/material/styles';
import { useEffect, useMemo } from 'react';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import { createAppTheme } from '@/theme/appTheme';
import {
    applyColorScheme,
    defaultThemeMode,
    isThemeMode,
    readStoredThemeMode,
    resolveColorScheme,
    storeThemeMode,
    type ThemeMode,
} from './colorScheme';
import { ColorSchemeContextProvider } from './colorSchemeContext';
import { useSystemPrefersDark } from './useSystemPrefersDark';

const ColorSchemeProvider = ({ children }: { children: React.ReactNode }) => {
    const { settings, isLoading, hasError } = useSettings();
    const systemPrefersDark = useSystemPrefersDark();

    const serverThemeMode = settings.appearance?.theme_mode;
    const isServerThemeModeKnown = !isLoading && !hasError && isThemeMode(serverThemeMode);
    const themeMode: ThemeMode = isServerThemeModeKnown
        ? serverThemeMode
        : (readStoredThemeMode() ?? defaultThemeMode);
    const colorScheme = resolveColorScheme(themeMode, systemPrefersDark);

    useEffect(() => {
        if (isServerThemeModeKnown) {
            storeThemeMode(themeMode);
        }
    }, [isServerThemeModeKnown, themeMode]);

    useEffect(() => {
        applyColorScheme(colorScheme);
    }, [colorScheme]);

    const muiTheme = useMemo(() => createAppTheme(colorScheme), [colorScheme]);
    const contextValue = useMemo(() => ({ themeMode, colorScheme }), [themeMode, colorScheme]);

    return (
        <ColorSchemeContextProvider value={contextValue}>
            <ThemeProvider theme={muiTheme}>{children}</ThemeProvider>
        </ColorSchemeContextProvider>
    );
};

export default ColorSchemeProvider;

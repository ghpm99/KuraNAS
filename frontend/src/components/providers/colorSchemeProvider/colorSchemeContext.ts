import { createContext, useContext } from 'react';
import type { ColorScheme } from '@/theme/visualTokens';
import { defaultThemeMode, type ThemeMode } from './colorScheme';

export type ColorSchemeContextType = {
    themeMode: ThemeMode;
    colorScheme: ColorScheme;
};

const ColorSchemeContext = createContext<ColorSchemeContextType>({
    themeMode: defaultThemeMode,
    colorScheme: 'dark',
});

export const ColorSchemeContextProvider = ColorSchemeContext.Provider;

export const useColorScheme = () => useContext(ColorSchemeContext);

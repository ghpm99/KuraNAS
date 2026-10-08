import { createAppTheme } from './appTheme';
import { buildColorSchemeCssVariables, colorSchemeTokens } from './visualTokens';

describe('theme/visualTokens', () => {
    it('defines every css variable of the dark scheme in the light scheme', () => {
        expect(Object.keys(buildColorSchemeCssVariables('light')).sort()).toEqual(
            Object.keys(buildColorSchemeCssVariables('dark')).sort()
        );
    });

    it('defines every color token of the dark scheme in the light scheme', () => {
        expect(Object.keys(colorSchemeTokens.light.colors).sort()).toEqual(
            Object.keys(colorSchemeTokens.dark.colors).sort()
        );
    });

    it('inverts the ink channel between schemes', () => {
        expect(buildColorSchemeCssVariables('dark')['--app-color-ink-rgb']).toBe('255, 255, 255');
        expect(buildColorSchemeCssVariables('light')['--app-color-ink-rgb']).not.toBe(
            '255, 255, 255'
        );
    });
});

describe('theme/appTheme', () => {
    it.each(['dark', 'light'] as const)('builds the %s MUI palette from its tokens', (scheme) => {
        const theme = createAppTheme(scheme);

        expect(theme.palette.mode).toBe(scheme);
        expect(theme.palette.background.default).toBe(
            colorSchemeTokens[scheme].colors.backgroundRoot
        );
        expect(theme.palette.text.primary).toBe(colorSchemeTokens[scheme].colors.textPrimary);
    });

    it('emits css variables per data-theme selector', () => {
        const overrides = createAppTheme('light').components?.MuiCssBaseline
            ?.styleOverrides as Record<string, Record<string, string>>;

        expect(overrides[':root[data-theme="light"]']?.['--app-color-text-primary']).toBe(
            colorSchemeTokens.light.colors.textPrimary
        );
        expect(overrides[':root[data-theme="dark"]']?.['--app-color-text-primary']).toBe(
            colorSchemeTokens.dark.colors.textPrimary
        );
    });
});

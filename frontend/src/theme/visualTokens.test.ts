import { createAppTheme } from './appTheme';
import {
    belowBreakpointMediaQuery,
    breakpointsPx,
    buildColorSchemeCssVariables,
    colorSchemeTokens,
    structuralCssVariables,
    viewportMediaQueries,
} from './visualTokens';

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

describe('breakpoint tokens', () => {
    it('keeps the four breakpoints in ascending order', () => {
        expect(breakpointsPx).toEqual({ phone: 600, tablet: 900, desktop: 1200, wide: 1600 });
    });

    it('builds the below-breakpoint query just under the breakpoint', () => {
        expect(viewportMediaQueries.belowPhone).toBe('(max-width: 599.95px)');
        expect(viewportMediaQueries.belowTablet).toBe('(max-width: 899.95px)');
        expect(viewportMediaQueries.belowDesktop).toBe('(max-width: 1199.95px)');
        expect(belowBreakpointMediaQuery(breakpointsPx.wide)).toBe('(max-width: 1599.95px)');
    });

    it('aligns the MUI theme breakpoints with the tokens', () => {
        const { values } = createAppTheme('dark').breakpoints;
        expect([values.sm, values.md, values.lg, values.xl]).toEqual([600, 900, 1200, 1600]);
    });

    it('builds the compact desktop query between tablet and desktop', () => {
        expect(viewportMediaQueries.compactDesktop).toBe(
            '(min-width: 900px) and (max-width: 1199.95px)'
        );
    });
});

describe('intrinsic size tokens', () => {
    it('exposes the offscreen placeholder size of every long list item as css variables', () => {
        expect(structuralCssVariables).toEqual(
            expect.objectContaining({
                '--app-intrinsic-file-card-height': '220px',
                '--app-intrinsic-file-row-height': '56px',
                '--app-intrinsic-image-tile-height': '180px',
                '--app-intrinsic-track-row-height': '56px',
                '--app-intrinsic-video-card-height': '240px',
            })
        );
    });
});

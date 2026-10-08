import { createTheme } from '@mui/material/styles';
import {
    buildColorSchemeCssVariables,
    colorSchemeTokens,
    structuralCssVariables,
    visualTokens,
    type ColorScheme,
} from './visualTokens';

export const coarsePointerMediaQuery = '@media (pointer: coarse)';
export const minimumTouchTargetPx = 44;

export const createAppTheme = (colorScheme: ColorScheme) => {
    const schemeTokens = colorSchemeTokens[colorScheme];
    const schemeColors = schemeTokens.colors;
    return createTheme({
        palette: {
            mode: colorScheme,
            primary: {
                main: schemeColors.primary,
                light: schemeColors.primaryHover,
                dark: '#5848de',
            },
            secondary: {
                main: schemeColors.cyanAccent,
                light: '#38d6eb',
                dark: '#0891b2',
            },
            success: {
                main: schemeColors.success,
            },
            warning: {
                main: schemeColors.warning,
            },
            error: {
                main: schemeColors.danger,
            },
            background: {
                default: schemeColors.backgroundRoot,
                paper: schemeColors.surface2,
            },
            text: {
                primary: schemeColors.textPrimary,
                secondary: schemeColors.textSecondary,
                disabled: schemeColors.textDisabled,
            },
            divider: schemeColors.borderSubtle,
        },
        shape: {
            borderRadius: 14,
        },
        typography: {
            fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
            h1: {
                fontSize: '1.75rem',
                lineHeight: 1.285,
                fontWeight: 600,
            },
            h2: {
                fontSize: '1.5rem',
                lineHeight: 1.333,
                fontWeight: 600,
            },
            h3: {
                fontSize: '1.25rem',
                lineHeight: 1.4,
                fontWeight: 600,
            },
            body1: {
                fontSize: '0.875rem',
                lineHeight: 1.57,
            },
            body2: {
                fontSize: '0.8125rem',
                lineHeight: 1.54,
            },
            caption: {
                fontSize: '0.75rem',
                lineHeight: 1.5,
                fontWeight: 500,
            },
        },
        components: {
            MuiCssBaseline: {
                styleOverrides: {
                    ':root': {
                        ...structuralCssVariables,
                        ...buildColorSchemeCssVariables('dark'),
                    },
                    ':root[data-theme="dark"]': buildColorSchemeCssVariables('dark'),
                    ':root[data-theme="light"]': buildColorSchemeCssVariables('light'),
                    '*': {
                        boxSizing: 'border-box',
                    },
                    'html, body, #root': {
                        height: '100%',
                    },
                    body: {
                        margin: 0,
                        backgroundColor: schemeColors.backgroundRoot,
                        backgroundImage: 'var(--app-background-app)',
                        backgroundAttachment: 'fixed',
                        color: schemeColors.textPrimary,
                    },
                    a: {
                        color: 'inherit',
                        textDecoration: 'none',
                    },
                    '*::-webkit-scrollbar': {
                        width: '10px',
                        height: '10px',
                    },
                    '@media (max-width: 640px)': {
                        '*::-webkit-scrollbar': {
                            width: '4px',
                            height: '4px',
                        },
                    },
                    '*::-webkit-scrollbar-track': {
                        backgroundColor: schemeTokens.scrollbar.track,
                    },
                    '*::-webkit-scrollbar-thumb': {
                        backgroundColor: schemeTokens.scrollbar.thumb,
                        borderRadius: visualTokens.radius.pill,
                        border: '2px solid transparent',
                        backgroundClip: 'padding-box',
                    },
                },
            },
            MuiCard: {
                styleOverrides: {
                    root: {
                        backgroundImage: 'var(--app-background-panel)',
                        backgroundColor: schemeColors.surface2,
                        border: `1px solid ${schemeColors.borderSubtle}`,
                        borderRadius: visualTokens.radius.lg,
                        boxShadow: schemeTokens.shadow.card,
                    },
                },
            },
            MuiPaper: {
                styleOverrides: {
                    root: {
                        backgroundImage: 'none',
                    },
                },
            },
            MuiDrawer: {
                styleOverrides: {
                    paper: {
                        backgroundImage: 'var(--app-background-panel-elevated)',
                        backgroundColor: schemeColors.surface2,
                        borderColor: schemeColors.borderSubtle,
                    },
                },
            },
            MuiButton: {
                styleOverrides: {
                    root: {
                        borderRadius: visualTokens.radius.md,
                        textTransform: 'none',
                        fontWeight: 600,
                    },
                },
            },
            MuiInputBase: {
                styleOverrides: {
                    root: {
                        borderRadius: visualTokens.radius.pill,
                    },
                    input: {
                        padding: 0,
                    },
                },
            },
            MuiListItemButton: {
                styleOverrides: {
                    root: {
                        borderRadius: visualTokens.radius.md,
                        transition: `background-color ${visualTokens.motion.fast} ease, border-color ${visualTokens.motion.fast} ease, box-shadow ${visualTokens.motion.fast} ease`,
                        '&.Mui-selected': {
                            backgroundColor: 'rgba(var(--app-color-primary-rgb), 0.12)',
                        },
                    },
                },
            },
            MuiIconButton: {
                styleOverrides: {
                    root: {
                        transition: `background-color ${visualTokens.motion.fast} ease, border-color ${visualTokens.motion.fast} ease`,
                    },
                    sizeSmall: {
                        [coarsePointerMediaQuery]: {
                            minWidth: minimumTouchTargetPx,
                            minHeight: minimumTouchTargetPx,
                        },
                    },
                },
            },
            MuiAvatar: {
                styleOverrides: {
                    root: {
                        border: `1px solid ${schemeColors.borderStrong}`,
                    },
                },
            },
        },
    });
};

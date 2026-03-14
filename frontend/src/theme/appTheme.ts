import { createTheme } from '@mui/material/styles';
import { appCssVariables, visualTokens } from './visualTokens';

export const appTheme = createTheme({
	palette: {
		mode: 'dark',
		primary: {
			main: visualTokens.colors.primary,
			light: visualTokens.colors.primaryHover,
			dark: '#5848de',
		},
		secondary: {
			main: visualTokens.colors.cyanAccent,
			light: '#38d6eb',
			dark: '#0891b2',
		},
		success: {
			main: visualTokens.colors.success,
		},
		warning: {
			main: visualTokens.colors.warning,
		},
		error: {
			main: visualTokens.colors.danger,
		},
		background: {
			default: visualTokens.colors.backgroundRoot,
			paper: visualTokens.colors.surface2,
		},
		text: {
			primary: visualTokens.colors.textPrimary,
			secondary: visualTokens.colors.textSecondary,
			disabled: visualTokens.colors.textDisabled,
		},
		divider: visualTokens.colors.borderSubtle,
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
				':root': appCssVariables,
				'*': {
					boxSizing: 'border-box',
				},
				'html, body, #root': {
					height: '100%',
				},
				body: {
					margin: 0,
					backgroundColor: visualTokens.colors.backgroundRoot,
					backgroundImage: visualTokens.backgrounds.app,
					backgroundAttachment: 'fixed',
					color: visualTokens.colors.textPrimary,
				},
				a: {
					color: 'inherit',
					textDecoration: 'none',
				},
				'*::-webkit-scrollbar': {
					width: '10px',
					height: '10px',
				},
				'*::-webkit-scrollbar-track': {
					backgroundColor: 'rgba(7, 10, 16, 0.26)',
				},
				'*::-webkit-scrollbar-thumb': {
					backgroundColor: 'rgba(126, 138, 163, 0.38)',
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
					backgroundColor: visualTokens.colors.surface2,
					border: `1px solid ${visualTokens.colors.borderSubtle}`,
					borderRadius: visualTokens.radius.lg,
					boxShadow: visualTokens.shadow.card,
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
					backgroundColor: visualTokens.colors.surface2,
					borderColor: visualTokens.colors.borderSubtle,
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
						backgroundColor: 'rgba(109, 93, 246, 0.12)',
					},
				},
			},
		},
		MuiIconButton: {
			styleOverrides: {
				root: {
					transition: `background-color ${visualTokens.motion.fast} ease, border-color ${visualTokens.motion.fast} ease`,
				},
			},
		},
		MuiAvatar: {
			styleOverrides: {
				root: {
					border: `1px solid ${visualTokens.colors.borderStrong}`,
				},
			},
		},
	},
});

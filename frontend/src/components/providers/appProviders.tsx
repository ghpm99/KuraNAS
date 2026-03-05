import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { StrictMode } from 'react';
import { BrowserRouter } from 'react-router-dom';
import I18nProvider from '../i18n/provider';

const darkTheme = createTheme({
	palette: {
		mode: 'dark',
		primary: {
			main: '#6366f1',
			light: '#818cf8',
			dark: '#4f46e5',
		},
		secondary: {
			main: '#a78bfa',
			light: '#c4b5fd',
			dark: '#7c3aed',
		},
		background: {
			default: '#0f0f13',
			paper: '#1a1a24',
		},
		text: {
			primary: '#e4e4e7',
			secondary: '#a1a1aa',
		},
		divider: 'rgba(255, 255, 255, 0.06)',
	},
	shape: {
		borderRadius: 12,
	},
	typography: {
		fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
	},
	components: {
		MuiCard: {
			styleOverrides: {
				root: {
					backgroundImage: 'none',
					backgroundColor: '#1a1a24',
					border: '1px solid rgba(255, 255, 255, 0.06)',
					borderRadius: 16,
					transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
					'&:hover': {
						borderColor: 'rgba(255, 255, 255, 0.12)',
						boxShadow: '0 4px 24px rgba(0, 0, 0, 0.3)',
					},
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
		MuiButton: {
			styleOverrides: {
				root: {
					borderRadius: 8,
					textTransform: 'none' as const,
				},
			},
		},
		MuiInputBase: {
			styleOverrides: {
				root: {
					borderRadius: 10,
				},
			},
		},
		MuiListItemButton: {
			styleOverrides: {
				root: {
					borderRadius: 8,
					'&.Mui-selected': {
						backgroundColor: 'rgba(99, 102, 241, 0.12)',
						'&:hover': {
							backgroundColor: 'rgba(99, 102, 241, 0.18)',
						},
					},
				},
			},
		},
		MuiIconButton: {
			styleOverrides: {
				root: {
					transition: 'background-color 0.2s ease',
				},
			},
		},
	},
});

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			refetchOnWindowFocus: false,
			refetchOnReconnect: false,
			staleTime: 1000 * 60 * 5, // 5 minutes
		},
	},
});

const AppProviders = ({ children }: { children: React.ReactNode }) => {
	return (
		<QueryClientProvider client={queryClient}>
			<StrictMode>
				<I18nProvider>
					<SnackbarProvider maxSnack={3} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
						<ThemeProvider theme={darkTheme}>
							<CssBaseline />
							<BrowserRouter>{children}</BrowserRouter>
						</ThemeProvider>
					</SnackbarProvider>
				</I18nProvider>
			</StrictMode>
		</QueryClientProvider>
	);
};

export default AppProviders;

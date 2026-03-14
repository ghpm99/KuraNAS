import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider } from '@mui/material/styles';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { StrictMode } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { appTheme } from '@/theme/appTheme';
import I18nProvider from '../i18n/provider';

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
						<ThemeProvider theme={appTheme}>
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

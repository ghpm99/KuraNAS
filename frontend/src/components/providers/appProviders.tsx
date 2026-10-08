import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { StrictMode } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { appTheme } from '@/theme/appTheme';
import I18nProvider from '../i18n/provider';
import GlobalSearchProvider from '../search/GlobalSearchProvider';
import NotificationProvider from './notificationProvider';
import SettingsProvider from './settingsProvider';

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
    const isCompactViewport = useMediaQuery('(max-width: 900px)');
    const toastHorizontalAnchor = isCompactViewport ? 'center' : 'right';

    return (
        <QueryClientProvider client={queryClient}>
            <StrictMode>
                <I18nProvider>
                    <SettingsProvider>
                        <SnackbarProvider
                            maxSnack={3}
                            anchorOrigin={{ vertical: 'bottom', horizontal: toastHorizontalAnchor }}
                        >
                            <ThemeProvider theme={appTheme}>
                                <CssBaseline />
                                <BrowserRouter>
                                    <NotificationProvider>
                                        <GlobalSearchProvider>{children}</GlobalSearchProvider>
                                    </NotificationProvider>
                                </BrowserRouter>
                            </ThemeProvider>
                        </SnackbarProvider>
                    </SettingsProvider>
                </I18nProvider>
            </StrictMode>
        </QueryClientProvider>
    );
};

export default AppProviders;

import CssBaseline from '@mui/material/CssBaseline';
import useMediaQuery from '@mui/material/useMediaQuery';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { StrictMode } from 'react';
import { BrowserRouter } from 'react-router-dom';
import I18nProvider from '../i18n/provider';
import GlobalSearchProvider from '../search/GlobalSearchProvider';
import GlobalShortcutsProvider from '../shortcuts/GlobalShortcutsProvider';
import DocumentLanguageSync from '../layout/documentMetadata/DocumentLanguageSync';
import NotificationProvider from './notificationProvider';
import ColorSchemeProvider from './colorSchemeProvider';
import SettingsProvider from './settingsProvider';
import { shouldRetryQuery } from './queryRetryPolicy';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            refetchOnWindowFocus: false,
            refetchOnReconnect: true,
            retry: shouldRetryQuery,
            staleTime: 1000 * 60 * 5,
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
                            <ColorSchemeProvider>
                                <CssBaseline />
                                <DocumentLanguageSync />
                                <BrowserRouter>
                                    <NotificationProvider>
                                        <GlobalSearchProvider>
                                            <GlobalShortcutsProvider>
                                                {children}
                                            </GlobalShortcutsProvider>
                                        </GlobalSearchProvider>
                                    </NotificationProvider>
                                </BrowserRouter>
                            </ColorSchemeProvider>
                        </SnackbarProvider>
                    </SettingsProvider>
                </I18nProvider>
            </StrictMode>
        </QueryClientProvider>
    );
};

export default AppProviders;

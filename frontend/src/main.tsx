import './index.css';
import { StrictMode } from 'react';
import App from './app/App.tsx';

import FileProvider from './components/providers/fileprovider/index.tsx';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createRoot } from 'react-dom/client';
import I18nProvider from './components/i18n/provider/index.tsx';

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			refetchOnWindowFocus: false,
			refetchOnReconnect: false,
			staleTime: 1000 * 60 * 5, // 5 minutes
		},
	},
});

createRoot(document.getElementById('root')!).render(
	<QueryClientProvider client={queryClient}>
		<StrictMode>
			<I18nProvider>
				<FileProvider>
					<App />
				</FileProvider>
			</I18nProvider>
		</StrictMode>
	</QueryClientProvider>
);

import './index.css';
import { StrictMode } from 'react';
import App from './app/App.tsx';

import FileProvider from './components/providers/fileprovider/index.tsx';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createRoot } from 'react-dom/client'

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			refetchOnWindowFocus: false,
			refetchOnReconnect: false,
			staleTime: 1000 * 60 * 5, // 5 minutes
		},
	},
})

createRoot(document.getElementById('root')!).render(
	<QueryClientProvider client={queryClient}>
		<StrictMode>
			<FileProvider>
				<App />
			</FileProvider>
		</StrictMode>
	</QueryClientProvider>
);

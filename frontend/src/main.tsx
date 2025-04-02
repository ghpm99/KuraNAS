import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import App from './app/App.tsx';
import { PersistQueryClientProvider } from '@tanstack/react-query-persist-client';
import { QueryClient } from '@tanstack/react-query';
import { createSyncStoragePersister } from '@tanstack/query-sync-storage-persister';

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			gcTime: 1000 * 60 * 60 * 24,
		},
	},
});

const persister = createSyncStoragePersister({
	storage: localStorage,
});

createRoot(document.getElementById('root')!).render(
	<PersistQueryClientProvider client={queryClient} persistOptions={{ persister }}>
		<StrictMode>
			<App />
		</StrictMode>
		,
	</PersistQueryClientProvider>
);

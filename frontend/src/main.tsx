import './index.css';
import ReactDOM from 'react-dom';
import { StrictMode } from 'react';
import App from './app/App.tsx';

import FileProvider from './components/providers/fileprovider/index.tsx';
import { QueryClient, QueryClientProvider } from 'react-query';

ReactDOM.render(
	<QueryClientProvider client={new QueryClient()}>
		<StrictMode>
			<FileProvider>
				<App />
			</FileProvider>
		</StrictMode>
	</QueryClientProvider>,

	document.getElementById('root')
);

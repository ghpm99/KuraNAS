import { BrowserRouter } from 'react-router-dom';
import ActivityDiaryProvider from '../hooks/ActivityDiaryProvider';
import FileProvider from '../hooks/fileProvider';
import { UIProvider } from '../hooks/UI';
import Layout from '../layout/Layout/Layout';
import ActivePageListener from '../activePageListener';

const AppProviders = ({ children }: { children: React.ReactNode }) => {
	return (
		<UIProvider>
			<FileProvider>
				<ActivityDiaryProvider>
					<BrowserRouter>
						<Layout>
							<ActivePageListener />
							{children}
						</Layout>
					</BrowserRouter>
				</ActivityDiaryProvider>
			</FileProvider>
		</UIProvider>
	);
};

export default AppProviders;

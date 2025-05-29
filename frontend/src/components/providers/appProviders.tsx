import { BrowserRouter } from 'react-router-dom';
import ActivePageListener from '../activePageListener';
import { AboutProvider } from '../hooks/AboutProvider';
import ActivityDiaryProvider from '../hooks/ActivityDiaryProvider';
import FileProvider from '../hooks/fileProvider';
import { UIProvider } from '../hooks/UI';
import Layout from '../layout/Layout/Layout';

const AppProviders = ({ children }: { children: React.ReactNode }) => {
	return (
		<UIProvider>
			<FileProvider>
				<ActivityDiaryProvider>
					<AboutProvider>
						<BrowserRouter>
							<Layout>
								<ActivePageListener />
								{children}
							</Layout>
						</BrowserRouter>
					</AboutProvider>
				</ActivityDiaryProvider>
			</FileProvider>
		</UIProvider>
	);
};

export default AppProviders;

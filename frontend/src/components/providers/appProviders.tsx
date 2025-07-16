import { BrowserRouter } from 'react-router-dom';
import ActivePageListener from '../activePageListener';
import { AboutProvider } from '../hooks/AboutProvider';
import ActivityDiaryProvider from '../hooks/ActivityDiaryProvider';
import FileProvider from '../hooks/fileProvider';
import { UIProvider } from '../hooks/UI';
import Layout from '../layout/Layout/Layout';
import { AnalyticsProvider } from '../contexts/AnalyticsContext';

const AppProviders = ({ children }: { children: React.ReactNode }) => {
	return (
		<UIProvider>
			<FileProvider>
				<ActivityDiaryProvider>
					<AboutProvider>
						<AnalyticsProvider>
							<BrowserRouter>
								<Layout>
									<ActivePageListener />
									{children}
								</Layout>
							</BrowserRouter>
						</AnalyticsProvider>
					</AboutProvider>
				</ActivityDiaryProvider>
			</FileProvider>
		</UIProvider>
	);
};

export default AppProviders;

import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import { BrowserRouter } from 'react-router-dom';
import ActivePageListener from '../activePageListener';
import { AnalyticsProvider } from '../contexts/AnalyticsContext';
import { AboutProvider } from '../hooks/AboutProvider';
import ActivityDiaryProvider from '../hooks/ActivityDiaryProvider';
import FileProvider from '../hooks/fileProvider';
import { UIProvider } from '../hooks/UI';
import Layout from '../layout/Layout/Layout';

const darkTheme = createTheme({
	palette: {
		mode: 'dark',
	},
});

const AppProviders = ({ children }: { children: React.ReactNode }) => {
	return (
		<ThemeProvider theme={darkTheme}>
			<CssBaseline />
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
		</ThemeProvider>
	);
};

export default AppProviders;

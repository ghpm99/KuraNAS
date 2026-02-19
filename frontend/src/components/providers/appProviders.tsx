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
import { SnackbarProvider } from 'notistack';
import { ImageProvider } from '../hooks/imageProvider/imageProvider';
import { MusicProvider } from '../hooks/musicProvider/musicProvider';

const darkTheme = createTheme({
	palette: {
		mode: 'dark',
	},
});

const AppProviders = ({ children }: { children: React.ReactNode }) => {
	return (
		<SnackbarProvider maxSnack={3} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
			<ThemeProvider theme={darkTheme}>
				<CssBaseline />
				<UIProvider>
					<FileProvider>
						<ImageProvider>
							<MusicProvider>
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
							</MusicProvider>
						</ImageProvider>
					</FileProvider>
				</UIProvider>
			</ThemeProvider>
		</SnackbarProvider>
	);
};

export default AppProviders;

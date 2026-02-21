import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import { SnackbarProvider } from 'notistack';
import { BrowserRouter } from 'react-router-dom';
import ActivePageListener from '../activePageListener';
import { AnalyticsProvider } from '../contexts/AnalyticsContext';
import { AboutProvider } from '../hooks/AboutProvider';
import ActivityDiaryProvider from '../hooks/ActivityDiaryProvider';
import FileProvider from '../hooks/fileProvider';
import { ImageProvider } from '../hooks/imageProvider/imageProvider';
import { UIProvider } from '../hooks/UI';
import { VideoPlayerProvider } from '../hooks/videoPlayerProvider/videoPlayerProvider';
import Layout from '../layout/Layout/Layout';

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
							<ActivityDiaryProvider>
								<AboutProvider>
									<AnalyticsProvider>
										<BrowserRouter>
											<VideoPlayerProvider>
												<Layout>
													<ActivePageListener />
													{children}
												</Layout>
											</VideoPlayerProvider>
										</BrowserRouter>
									</AnalyticsProvider>
								</AboutProvider>
							</ActivityDiaryProvider>
						</ImageProvider>
					</FileProvider>
				</UIProvider>
			</ThemeProvider>
		</SnackbarProvider>
	);
};

export default AppProviders;

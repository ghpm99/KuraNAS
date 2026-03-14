import AppProviders from '@/components/providers/appProviders';

import { appRoutes, isVideoPlayerRoute } from '@/app/routes';
import ActivityDiaryPage from '@/pages/activityDiary';
import AnalyticsPage from '@/pages/analytics';
import FavoritesPage from '@/pages/favorites';
import FilePage from '@/pages/files';
import HomePage from '@/pages/home';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import AboutPage from '@/pages/about';
import ImagesPage from '@/pages/images';
import MusicPage from '@/pages/music';
import SettingsPage from '@/pages/settings';
import VideosPage from '@/pages/videos/videos';
import VideoPlayerPage from '@/pages/videoPlayer/videoPlayer';
import { GlobalMusicProvider } from '@/components/providers/GlobalMusicProvider';
import GlobalPlayerControl from '@/components/player/GlobalPlayerControl';
import ErrorBoundary from '@/components/ErrorBoundary';

function AppContent() {
	const location = useLocation();
	const hidePlayer = isVideoPlayerRoute(location.pathname);

	return (
		<>
			<Routes>
				<Route path={appRoutes.root} element={<Navigate to={appRoutes.home} replace />} />
				<Route path={appRoutes.home} element={<HomePage />} />
				<Route path={appRoutes.files} element={<FilePage />} />
				<Route path={appRoutes.favorites} element={<FavoritesPage />} />
				<Route path={appRoutes.legacyFavorites} element={<Navigate to={appRoutes.favorites} replace />} />
				<Route path={appRoutes.settings} element={<SettingsPage />} />
				<Route path={appRoutes.activityDiary} element={<ActivityDiaryPage />} />
				<Route path={appRoutes.analytics} element={<AnalyticsPage />} />
				<Route path={appRoutes.about} element={<AboutPage />} />
				<Route path={appRoutes.images} element={<ImagesPage />} />
				<Route path={appRoutes.music} element={<MusicPage />} />
				<Route path={appRoutes.videos} element={<VideosPage />} />
				<Route path={`${appRoutes.videoPlayerBase}/:id`} element={<VideoPlayerPage />} />
				<Route path='*' element={<Navigate to={appRoutes.home} replace />} />
			</Routes>
			{!hidePlayer && <GlobalPlayerControl />}
		</>
	);
}

export default function App() {
	return (
		<AppProviders>
			<ErrorBoundary>
				<GlobalMusicProvider>
					<AppContent />
				</GlobalMusicProvider>
			</ErrorBoundary>
		</AppProviders>
	);
}

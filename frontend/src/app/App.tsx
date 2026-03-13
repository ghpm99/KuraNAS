import AppProviders from '@/components/providers/appProviders';
import './App.css';

import ActivityDiaryPage from '@/pages/activityDiary';
import AnalyticsPage from '@/pages/analytics';
import FilePage from '@/pages/files';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import AboutPage from '@/pages/about';
import ImagesPage from '@/pages/images';
import MusicPage from '@/pages/music';
import VideosPage from '@/pages/videos/videos';
import VideoPlayerPage from '@/pages/videoPlayer/videoPlayer';
import { GlobalMusicProvider } from '@/components/providers/GlobalMusicProvider';
import GlobalPlayerControl from '@/components/player/GlobalPlayerControl';
import ErrorBoundary from '@/components/ErrorBoundary';

function AppContent() {
	const location = useLocation();
	const hidePlayer = location.pathname.startsWith('/video/');

	return (
		<>
			<Routes>
				<Route path='/' element={<FilePage />} />
				<Route path='/starred' element={<FilePage />} />
				<Route path='/activity-diary' element={<ActivityDiaryPage />} />
				<Route path='/analytics' element={<AnalyticsPage />} />
				<Route path='/about' element={<AboutPage />} />
				<Route path='/images' element={<ImagesPage />} />
				<Route path='/music' element={<MusicPage />} />
				<Route path='/videos' element={<VideosPage />} />
				<Route path='/video/:id' element={<VideoPlayerPage />} />
				<Route path='*' element={<Navigate to='/' replace />} />
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

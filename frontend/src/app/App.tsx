import AppProviders from '@/components/providers/appProviders';

import { appRoutes, getMusicRoute, isVideoPlayerRoute } from '@/app/routes';
import ActivityDiaryPage from '@/pages/activityDiary';
import AnalyticsPage from '@/pages/analytics';
import FavoritesPage from '@/pages/favorites';
import FilePage from '@/pages/files';
import HomePage from '@/pages/home';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import AboutPage from '@/pages/about';
import ImagesPage from '@/pages/images';
import MusicPage from '@/pages/music';
import AlbumsView from '@/pages/music/views/AlbumsView';
import ArtistsView from '@/pages/music/views/ArtistsView';
import FoldersView from '@/pages/music/views/FoldersView';
import GenresView from '@/pages/music/views/GenresView';
import PlaylistsView from '@/pages/music/views/PlaylistsView';
import MusicHomeScreen from '@/components/music/MusicHomeScreen';
import NotificationsPage from '@/pages/notifications';
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
                <Route path={`${appRoutes.files}/*`} element={<FilePage />} />
                <Route path={appRoutes.favorites} element={<FavoritesPage />} />
                <Route
                    path={appRoutes.legacyFavorites}
                    element={<Navigate to={appRoutes.favorites} replace />}
                />
                <Route path={appRoutes.settings} element={<SettingsPage />} />
                <Route path={appRoutes.activityDiary} element={<ActivityDiaryPage />} />
                <Route
                    path={appRoutes.legacyActivityDiary}
                    element={<Navigate to={appRoutes.activityDiary} replace />}
                />
                <Route path={`${appRoutes.analytics}/*`} element={<AnalyticsPage />} />
                <Route path={appRoutes.about} element={<AboutPage />} />
                <Route path={appRoutes.notifications} element={<NotificationsPage />} />
                <Route path={`${appRoutes.images}/*`} element={<ImagesPage />} />
                <Route path={`${appRoutes.music}/*`} element={<MusicPage />}>
                    <Route index element={<MusicHomeScreen />} />
                    <Route path="playlists" element={<PlaylistsView />} />
                    <Route path="artists" element={<ArtistsView />} />
                    <Route path="albums" element={<AlbumsView />} />
                    <Route path="genres" element={<GenresView />} />
                    <Route path="folders" element={<FoldersView />} />
                    <Route path="*" element={<Navigate to={getMusicRoute('home')} replace />} />
                </Route>
                <Route path={`${appRoutes.videos}/*`} element={<VideosPage />} />
                <Route path={`${appRoutes.videoPlayerBase}/:id`} element={<VideoPlayerPage />} />
                <Route path="*" element={<Navigate to={appRoutes.home} replace />} />
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

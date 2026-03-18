import { lazy, Suspense } from 'react';
import AppProviders from '@/components/providers/appProviders';

import { appRoutes, getMusicRoute, isVideoPlayerRoute } from '@/app/routes';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { GlobalMusicProvider } from '@/components/providers/GlobalMusicProvider';
import GlobalPlayerControl from '@/components/player/GlobalPlayerControl';
import ErrorBoundary from '@/components/ErrorBoundary';

const HomePage = lazy(() => import('@/pages/home'));
const FilePage = lazy(() => import('@/pages/files'));
const FavoritesPage = lazy(() => import('@/pages/favorites'));
const SettingsPage = lazy(() => import('@/pages/settings'));
const ActivityDiaryPage = lazy(() => import('@/pages/activityDiary'));
const AnalyticsPage = lazy(() => import('@/pages/analytics'));
const AboutPage = lazy(() => import('@/pages/about'));
const NotificationsPage = lazy(() => import('@/pages/notifications'));
const ImagesPage = lazy(() => import('@/pages/images'));
const MusicPage = lazy(() => import('@/pages/music'));
const MusicHomeScreen = lazy(() => import('@/components/music/MusicHomeScreen'));
const AlbumsView = lazy(() => import('@/pages/music/views/AlbumsView'));
const ArtistsView = lazy(() => import('@/pages/music/views/ArtistsView'));
const FoldersView = lazy(() => import('@/pages/music/views/FoldersView'));
const GenresView = lazy(() => import('@/pages/music/views/GenresView'));
const PlaylistsView = lazy(() => import('@/pages/music/views/PlaylistsView'));
const VideosPage = lazy(() => import('@/pages/videos/videos'));
const VideoPlayerPage = lazy(() => import('@/pages/videoPlayer/videoPlayer'));

function AppContent() {
    const location = useLocation();
    const hidePlayer = isVideoPlayerRoute(location.pathname);

    return (
        <Suspense>
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
        </Suspense>
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

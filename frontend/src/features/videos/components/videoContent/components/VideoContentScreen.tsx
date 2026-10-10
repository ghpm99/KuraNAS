import { CircularProgress, Typography } from '@mui/material';
import ErrorState from '@/components/errorState/errorState';
import type { VideoPlaylistDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoContentProvider } from '@/features/videos/providers/videoContentProvider';
import VideoFeedbackSnackbar from './VideoFeedbackSnackbar';
import VideoContextDetailView from './VideoContextDetailView';
import VideoContinueWatchingSection from './VideoContinueWatchingSection';
import VideoHomeScreen from './VideoHomeScreen';
import VideoLibrarySection from './VideoLibrarySection';
import VideoPlaylistDetailView from './VideoPlaylistDetailView';
import VideoSeriesDetailView from './VideoSeriesDetailView';
import VideoSectionPlaylistGrid from './VideoSectionPlaylistGrid';
import VideoSectionError from './VideoSectionError';
import VideoSectionSkeleton from './VideoSectionSkeleton';
import styles from '../videoContent.module.css';

export default function VideoContentScreen() {
    const { t } = useI18n();
    const {
        currentSection,
        selectedPlaylistDetail,
        selectedPlaylistSummary,
        isLoadingPlaylists,
        isLoadingVideos,
        isLoadingSelectedPlaylist,
        isLoadingHomeCatalog,
        isLoadingContinueWatching,
        playlistsFailure,
        videosFailure,
        selectedPlaylistFailure,
        homeCatalogFailure,
        continueWatchingFailure,
        isFetchingMoreVideos,
        hasMoreVideos,
        isAddingToPlaylist,
        isRenamingPlaylist,
        isRemovingFromPlaylist,
        isReorderingPlaylist,
        continueWatchingItems,
        seriesPlaylists,
        moviePlaylists,
        personalPlaylists,
        clipPlaylists,
        folderPlaylists,
        recentCatalogItems,
        filteredVideos,
        playlists,
        playlistMembershipMap,
        videoSearch,
        selectedPlaylistPerVideo,
        feedback,
        setVideoSearch,
        setSelectedPlaylistForVideo,
        closeFeedback,
        loadMoreVideos,
        selectPlaylist,
        clearSelectedPlaylist,
        playVideo,
        openPlaylistVideo,
        addVideoFromLibrary,
        renameSelectedPlaylist,
        removeVideoFromSelectedPlaylist,
        moveSelectedPlaylistItem,
        setVideoWatched,
    } = useVideoContentProvider();

    if (selectedPlaylistSummary) {
        if (selectedPlaylistFailure) {
            return (
                <div className={styles.page}>
                    <ErrorState
                        title={t('VIDEO_SECTION_LOAD_ERROR', {
                            section: selectedPlaylistSummary.name,
                        })}
                        backendMessage={selectedPlaylistFailure.message}
                        onRetry={selectedPlaylistFailure.retry}
                    />
                </div>
            );
        }
        if (isLoadingSelectedPlaylist || !selectedPlaylistDetail) {
            return (
                <div className={styles.loadingState}>
                    <CircularProgress size={40} />
                    <Typography variant="h6">{t('VIDEO_LOADING_PLAYLIST')}</Typography>
                </div>
            );
        }

        if (
            selectedPlaylistDetail.classification === 'series' ||
            selectedPlaylistDetail.classification === 'anime'
        ) {
            return (
                <VideoSeriesDetailView
                    playlist={selectedPlaylistDetail}
                    onBack={clearSelectedPlaylist}
                    onOpenVideo={openPlaylistVideo}
                    onSetWatched={setVideoWatched}
                />
            );
        }

        if (currentSection !== 'folders') {
            return (
                <VideoContextDetailView
                    playlist={selectedPlaylistDetail}
                    onBack={clearSelectedPlaylist}
                    onOpenVideo={openPlaylistVideo}
                    onSetWatched={setVideoWatched}
                />
            );
        }

        return (
            <VideoPlaylistDetailView
                playlist={selectedPlaylistDetail}
                isRenaming={isRenamingPlaylist}
                isRemoving={isRemovingFromPlaylist}
                isReordering={isReorderingPlaylist}
                onBack={clearSelectedPlaylist}
                onOpenVideo={openPlaylistVideo}
                onRename={renameSelectedPlaylist}
                onRemoveVideo={removeVideoFromSelectedPlaylist}
                onMoveItem={moveSelectedPlaylistItem}
                onSetWatched={setVideoWatched}
            />
        );
    }

    const renderPlaylistGridSection = (
        sectionName: string,
        playlistsOfSection: VideoPlaylistDto[]
    ) => {
        if (isLoadingPlaylists) {
            return <VideoSectionSkeleton />;
        }
        if (playlistsFailure) {
            return (
                <VideoSectionError
                    sectionTitleKey={`VIDEO_SECTION_${sectionName}`}
                    failure={playlistsFailure}
                />
            );
        }
        return (
            <VideoSectionPlaylistGrid
                titleKey={`VIDEO_SECTION_${sectionName}`}
                descriptionKey={`VIDEO_SECTION_${sectionName}_DESCRIPTION`}
                emptyKey={`VIDEO_SECTION_${sectionName}_EMPTY`}
                playlists={playlistsOfSection}
                onSelectPlaylist={selectPlaylist}
                onPlayVideo={playVideo}
            />
        );
    };

    const renderVideoLibrarySection = () => {
        if (isLoadingVideos) {
            return <VideoSectionSkeleton cardCount={6} layout="catalogRail" />;
        }
        if (videosFailure) {
            return <VideoSectionError sectionTitleKey="VIDEO_ALL" failure={videosFailure} />;
        }
        return (
            <VideoLibrarySection
                videos={filteredVideos}
                playlists={playlists}
                playlistMembershipMap={playlistMembershipMap}
                search={videoSearch}
                selectedPlaylistPerVideo={selectedPlaylistPerVideo}
                isAddingToPlaylist={isAddingToPlaylist}
                isFetchingMoreVideos={isFetchingMoreVideos}
                hasMoreVideos={hasMoreVideos}
                onSearchChange={setVideoSearch}
                onSelectPlaylistForVideo={setSelectedPlaylistForVideo}
                onPlayVideo={playVideo}
                onAddVideo={addVideoFromLibrary}
                onLoadMore={loadMoreVideos}
                onSetWatched={setVideoWatched}
            />
        );
    };

    const renderSectionContent = () => {
        switch (currentSection) {
            case 'continue':
                if (isLoadingContinueWatching) {
                    return <VideoSectionSkeleton layout="catalogRail" />;
                }
                if (continueWatchingFailure) {
                    return (
                        <VideoSectionError
                            sectionTitleKey="VIDEO_SECTION_CONTINUE"
                            failure={continueWatchingFailure}
                        />
                    );
                }
                return (
                    <VideoContinueWatchingSection
                        items={continueWatchingItems}
                        onPlayVideo={playVideo}
                        onSetWatched={setVideoWatched}
                    />
                );
            case 'series':
                return renderPlaylistGridSection('SERIES', seriesPlaylists);
            case 'movies':
                return renderPlaylistGridSection('MOVIES', moviePlaylists);
            case 'personal':
                return renderPlaylistGridSection('PERSONAL', personalPlaylists);
            case 'clips':
                return renderPlaylistGridSection('CLIPS', clipPlaylists);
            case 'folders':
                return (
                    <>
                        {renderPlaylistGridSection('FOLDERS', folderPlaylists)}
                        {renderVideoLibrarySection()}
                    </>
                );
            case 'home':
            default:
                return (
                    <VideoHomeScreen
                        continueWatchingItems={continueWatchingItems}
                        seriesPlaylists={seriesPlaylists}
                        moviePlaylists={moviePlaylists}
                        personalPlaylists={personalPlaylists}
                        clipPlaylists={clipPlaylists}
                        folderPlaylists={folderPlaylists}
                        recentCatalogItems={recentCatalogItems}
                        isLoadingContinueWatching={isLoadingContinueWatching}
                        isLoadingPlaylists={isLoadingPlaylists}
                        isLoadingHomeCatalog={isLoadingHomeCatalog}
                        playlistsFailure={playlistsFailure}
                        continueWatchingFailure={continueWatchingFailure}
                        homeCatalogFailure={homeCatalogFailure}
                        onSelectPlaylist={selectPlaylist}
                        onPlayVideo={playVideo}
                        onSetWatched={setVideoWatched}
                    />
                );
        }
    };

    return (
        <div className={styles.page}>
            {renderSectionContent()}
            <VideoFeedbackSnackbar
                open={feedback.open}
                message={feedback.message}
                severity={feedback.severity}
                onClose={closeFeedback}
            />
        </div>
    );
}

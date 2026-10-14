import { CircularProgress, Typography } from '@mui/material';
import type { ReactNode } from 'react';
import ErrorState from '@/components/errorState/errorState';
import type { VideoPlaylistDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoContentProvider } from '@/features/videos/providers/videoContentProvider';
import VideoFolderBrowser from '../folderBrowser/VideoFolderBrowser';
import VideoMovieBrowser from '../movieBrowser/VideoMovieBrowser';
import VideoFeedbackSnackbar from './VideoFeedbackSnackbar';
import VideoContextDetailView from './VideoContextDetailView';
import VideoContinueWatchingSection from './VideoContinueWatchingSection';
import VideoHiddenPlaylistsLink from './VideoHiddenPlaylistsLink';
import VideoHiddenPlaylistsSection from './VideoHiddenPlaylistsSection';
import VideoNewPlaylistButton from './VideoNewPlaylistButton';
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
        isFetchingMoreSelectedPlaylistItems,
        hasMoreSelectedPlaylistItems,
        isLoadingHomeCatalog,
        isLoadingContinueWatching,
        playlistsFailure,
        videosFailure,
        selectedPlaylistFailure,
        homeCatalogFailure,
        continueWatchingFailure,
        isFetchingMoreVideos,
        hasMoreVideos,
        isFetchingMoreSectionPlaylists,
        hasMoreSectionPlaylists,
        isFetchingMoreHiddenPlaylists,
        hasMoreHiddenPlaylists,
        isChangingPlaylistVisibility,
        isAddingToPlaylist,
        isRenamingPlaylist,
        isRemovingFromPlaylist,
        isReorderingPlaylist,
        isCreatingPlaylist,
        isDeletingPlaylist,
        continueWatchingItems,
        seriesPlaylists,
        moviePlaylists,
        personalPlaylists,
        clipPlaylists,
        folderPlaylists,
        hiddenPlaylists,
        recentCatalogItems,
        filteredVideos,
        videoSearch,
        librarySort,
        libraryFilters,
        feedback,
        setVideoSearch,
        setLibrarySort,
        setLibraryFilters,
        closeFeedback,
        loadMoreVideos,
        loadMoreSectionPlaylists,
        loadMoreHiddenPlaylists,
        loadMoreSelectedPlaylistItems,
        selectPlaylist,
        clearSelectedPlaylist,
        playVideo,
        openPlaylistVideo,
        addVideoFromLibrary,
        createPlaylist,
        createPlaylistWithVideo,
        deleteSelectedPlaylist,
        hidePlaylist,
        hideSelectedPlaylist,
        showPlaylist,
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
                    isChangingVisibility={isChangingPlaylistVisibility}
                    onHide={hideSelectedPlaylist}
                    onBack={clearSelectedPlaylist}
                    onOpenVideo={openPlaylistVideo}
                    onSetWatched={setVideoWatched}
                    hasMoreItems={hasMoreSelectedPlaylistItems}
                    isFetchingMoreItems={isFetchingMoreSelectedPlaylistItems}
                    onLoadMoreItems={loadMoreSelectedPlaylistItems}
                />
            );
        }

        if (currentSection !== 'folders') {
            return (
                <VideoContextDetailView
                    playlist={selectedPlaylistDetail}
                    isDeleting={isDeletingPlaylist}
                    onDelete={deleteSelectedPlaylist}
                    isChangingVisibility={isChangingPlaylistVisibility}
                    onHide={hideSelectedPlaylist}
                    onBack={clearSelectedPlaylist}
                    onOpenVideo={openPlaylistVideo}
                    onSetWatched={setVideoWatched}
                    hasMoreItems={hasMoreSelectedPlaylistItems}
                    isFetchingMoreItems={isFetchingMoreSelectedPlaylistItems}
                    onLoadMoreItems={loadMoreSelectedPlaylistItems}
                />
            );
        }

        return (
            <VideoPlaylistDetailView
                playlist={selectedPlaylistDetail}
                isRenaming={isRenamingPlaylist}
                isRemoving={isRemovingFromPlaylist}
                isReordering={isReorderingPlaylist}
                isDeleting={isDeletingPlaylist}
                onDelete={deleteSelectedPlaylist}
                isChangingVisibility={isChangingPlaylistVisibility}
                onHide={hideSelectedPlaylist}
                onBack={clearSelectedPlaylist}
                onOpenVideo={openPlaylistVideo}
                onRename={renameSelectedPlaylist}
                onRemoveVideo={removeVideoFromSelectedPlaylist}
                onMoveItem={moveSelectedPlaylistItem}
                onSetWatched={setVideoWatched}
                hasMoreItems={hasMoreSelectedPlaylistItems}
                isFetchingMoreItems={isFetchingMoreSelectedPlaylistItems}
                onLoadMoreItems={loadMoreSelectedPlaylistItems}
            />
        );
    }

    const renderPlaylistGridSection = (
        sectionName: string,
        playlistsOfSection: VideoPlaylistDto[],
        sectionAction?: ReactNode
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
                action={sectionAction}
                hasNextPage={hasMoreSectionPlaylists}
                isFetchingNextPage={isFetchingMoreSectionPlaylists}
                fetchNextPage={loadMoreSectionPlaylists}
                onSelectPlaylist={selectPlaylist}
                onPlayVideo={playVideo}
                onHidePlaylist={hidePlaylist}
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
                search={videoSearch}
                sort={librarySort}
                filters={libraryFilters}
                isAddingToPlaylist={isAddingToPlaylist}
                isCreatingPlaylist={isCreatingPlaylist}
                isFetchingMoreVideos={isFetchingMoreVideos}
                hasMoreVideos={hasMoreVideos}
                onSearchChange={setVideoSearch}
                onSortChange={setLibrarySort}
                onFiltersChange={setLibraryFilters}
                onPlayVideo={playVideo}
                onAddVideo={addVideoFromLibrary}
                onCreatePlaylistWithVideo={createPlaylistWithVideo}
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
                return <VideoMovieBrowser onPlayVideo={playVideo} />;
            case 'personal':
                return renderPlaylistGridSection(
                    'PERSONAL',
                    personalPlaylists,
                    <VideoNewPlaylistButton
                        isCreating={isCreatingPlaylist}
                        onCreate={createPlaylist}
                    />
                );
            case 'clips':
                return renderPlaylistGridSection('CLIPS', clipPlaylists);
            case 'folders':
                return (
                    <>
                        <VideoFolderBrowser onPlayVideo={playVideo} />
                        {renderVideoLibrarySection()}
                    </>
                );
            case 'hidden':
                if (isLoadingPlaylists) {
                    return <VideoSectionSkeleton />;
                }
                if (playlistsFailure) {
                    return (
                        <VideoSectionError
                            sectionTitleKey="VIDEO_SECTION_HIDDEN"
                            failure={playlistsFailure}
                        />
                    );
                }
                return (
                    <VideoHiddenPlaylistsSection
                        playlists={hiddenPlaylists}
                        isChangingVisibility={isChangingPlaylistVisibility}
                        hasNextPage={hasMoreHiddenPlaylists}
                        isFetchingNextPage={isFetchingMoreHiddenPlaylists}
                        fetchNextPage={loadMoreHiddenPlaylists}
                        onShowPlaylist={showPlaylist}
                    />
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
                        onHidePlaylist={hidePlaylist}
                        onSetWatched={setVideoWatched}
                    />
                );
        }
    };

    const canLinkToHiddenPlaylists = currentSection !== 'hidden' && currentSection !== 'continue';

    return (
        <div className={styles.page}>
            {renderSectionContent()}
            {canLinkToHiddenPlaylists && <VideoHiddenPlaylistsLink />}
            <VideoFeedbackSnackbar
                open={feedback.open}
                message={feedback.message}
                severity={feedback.severity}
                action={feedback.action}
                onClose={closeFeedback}
            />
        </div>
    );
}

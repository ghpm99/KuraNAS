import { CircularProgress, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoContentProvider } from '@/components/providers/videoContentProvider';
import VideoFeedbackSnackbar from './VideoFeedbackSnackbar';
import VideoContextDetailView from './VideoContextDetailView';
import VideoHomeScreen from './VideoHomeScreen';
import VideoLibrarySection from './VideoLibrarySection';
import VideoPlaylistDetailView from './VideoPlaylistDetailView';
import VideoSeriesDetailView from './VideoSeriesDetailView';
import VideoSectionPlaylistGrid from './VideoSectionPlaylistGrid';
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
		isFetchingMoreVideos,
		hasMoreVideos,
		isAddingToPlaylist,
		isRenamingPlaylist,
		isRemovingFromPlaylist,
		isReorderingPlaylist,
		continuePlaylists,
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
	} = useVideoContentProvider();

	if (isLoadingPlaylists || isLoadingVideos || isLoadingHomeCatalog) {
		return (
			<div className={styles.loadingState}>
				<CircularProgress size={44} />
				<Typography variant='h6'>{t('VIDEO_LOADING_VIDEOS')}</Typography>
			</div>
		);
	}

	if (selectedPlaylistSummary) {
		if (isLoadingSelectedPlaylist || !selectedPlaylistDetail) {
			return (
				<div className={styles.loadingState}>
					<CircularProgress size={40} />
					<Typography variant='h6'>{t('VIDEO_LOADING_PLAYLIST')}</Typography>
				</div>
			);
		}

		if (selectedPlaylistDetail.classification === 'series' || selectedPlaylistDetail.classification === 'anime') {
			return (
				<VideoSeriesDetailView
					playlist={selectedPlaylistDetail}
					onBack={clearSelectedPlaylist}
					onOpenVideo={openPlaylistVideo}
				/>
			);
		}

		if (currentSection !== 'folders') {
			return (
				<VideoContextDetailView
					playlist={selectedPlaylistDetail}
					onBack={clearSelectedPlaylist}
					onOpenVideo={openPlaylistVideo}
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
			/>
		);
	}

	const renderSectionContent = () => {
		switch (currentSection) {
			case 'continue':
				return (
					<VideoSectionPlaylistGrid
						titleKey='VIDEO_SECTION_CONTINUE'
						descriptionKey='VIDEO_SECTION_CONTINUE_DESCRIPTION'
						emptyKey='VIDEO_NO_RECENT_PLAYLISTS'
						playlists={continuePlaylists}
						onSelectPlaylist={selectPlaylist}
						onPlayVideo={playVideo}
						badge={t('VIDEO_CONTINUE_BADGE_RESUME')}
					/>
				);
			case 'series':
				return (
					<VideoSectionPlaylistGrid
						titleKey='VIDEO_SECTION_SERIES'
						descriptionKey='VIDEO_SECTION_SERIES_DESCRIPTION'
						emptyKey='VIDEO_SECTION_SERIES_EMPTY'
						playlists={seriesPlaylists}
						onSelectPlaylist={selectPlaylist}
						onPlayVideo={playVideo}
					/>
				);
			case 'movies':
				return (
					<VideoSectionPlaylistGrid
						titleKey='VIDEO_SECTION_MOVIES'
						descriptionKey='VIDEO_SECTION_MOVIES_DESCRIPTION'
						emptyKey='VIDEO_SECTION_MOVIES_EMPTY'
						playlists={moviePlaylists}
						onSelectPlaylist={selectPlaylist}
						onPlayVideo={playVideo}
					/>
				);
			case 'personal':
				return (
					<VideoSectionPlaylistGrid
						titleKey='VIDEO_SECTION_PERSONAL'
						descriptionKey='VIDEO_SECTION_PERSONAL_DESCRIPTION'
						emptyKey='VIDEO_SECTION_PERSONAL_EMPTY'
						playlists={personalPlaylists}
						onSelectPlaylist={selectPlaylist}
						onPlayVideo={playVideo}
					/>
				);
			case 'clips':
				return (
					<VideoSectionPlaylistGrid
						titleKey='VIDEO_SECTION_CLIPS'
						descriptionKey='VIDEO_SECTION_CLIPS_DESCRIPTION'
						emptyKey='VIDEO_SECTION_CLIPS_EMPTY'
						playlists={clipPlaylists}
						onSelectPlaylist={selectPlaylist}
						onPlayVideo={playVideo}
					/>
				);
			case 'folders':
				return (
					<>
						<VideoSectionPlaylistGrid
							titleKey='VIDEO_SECTION_FOLDERS'
							descriptionKey='VIDEO_SECTION_FOLDERS_DESCRIPTION'
							emptyKey='VIDEO_SECTION_FOLDERS_EMPTY'
							playlists={folderPlaylists}
							onSelectPlaylist={selectPlaylist}
							onPlayVideo={playVideo}
						/>
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
						/>
					</>
				);
			case 'home':
			default:
				return (
					<VideoHomeScreen
						continuePlaylists={continuePlaylists}
						seriesPlaylists={seriesPlaylists}
						moviePlaylists={moviePlaylists}
						personalPlaylists={personalPlaylists}
						clipPlaylists={clipPlaylists}
						folderPlaylists={folderPlaylists}
						recentCatalogItems={recentCatalogItems}
						onSelectPlaylist={selectPlaylist}
						onPlayVideo={playVideo}
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

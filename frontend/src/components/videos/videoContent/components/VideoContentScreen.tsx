import { CircularProgress, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoContentProvider } from '@/components/providers/videoContentProvider';
import VideoCatalogSections from './VideoCatalogSections';
import VideoFeedbackSnackbar from './VideoFeedbackSnackbar';
import VideoLibrarySection from './VideoLibrarySection';
import VideoPlaylistDetailView from './VideoPlaylistDetailView';
import styles from '../videoContent.module.css';

export default function VideoContentScreen() {
	const { t } = useI18n();
	const {
		selectedPlaylistDetail,
		selectedPlaylistSummary,
		isLoadingPlaylists,
		isLoadingVideos,
		isLoadingSelectedPlaylist,
		isAddingToPlaylist,
		isRenamingPlaylist,
		isRemovingFromPlaylist,
		isReorderingPlaylist,
		continuePlaylists,
		groupedPlaylists,
		filteredVideos,
		playlists,
		playlistMembershipMap,
		videoSearch,
		selectedPlaylistPerVideo,
		feedback,
		setVideoSearch,
		setSelectedPlaylistForVideo,
		closeFeedback,
		selectPlaylist,
		clearSelectedPlaylist,
		playVideo,
		openPlaylistVideo,
		addVideoFromLibrary,
		renameSelectedPlaylist,
		removeVideoFromSelectedPlaylist,
		moveSelectedPlaylistItem,
	} = useVideoContentProvider();

	if (isLoadingPlaylists || isLoadingVideos) {
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

	return (
		<div className={styles.page}>
			<VideoCatalogSections
				continuePlaylists={continuePlaylists}
				groupedPlaylists={groupedPlaylists}
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
				onSearchChange={setVideoSearch}
				onSelectPlaylistForVideo={setSelectedPlaylistForVideo}
				onPlayVideo={playVideo}
				onAddVideo={addVideoFromLibrary}
			/>
			<VideoFeedbackSnackbar
				open={feedback.open}
				message={feedback.message}
				severity={feedback.severity}
				onClose={closeFeedback}
			/>
		</div>
	);
}

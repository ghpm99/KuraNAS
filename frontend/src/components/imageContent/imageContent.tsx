import { CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import ImageGroupsGrid from './components/ImageGroupsGrid';
import ImageToolbar from './components/ImageToolbar';
import ImageViewerModal from './components/ImageViewerModal';
import ImageCollectionsPanel from './components/ImageCollectionsPanel';
import { useImageContent } from './useImageContent';
import styles from './ImageContent.module.css';

export default function ImageContent() {
	const { t } = useI18n();
	const {
		activeImage,
		activeIndex,
		activeImageDate,
		activeSection,
		activeSelection,
		activeSelectionDescription,
		activeSelectionTitle,
		dateFormatter,
		emptyState,
		filteredAlbumCards,
		filteredFolderCards,
		filteredImages,
		goNext,
		goPrevious,
		groupByLabels,
		groupedImages,
		handleCloseViewer,
		handleOpenImage,
		handleOpenFolder,
		handleToggleFavorite,
		handleSelectAlbum,
		handleSelectFolder,
		hasNextPage,
		imageGroupBy,
		increaseZoom,
		isFavoritePending,
		isFetchingNextPage,
		isSlideshowPlaying,
		lastVisibleImageId,
		loadMoreRef,
		resetZoom,
		search,
		setImageGroupBy,
		setSearch,
		setShowDetails,
		setShowFilmstrip,
		showDetails,
		showFilmstrip,
		status,
		summary,
		title,
		toggleSlideshow,
		viewMode,
		zoom,
		decreaseZoom,
		selectedAlbum,
		selectedFolder,
	} = useImageContent();
	const isInitialGridLoading = status === 'pending' && filteredImages.length === 0 && viewMode === 'grid';

	return (
		<div className={styles.content}>
			<ImageToolbar
				title={title}
				summary={summary}
				search={search}
				groupBy={imageGroupBy}
				groupByLabels={groupByLabels}
				showGrouping={viewMode === 'grid'}
				onSearchChange={setSearch}
				onGroupByChange={setImageGroupBy}
			/>
			{activeSelection && (
				<div className={styles.selectionSummary}>
					<div className={styles.selectionMeta}>
						<h3>{activeSelectionTitle}</h3>
						<p>{activeSelectionDescription}</p>
					</div>
					<button
						type='button'
						className={styles.backButton}
						onClick={() => {
							if (activeSection === 'folders') {
								handleSelectFolder(null);
								return;
							}

							handleSelectAlbum(null);
						}}
					>
						{activeSection === 'folders' ? t('IMAGES_BACK_TO_FOLDERS') : t('IMAGES_BACK_TO_ALBUMS')}
					</button>
				</div>
			)}
			{isInitialGridLoading ? (
				<div className={styles.loading}>
					<CircularProgress size={40} />
				</div>
			) : null}
			{viewMode === 'folders' && !selectedFolder ? (
				<ImageCollectionsPanel
					cards={filteredFolderCards}
					emptyTitle={emptyState.title}
					emptyDescription={emptyState.description}
					onSelect={handleSelectFolder}
				/>
			) : null}
			{viewMode === 'albums' && !selectedAlbum ? (
				<ImageCollectionsPanel
					cards={filteredAlbumCards}
					emptyTitle={emptyState.title}
					emptyDescription={emptyState.description}
					onSelect={handleSelectAlbum}
				/>
			) : null}
			{viewMode === 'grid' && !isInitialGridLoading && (
				<ImageGroupsGrid
					groups={groupedImages}
					totalImages={filteredImages.length}
					isFetchingNextPage={isFetchingNextPage}
					hasNextPage={hasNextPage}
					lastVisibleImageId={lastVisibleImageId}
					loadMoreRef={loadMoreRef}
					onOpenImage={handleOpenImage}
				/>
			)}
			{activeImage && (
				<ImageViewerModal
					activeImage={activeImage}
					activeIndex={activeIndex}
					activeImageDate={activeImageDate}
					dateFormatter={dateFormatter}
					filteredImages={filteredImages}
					zoom={zoom}
					showDetails={showDetails}
					showFilmstrip={showFilmstrip}
					isSlideshowPlaying={isSlideshowPlaying}
					isFavoritePending={isFavoritePending}
					onToggleDetails={() => setShowDetails((value) => !value)}
					onToggleFilmstrip={() => setShowFilmstrip((value) => !value)}
					onToggleSlideshow={toggleSlideshow}
					onToggleFavorite={handleToggleFavorite}
					onOpenFolder={handleOpenFolder}
					onDecreaseZoom={decreaseZoom}
					onResetZoom={resetZoom}
					onIncreaseZoom={increaseZoom}
					onClose={handleCloseViewer}
					onPrevious={goPrevious}
					onNext={goNext}
					onOpenImage={handleOpenImage}
				/>
			)}
		</div>
	);
}

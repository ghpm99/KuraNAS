import { CircularProgress } from '@mui/material';
import EmptyState from '@/components/emptyState/emptyState';
import ErrorState from '@/components/errorState/errorState';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import ImageCollectionsPanel from './components/ImageCollectionsPanel';
import ImageDateScrubber from './components/ImageDateScrubber';
import ImageFilterBar from './components/ImageFilterBar';
import ImageGroupsGrid from './components/ImageGroupsGrid';
import ImageToolbar from './components/ImageToolbar';
import ImageViewerModal from './components/ImageViewerModal';
import { useImageContent, type ImageEmptyKind } from './useImageContent';
import styles from './ImageContent.module.css';

const emptyStateKeys: Record<ImageEmptyKind, { titleKey: string; descriptionKey: string }> = {
    library: { titleKey: 'IMAGES_EMPTY_TITLE', descriptionKey: 'IMAGES_EMPTY_DESC' },
    filtered: {
        titleKey: 'IMAGES_EMPTY_FILTERED_TITLE',
        descriptionKey: 'IMAGES_EMPTY_FILTERED_DESC',
    },
    favorites: {
        titleKey: 'IMAGES_EMPTY_FAVORITES_TITLE',
        descriptionKey: 'IMAGES_EMPTY_FAVORITES_DESC',
    },
};

export default function ImageContent() {
    const { t } = useI18n();
    const {
        view,
        viewMode,
        title,
        summary,
        groups,
        folderCards,
        albumCards,
        timeline,
        status,
        hasLoadError,
        loadErrorMessage,
        retryLoading,
        hasNextPage,
        isFetchingNextPage,
        loadNextPage,
        isEmpty,
        emptyKind,
        typedNameQuery,
        setTypedNameQuery,
        controls,
        dateFormatter,
        viewerImages,
        viewer,
        activeImageDate,
        isFavoritePending,
        toggleStar,
        handleOpenImage,
        handleCloseViewer,
        handleToggleFavoriteOfActiveImage,
        handleOpenActiveImageFolder,
    } = useImageContent();
    const { activeImage } = viewer;
    const isAlbumPicker = viewMode === 'albums';
    const isInitialLoading = !isAlbumPicker && status === 'pending';
    const hasSelection = Boolean(view.selectedFolder || view.selectedAlbum);
    const hasScrubber = viewMode === 'grid' && view.isKeyset && timeline.length > 0;
    const emptyKeys = emptyStateKeys[emptyKind];

    return (
        <div className={styles.content}>
            <ImageToolbar
                title={title}
                summary={summary}
                search={typedNameQuery}
                isSearchVisible={!isAlbumPicker}
                onSearchChange={setTypedNameQuery}
            />
            {!isAlbumPicker && (
                <ImageFilterBar
                    filters={view.filters}
                    ordering={view.ordering}
                    hasUserFilters={view.hasUserFilters}
                    onTakenFromChange={controls.setTakenFrom}
                    onTakenToChange={controls.setTakenTo}
                    onFormatToggle={controls.toggleFormat}
                    onSortChange={controls.setSort}
                    onSortOrderToggle={controls.toggleSortOrder}
                    onClearFilters={controls.clearUserFilters}
                />
            )}
            {hasSelection && (
                <div className={styles.selectionSummary}>
                    <button
                        type="button"
                        className={styles.backButton}
                        onClick={() =>
                            view.selectedFolder
                                ? controls.selectFolder(null)
                                : controls.selectAlbum(null)
                        }
                    >
                        {view.selectedFolder
                            ? t('IMAGES_BACK_TO_FOLDERS')
                            : t('IMAGES_BACK_TO_ALBUMS')}
                    </button>
                </div>
            )}
            {view.takenBefore && (
                <div>
                    <button
                        type="button"
                        className={styles.backButton}
                        onClick={controls.clearJump}
                    >
                        {t('IMAGES_JUMP_BACK_LATEST')}
                    </button>
                </div>
            )}
            {isAlbumPicker && (
                <ImageCollectionsPanel
                    cards={albumCards}
                    emptyTitle={t('IMAGES_ALBUMS_EMPTY_TITLE')}
                    emptyDescription={t('IMAGES_ALBUMS_EMPTY_DESC')}
                    onSelect={controls.selectAlbum}
                />
            )}
            {isInitialLoading && (
                <div className={styles.loading}>
                    <CircularProgress size={40} />
                </div>
            )}
            {viewMode === 'folders' && !isInitialLoading && !isEmpty && (
                <ImageCollectionsPanel
                    cards={folderCards}
                    emptyTitle={t('IMAGES_FOLDERS_EMPTY_TITLE')}
                    emptyDescription={t('IMAGES_FOLDERS_EMPTY_DESC')}
                    onSelect={controls.selectFolder}
                />
            )}
            {!isAlbumPicker && isEmpty && (
                <EmptyState
                    title={
                        viewMode === 'folders'
                            ? t('IMAGES_FOLDERS_EMPTY_TITLE')
                            : t(emptyKeys.titleKey)
                    }
                    description={
                        viewMode === 'folders'
                            ? t('IMAGES_FOLDERS_EMPTY_DESC')
                            : t(emptyKeys.descriptionKey)
                    }
                />
            )}
            {viewMode === 'grid' && groups.length > 0 && (
                <div className={styles.galleryLayout}>
                    {hasScrubber && (
                        <div className={styles.scrubberSlot}>
                            <ImageDateScrubber
                                buckets={timeline}
                                onSelectMonth={controls.jumpToMonth}
                            />
                        </div>
                    )}
                    <div className={styles.galleryMain}>
                        <ImageGroupsGrid
                            groups={groups}
                            onOpenImage={handleOpenImage}
                            onToggleStar={toggleStar}
                        />
                    </div>
                </div>
            )}
            {hasLoadError && (
                <ErrorState
                    title={t('IMAGES_ERROR_TITLE')}
                    backendMessage={loadErrorMessage}
                    onRetry={retryLoading}
                />
            )}
            {!isAlbumPicker && (
                <LoadMoreSentinel
                    hasNextPage={hasNextPage && !hasLoadError}
                    isFetchingNextPage={isFetchingNextPage}
                    fetchNextPage={loadNextPage}
                />
            )}
            {!hasNextPage && !isAlbumPicker && groups.length + folderCards.length > 0 && (
                <div className={styles.endMessage}>{t('IMAGES_END_MESSAGE')}</div>
            )}
            {activeImage && (
                <ImageViewerModal
                    activeImage={activeImage}
                    activeIndex={viewer.activeIndex}
                    activeImageDate={activeImageDate}
                    dateFormatter={dateFormatter}
                    filteredImages={viewerImages}
                    zoom={viewer.zoom}
                    showDetails={viewer.showDetails}
                    showFilmstrip={viewer.showFilmstrip}
                    isSlideshowPlaying={viewer.isSlideshowPlaying}
                    isFavoritePending={isFavoritePending}
                    onToggleDetails={() => viewer.setShowDetails((isShown) => !isShown)}
                    onToggleFilmstrip={() => viewer.setShowFilmstrip((isShown) => !isShown)}
                    onToggleSlideshow={viewer.toggleSlideshow}
                    onToggleFavorite={handleToggleFavoriteOfActiveImage}
                    onOpenFolder={handleOpenActiveImageFolder}
                    onDecreaseZoom={viewer.decreaseZoom}
                    onResetZoom={viewer.resetZoom}
                    onIncreaseZoom={viewer.increaseZoom}
                    onClose={handleCloseViewer}
                    onPrevious={viewer.goPrevious}
                    onNext={viewer.goNext}
                    onOpenImage={handleOpenImage}
                />
            )}
        </div>
    );
}

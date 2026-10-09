import {
    ChevronLeft,
    ChevronRight,
    Download,
    Expand,
    FolderOpen,
    Info,
    Minus,
    Pause,
    Play,
    Plus,
    RotateCw,
    Star,
    X,
} from 'lucide-react';
import { createPortal } from 'react-dom';
import ImageViewerDetailItem from './ImageViewerDetailItem';
import { useImageViewerGestures } from './useImageViewerGestures';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    downloadOriginalUrl,
    FILMSTRIP_THUMBNAIL_SIZE,
    thumbnailUrl,
    viewerImageUrl,
} from '../imageThumbnailSources';
import { useImageViewerModal } from './useImageViewerModal';
import styles from './ImageViewerModal.module.css';

type ImageViewerModalProps = {
    activeImage: ImageLibraryItem;
    activeIndex: number;
    activeImageDate: Date | null;
    dateFormatter: Intl.DateTimeFormat;
    filteredImages: ImageLibraryItem[];
    zoom: number;
    totalImages?: number | null;
    rotation?: number;
    pan?: { x: number; y: number };
    canGoPrevious?: boolean;
    canGoNext?: boolean;
    showDetails: boolean;
    showFilmstrip: boolean;
    isSlideshowPlaying: boolean;
    isFavoritePending: boolean;
    onToggleDetails: () => void;
    onToggleFilmstrip: () => void;
    onToggleSlideshow: () => void;
    onToggleFavorite: () => void;
    onOpenFolder: () => void;
    onDecreaseZoom: () => void;
    onResetZoom: () => void;
    onIncreaseZoom: () => void;
    onZoomChange?: (zoom: number) => void;
    onPanChange?: (panX: number, panY: number) => void;
    onRotate?: () => void;
    onClose: () => void;
    onPrevious: () => void;
    onNext: () => void;
    onOpenImage: (id: number) => void;
};

const noPan = { x: 0, y: 0 };
const ignoreZoomChange = () => undefined;
const ignorePanChange = () => undefined;
const ignoreRotate = () => undefined;

export default function ImageViewerModal({
    activeImage,
    activeIndex,
    activeImageDate,
    dateFormatter,
    filteredImages,
    zoom,
    totalImages,
    rotation = 0,
    pan = noPan,
    canGoPrevious = true,
    canGoNext = true,
    showDetails,
    showFilmstrip,
    isSlideshowPlaying,
    isFavoritePending,
    onToggleDetails,
    onToggleFilmstrip,
    onToggleSlideshow,
    onToggleFavorite,
    onOpenFolder,
    onDecreaseZoom,
    onResetZoom,
    onIncreaseZoom,
    onZoomChange = ignoreZoomChange,
    onPanChange = ignorePanChange,
    onRotate = ignoreRotate,
    onClose,
    onPrevious,
    onNext,
    onOpenImage,
}: ImageViewerModalProps) {
    const { t } = useI18n();
    const { details, folderPath, positionLabel } = useImageViewerModal({
        activeImage,
        activeImageDate,
        activeIndex,
        totalImages: totalImages === undefined ? filteredImages.length : totalImages,
        dateFormatter,
    });
    const gestureHandlers = useImageViewerGestures({
        zoom,
        pan,
        onZoomChange,
        onPanChange,
        onPrevious,
        onNext,
    });
    const isFavorite = activeImage.starred;
    const canToggleSlideshow = filteredImages.length > 1;

    return createPortal(
        <div
            className={styles.overlay}
            role="dialog"
            aria-modal="true"
            aria-label={activeImage.name}
        >
            <header className={styles.header}>
                <div className={styles.headerContent}>
                    <strong className={styles.title}>{activeImage.name}</strong>
                    <div className={styles.subtitleRow}>
                        {positionLabel ? <span>{positionLabel}</span> : null}
                        <span>
                            {activeImageDate
                                ? dateFormatter.format(activeImageDate)
                                : t('IMAGES_DATE_UNAVAILABLE')}
                        </span>
                        <span>{folderPath}</span>
                    </div>
                </div>

                <div className={styles.headerActions}>
                    <button
                        type="button"
                        className={`${styles.actionButton} ${isFavorite ? styles.actionButtonActive : ''}`}
                        onClick={onToggleFavorite}
                        disabled={isFavoritePending}
                        aria-pressed={isFavorite}
                        aria-label={
                            isFavorite
                                ? t('IMAGES_VIEWER_REMOVE_FAVORITE')
                                : t('IMAGES_VIEWER_ADD_FAVORITE')
                        }
                    >
                        <Star size={16} fill={isFavorite ? 'currentColor' : 'none'} />
                        <span>
                            {isFavorite
                                ? t('IMAGES_VIEWER_REMOVE_FAVORITE')
                                : t('IMAGES_VIEWER_ADD_FAVORITE')}
                        </span>
                    </button>
                    <button
                        type="button"
                        className={styles.actionButton}
                        onClick={onOpenFolder}
                        aria-label={t('IMAGES_VIEWER_OPEN_FOLDER')}
                    >
                        <FolderOpen size={16} />
                        <span>{t('IMAGES_VIEWER_OPEN_FOLDER')}</span>
                    </button>
                    <a
                        className={styles.actionButton}
                        href={downloadOriginalUrl(activeImage.file_id)}
                        download={activeImage.name}
                        aria-label={t('IMAGES_VIEWER_DOWNLOAD')}
                    >
                        <Download size={16} />
                        <span>{t('IMAGES_VIEWER_DOWNLOAD')}</span>
                    </a>
                    <button
                        type="button"
                        className={styles.actionButton}
                        onClick={onToggleSlideshow}
                        disabled={!canToggleSlideshow}
                        aria-pressed={isSlideshowPlaying}
                        aria-label={
                            isSlideshowPlaying
                                ? t('IMAGES_VIEWER_STOP_SLIDESHOW')
                                : t('IMAGES_VIEWER_START_SLIDESHOW')
                        }
                    >
                        {isSlideshowPlaying ? <Pause size={16} /> : <Play size={16} />}
                        <span>
                            {isSlideshowPlaying
                                ? t('IMAGES_VIEWER_STOP_SLIDESHOW')
                                : t('IMAGES_VIEWER_START_SLIDESHOW')}
                        </span>
                    </button>
                    <div className={styles.utilityActions}>
                        <button
                            type="button"
                            className={styles.iconButton}
                            onClick={onToggleFilmstrip}
                            aria-pressed={showFilmstrip}
                            aria-label={
                                showFilmstrip
                                    ? t('IMAGES_VIEWER_HIDE_FILMSTRIP')
                                    : t('IMAGES_VIEWER_SHOW_FILMSTRIP')
                            }
                        >
                            {showFilmstrip
                                ? t('IMAGES_VIEWER_HIDE_FILMSTRIP_SHORT')
                                : t('IMAGES_VIEWER_SHOW_FILMSTRIP_SHORT')}
                        </button>
                        <button
                            type="button"
                            className={styles.iconButton}
                            onClick={onToggleDetails}
                            aria-pressed={showDetails}
                            aria-label={t('IMAGES_TOGGLE_DETAILS')}
                        >
                            <Info size={16} />
                        </button>
                        <button
                            type="button"
                            className={styles.iconButton}
                            onClick={onRotate}
                            aria-label={t('IMAGES_VIEWER_ROTATE')}
                        >
                            <RotateCw size={16} />
                        </button>
                        <button
                            type="button"
                            className={styles.iconButton}
                            onClick={onDecreaseZoom}
                            aria-label={t('IMAGES_DECREASE_ZOOM')}
                        >
                            <Minus size={16} />
                        </button>
                        <button
                            type="button"
                            className={styles.iconButton}
                            onClick={onResetZoom}
                            aria-label={t('IMAGES_RESET_ZOOM')}
                        >
                            <Expand size={16} />
                        </button>
                        <button
                            type="button"
                            className={styles.iconButton}
                            onClick={onIncreaseZoom}
                            aria-label={t('IMAGES_INCREASE_ZOOM')}
                        >
                            <Plus size={16} />
                        </button>
                        <button
                            type="button"
                            className={styles.iconButton}
                            onClick={onClose}
                            aria-label={t('IMAGES_CLOSE_VIEWER')}
                        >
                            <X size={16} />
                        </button>
                    </div>
                </div>
            </header>

            <div
                className={
                    showDetails
                        ? styles.viewerShell
                        : `${styles.viewerShell} ${styles.viewerShellCompact}`
                }
            >
                <section
                    className={styles.stagePanel}
                    onWheel={(event) => {
                        event.preventDefault();
                        if (event.deltaY < 0) {
                            onIncreaseZoom();
                        }
                        if (event.deltaY > 0) {
                            onDecreaseZoom();
                        }
                    }}
                >
                    <button
                        type="button"
                        className={`${styles.navButton} ${styles.navButtonLeft}`}
                        onClick={onPrevious}
                        disabled={!canGoPrevious}
                        aria-label={t('IMAGES_PREVIOUS')}
                    >
                        <ChevronLeft size={26} />
                    </button>
                    <div
                        className={styles.stageFrame}
                        data-testid="image-viewer-stage"
                        {...gestureHandlers}
                    >
                        <img
                            src={viewerImageUrl(activeImage.file_id, zoom, activeImage.format)}
                            alt={activeImage.name}
                            className={styles.image}
                            draggable={false}
                            style={{
                                transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom}) rotate(${rotation}deg)`,
                            }}
                        />
                    </div>
                    <button
                        type="button"
                        className={`${styles.navButton} ${styles.navButtonRight}`}
                        onClick={onNext}
                        disabled={!canGoNext}
                        aria-label={t('IMAGES_NEXT')}
                    >
                        <ChevronRight size={26} />
                    </button>
                    <div className={styles.stageFooter}>
                        <span>
                            {t('IMAGES_ZOOM_LABEL')}: {Math.round(zoom * 100)}%
                        </span>
                        {positionLabel ? <span>{positionLabel}</span> : null}
                        <span>{t('IMAGES_VIEWER_KEYBOARD_HINT')}</span>
                    </div>
                </section>

                {showDetails ? (
                    <aside className={styles.detailsPanel}>
                        {details.map((section) => (
                            <section key={section.title} className={styles.detailsSection}>
                                <h4>{section.title}</h4>
                                <div className={styles.detailsList}>
                                    {section.items.map((item) => (
                                        <ImageViewerDetailItem
                                            key={`${section.title}-${item.label}`}
                                            item={item}
                                        />
                                    ))}
                                </div>
                            </section>
                        ))}
                    </aside>
                ) : null}
            </div>

            {showFilmstrip ? (
                <div className={styles.filmstrip}>
                    {filteredImages
                        .slice(Math.max(0, activeIndex - 8), activeIndex + 9)
                        .map((item) => (
                            <button
                                type="button"
                                key={item.file_id}
                                onClick={() => onOpenImage(item.file_id)}
                                className={
                                    item.file_id === activeImage.file_id
                                        ? `${styles.filmstripItem} ${styles.filmstripItemActive}`
                                        : styles.filmstripItem
                                }
                                aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: item.name })}
                            >
                                <img
                                    src={thumbnailUrl(item.file_id, FILMSTRIP_THUMBNAIL_SIZE)}
                                    alt={item.name}
                                    loading="lazy"
                                />
                            </button>
                        ))}
                </div>
            ) : null}
        </div>,
        document.body
    );
}

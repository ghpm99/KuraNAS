import {
    ChevronLeft,
    ChevronRight,
    Expand,
    FolderOpen,
    Info,
    Minus,
    Pause,
    Play,
    Plus,
    Star,
    X,
} from 'lucide-react';
import type { IImageData } from '@/components/providers/imageProvider/imageProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { useImageViewerModal } from './useImageViewerModal';
import styles from './ImageViewerModal.module.css';

const thumbnailWidth = 960;
const thumbnailHeight = 720;

const thumbnailUrl = (id: number) =>
    `${getApiV1BaseUrl()}/files/thumbnail/${id}?width=${thumbnailWidth}&height=${thumbnailHeight}`;
const blobUrl = (id: number) => `${getApiV1BaseUrl()}/files/blob/${id}`;

type ImageViewerModalProps = {
    activeImage: IImageData;
    activeIndex: number;
    activeImageDate: Date | null;
    dateFormatter: Intl.DateTimeFormat;
    filteredImages: IImageData[];
    zoom: number;
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
    onClose: () => void;
    onPrevious: () => void;
    onNext: () => void;
    onOpenImage: (id: number) => void;
};

export default function ImageViewerModal({
    activeImage,
    activeIndex,
    activeImageDate,
    dateFormatter,
    filteredImages,
    zoom,
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
        totalImages: filteredImages.length,
        dateFormatter,
    });
    const isFavorite = activeImage.starred;
    const canToggleSlideshow = filteredImages.length > 1;

    return (
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
                        <span>{positionLabel}</span>
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
                        aria-label={t('IMAGES_PREVIOUS')}
                    >
                        <ChevronLeft size={26} />
                    </button>
                    <div className={styles.stageFrame}>
                        <img
                            src={blobUrl(activeImage.id)}
                            alt={activeImage.name}
                            className={styles.image}
                            style={{ transform: `scale(${zoom})` }}
                        />
                    </div>
                    <button
                        type="button"
                        className={`${styles.navButton} ${styles.navButtonRight}`}
                        onClick={onNext}
                        aria-label={t('IMAGES_NEXT')}
                    >
                        <ChevronRight size={26} />
                    </button>
                    <div className={styles.stageFooter}>
                        <span>
                            {t('IMAGES_ZOOM_LABEL')}: {Math.round(zoom * 100)}%
                        </span>
                        <span>{positionLabel}</span>
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
                                        <div
                                            key={`${section.title}-${item.label}`}
                                            className={styles.detailsItem}
                                        >
                                            <span className={styles.detailsLabel}>
                                                {item.label}
                                            </span>
                                            <span className={styles.detailsValue}>
                                                {item.value}
                                            </span>
                                        </div>
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
                                key={item.id}
                                onClick={() => onOpenImage(item.id)}
                                className={
                                    item.id === activeImage.id
                                        ? `${styles.filmstripItem} ${styles.filmstripItemActive}`
                                        : styles.filmstripItem
                                }
                                aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: item.name })}
                            >
                                <img src={thumbnailUrl(item.id)} alt={item.name} loading="lazy" />
                            </button>
                        ))}
                </div>
            ) : null}
        </div>
    );
}

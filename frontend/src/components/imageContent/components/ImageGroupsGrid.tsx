import type { MouseEvent } from 'react';
import { CalendarDays, Star } from 'lucide-react';
import { formatSize } from '@/utils';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    GRID_THUMBNAIL_SIZE,
    gridThumbnailSizes,
    thumbnailSrcSet,
    thumbnailUrl,
} from '../imageThumbnailSources';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import type { ImageDateGroup } from '../imageDateGroups';
import { useLongPress } from '../useLongPress';
import type { ImageSelection, MonthSelectionState } from '../useImageSelection';
import styles from '../ImageContent.module.css';

type ImageGroupsGridProps = {
    groups: ImageDateGroup[];
    onOpenImage: (fileId: number) => void;
    onToggleStar: (fileId: number, wasStarred: boolean) => void;
    selection?: ImageSelection;
};

const imageMetadataSummary = (image: ImageLibraryItem): string => {
    const format = image.format ? `${image.format.replace(/^\./, '')} - ` : '';
    return `${format}${formatSize(image.size)}`;
};

const orientationClassOf = (image: ImageLibraryItem) =>
    image.height > image.width ? styles.portrait : styles.landscape;

type MonthCheckboxProps = {
    label: string;
    state: MonthSelectionState;
    onToggle: () => void;
};

const MonthCheckbox = ({ label, state, onToggle }: MonthCheckboxProps) => {
    const { t } = useI18n();
    return (
        <input
            type="checkbox"
            className={styles.monthCheckbox}
            checked={state === 'all'}
            ref={(element) => {
                if (element) {
                    element.indeterminate = state === 'partial';
                }
            }}
            onChange={onToggle}
            aria-label={t('IMAGES_SELECT_MONTH_ARIA', { month: label })}
        />
    );
};

export default function ImageGroupsGrid({
    groups = [],
    onOpenImage,
    onToggleStar,
    selection,
}: ImageGroupsGridProps) {
    const { t } = useI18n();
    const hasSelection = selection?.hasSelection ?? false;
    const loadedImages = groups.flatMap((group) => group.items);
    const { startLongPress, cancelLongPress, consumeFiredLongPress } = useLongPress(
        (image: ImageLibraryItem) => selection?.toggle(image)
    );

    const handleImageClick = (event: MouseEvent, image: ImageLibraryItem) => {
        if (consumeFiredLongPress()) {
            return;
        }
        if (!selection) {
            onOpenImage(image.file_id);
            return;
        }
        if (event.shiftKey) {
            selection.selectRange(image, loadedImages);
            return;
        }
        if (event.ctrlKey || event.metaKey || selection.hasSelection) {
            selection.toggle(image);
            return;
        }
        onOpenImage(image.file_id);
    };

    const handleCheckboxClick = (event: MouseEvent, image: ImageLibraryItem) => {
        if (event.shiftKey) {
            selection?.selectRange(image, loadedImages);
            return;
        }
        selection?.toggle(image);
    };

    return (
        <div
            className={`${styles.sections} ${hasSelection ? styles.selectionActive : ''}`}
        >
            {groups.map((group) => (
                <section key={group.key} className={styles.group}>
                    {group.label && (
                        <header className={styles.groupHeader}>
                            {selection && (
                                <MonthCheckbox
                                    label={group.label}
                                    state={selection.readMonthSelectionState(group.items)}
                                    onToggle={() => selection.toggleMonth(group.items)}
                                />
                            )}
                            <CalendarDays size={16} />
                            <h3>{group.label}</h3>
                            <span>
                                {t('IMAGES_PHOTOS_COUNT', {
                                    count: String(group.totalCount ?? group.items.length),
                                })}
                            </span>
                        </header>
                    )}
                    <div className={styles.grid}>
                        {group.items.map((item) => (
                            <div
                                key={item.file_id}
                                className={`${styles.photoCard} ${orientationClassOf(item)} ${
                                    selection?.isSelected(item.file_id) ? styles.photoCardSelected : ''
                                }`}
                            >
                                <button
                                    type="button"
                                    className={styles.photoButton}
                                    onClick={(event) => handleImageClick(event, item)}
                                    onPointerDown={(event) => startLongPress(event, item)}
                                    onPointerUp={cancelLongPress}
                                    onPointerLeave={cancelLongPress}
                                    onPointerCancel={cancelLongPress}
                                    onContextMenu={(event) => {
                                        if (selection) {
                                            event.preventDefault();
                                        }
                                    }}
                                    aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: item.name })}
                                >
                                    <img
                                        className={styles.thumbnail}
                                        src={thumbnailUrl(item.file_id, GRID_THUMBNAIL_SIZE)}
                                        srcSet={thumbnailSrcSet(item.file_id, GRID_THUMBNAIL_SIZE)}
                                        sizes={gridThumbnailSizes}
                                        alt={item.name}
                                        loading="lazy"
                                    />
                                    <span className={styles.photoOverlay}>
                                        <span className={styles.photoName}>{item.name}</span>
                                        <span>{imageMetadataSummary(item)}</span>
                                    </span>
                                </button>
                                {selection && (
                                    <input
                                        type="checkbox"
                                        className={styles.selectCheckbox}
                                        checked={selection.isSelected(item.file_id)}
                                        readOnly
                                        aria-label={t('IMAGES_SELECT_IMAGE_ARIA', { name: item.name })}
                                        onClick={(event) => handleCheckboxClick(event, item)}
                                    />
                                )}
                                <button
                                    type="button"
                                    className={
                                        item.starred
                                            ? `${styles.starButton} ${styles.starButtonActive}`
                                            : styles.starButton
                                    }
                                    aria-pressed={item.starred}
                                    aria-label={
                                        item.starred
                                            ? t('IMAGES_STAR_REMOVE_ARIA', { name: item.name })
                                            : t('IMAGES_STAR_ADD_ARIA', { name: item.name })
                                    }
                                    onClick={() => onToggleStar(item.file_id, item.starred)}
                                >
                                    <Star size={16} fill={item.starred ? 'currentColor' : 'none'} />
                                </button>
                            </div>
                        ))}
                    </div>
                </section>
            ))}
        </div>
    );
}

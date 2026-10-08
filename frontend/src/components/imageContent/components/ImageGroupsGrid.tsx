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
import styles from '../ImageContent.module.css';

type ImageGroupsGridProps = {
    groups: ImageDateGroup[];
    onOpenImage: (fileId: number) => void;
    onToggleStar: (fileId: number, wasStarred: boolean) => void;
};

const imageMetadataSummary = (image: ImageLibraryItem): string => {
    const format = image.format ? `${image.format.replace(/^\./, '')} - ` : '';
    return `${format}${formatSize(image.size)}`;
};

const orientationClassOf = (image: ImageLibraryItem) =>
    image.height > image.width ? styles.portrait : styles.landscape;

export default function ImageGroupsGrid({
    groups = [],
    onOpenImage,
    onToggleStar,
}: ImageGroupsGridProps) {
    const { t } = useI18n();

    return (
        <div className={styles.sections}>
            {groups.map((group) => (
                <section key={group.key} className={styles.group}>
                    {group.label && (
                        <header className={styles.groupHeader}>
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
                                className={`${styles.photoCard} ${orientationClassOf(item)}`}
                            >
                                <button
                                    type="button"
                                    className={styles.photoButton}
                                    onClick={() => onOpenImage(item.file_id)}
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

import type {
    VideoLibraryDuration,
    VideoLibraryFilters,
    VideoLibraryResolution,
} from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    hasActiveVideoLibraryFilters,
    noVideoLibraryFilters,
    toggleVideoLibraryFormat,
    videoLibraryDurationOptions,
    videoLibraryFormatOptions,
    videoLibraryResolutionOptions,
} from '../../../providers/videoContentProvider/videoLibraryFilters';
import styles from '../videoContent.module.css';

type VideoLibraryFilterBarProps = {
    filters?: VideoLibraryFilters;
    onFiltersChange?: (filters: VideoLibraryFilters) => void;
};

const resolutionLabelKeys: Record<VideoLibraryResolution, string> = {
    sd: 'VIDEO_FILTER_RESOLUTION_SD',
    hd: 'VIDEO_FILTER_RESOLUTION_HD',
    fhd: 'VIDEO_FILTER_RESOLUTION_FHD',
    uhd: 'VIDEO_FILTER_RESOLUTION_UHD',
};

const durationLabelKeys: Record<VideoLibraryDuration, string> = {
    short: 'VIDEO_FILTER_DURATION_SHORT',
    medium: 'VIDEO_FILTER_DURATION_MEDIUM',
    long: 'VIDEO_FILTER_DURATION_LONG',
};

export default function VideoLibraryFilterBar({
    filters = noVideoLibraryFilters,
    onFiltersChange,
}: VideoLibraryFilterBarProps) {
    const { t } = useI18n();
    const hasActiveFilters = hasActiveVideoLibraryFilters(filters);

    return (
        <div className={styles.filterBar}>
            <div className={styles.filterSelectRow}>
                <select
                    className={styles.playlistSelect}
                    aria-label={t('VIDEO_FILTER_RESOLUTION_LABEL')}
                    value={filters.resolution ?? ''}
                    onChange={(event) =>
                        onFiltersChange?.({
                            ...filters,
                            resolution: (event.target.value || undefined) as
                                | VideoLibraryResolution
                                | undefined,
                        })
                    }
                >
                    <option value="">{t('VIDEO_FILTER_RESOLUTION_ANY')}</option>
                    {videoLibraryResolutionOptions.map((resolution) => (
                        <option key={resolution} value={resolution}>
                            {t(resolutionLabelKeys[resolution])}
                        </option>
                    ))}
                </select>
                <select
                    className={styles.playlistSelect}
                    aria-label={t('VIDEO_FILTER_DURATION_LABEL')}
                    value={filters.duration ?? ''}
                    onChange={(event) =>
                        onFiltersChange?.({
                            ...filters,
                            duration: (event.target.value || undefined) as
                                | VideoLibraryDuration
                                | undefined,
                        })
                    }
                >
                    <option value="">{t('VIDEO_FILTER_DURATION_ANY')}</option>
                    {videoLibraryDurationOptions.map((duration) => (
                        <option key={duration} value={duration}>
                            {t(durationLabelKeys[duration])}
                        </option>
                    ))}
                </select>
                {hasActiveFilters && (
                    <button
                        type="button"
                        className={styles.actionBtn}
                        onClick={() => onFiltersChange?.(noVideoLibraryFilters)}
                    >
                        {t('VIDEO_FILTER_CLEAR')}
                    </button>
                )}
            </div>
            <div
                className={styles.filterChips}
                role="group"
                aria-label={t('VIDEO_FILTER_FORMAT_LABEL')}
            >
                {videoLibraryFormatOptions.map((format) => {
                    const isSelected = filters.formats.includes(format);
                    return (
                        <button
                            key={format}
                            type="button"
                            aria-pressed={isSelected}
                            className={isSelected ? styles.filterChipActive : styles.filterChip}
                            onClick={() =>
                                onFiltersChange?.(toggleVideoLibraryFormat(filters, format))
                            }
                        >
                            {format.slice(1).toUpperCase()}
                        </button>
                    );
                })}
            </div>
        </div>
    );
}

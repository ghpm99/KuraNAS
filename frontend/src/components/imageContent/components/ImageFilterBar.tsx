import { ArrowDownWideNarrow, ArrowUpNarrowWide } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    imageFormats,
    type ImageLibraryFilters,
    type ImageLibraryOrdering,
    type ImageLibrarySort,
} from '@/types/imageLibrary';
import styles from './ImageFilterBar.module.css';

const sortOptions: { sort: ImageLibrarySort; labelKey: string }[] = [
    { sort: 'taken_at', labelKey: 'IMAGES_SORT_TAKEN_AT' },
    { sort: 'name', labelKey: 'IMAGES_SORT_NAME' },
    { sort: 'size', labelKey: 'IMAGES_SORT_SIZE' },
];

type ImageFilterBarProps = {
    filters: ImageLibraryFilters;
    ordering: ImageLibraryOrdering;
    hasUserFilters: boolean;
    onTakenFromChange: (takenFrom: string) => void;
    onTakenToChange: (takenTo: string) => void;
    onFormatToggle: (format: string) => void;
    onSortChange: (sort: ImageLibrarySort) => void;
    onSortOrderToggle: () => void;
    onClearFilters: () => void;
};

export default function ImageFilterBar({
    filters,
    ordering,
    hasUserFilters,
    onTakenFromChange,
    onTakenToChange,
    onFormatToggle,
    onSortChange,
    onSortOrderToggle,
    onClearFilters,
}: ImageFilterBarProps) {
    const { t } = useI18n();
    const isDescending = ordering.order === 'desc';
    const orderLabel = isDescending ? t('IMAGES_SORT_ORDER_DESC') : t('IMAGES_SORT_ORDER_ASC');

    return (
        <section className={styles.bar} aria-label={t('IMAGES_FILTER_BAR_ARIA')}>
            <div className={styles.group}>
                <label className={styles.field}>
                    <span>{t('IMAGES_FILTER_FROM')}</span>
                    <input
                        type="date"
                        value={filters.takenFrom}
                        max={filters.takenTo || undefined}
                        onChange={(event) => onTakenFromChange(event.target.value)}
                    />
                </label>
                <label className={styles.field}>
                    <span>{t('IMAGES_FILTER_TO')}</span>
                    <input
                        type="date"
                        value={filters.takenTo}
                        min={filters.takenFrom || undefined}
                        onChange={(event) => onTakenToChange(event.target.value)}
                    />
                </label>
            </div>
            <div className={styles.group} role="group" aria-label={t('IMAGES_FILTER_FORMATS_ARIA')}>
                {imageFormats.map((format) => {
                    const isSelected = filters.formats.includes(format);
                    return (
                        <button
                            type="button"
                            key={format}
                            className={
                                isSelected ? `${styles.chip} ${styles.chipActive}` : styles.chip
                            }
                            aria-pressed={isSelected}
                            onClick={() => onFormatToggle(format)}
                        >
                            {format.toUpperCase()}
                        </button>
                    );
                })}
            </div>
            <div className={styles.group}>
                <label className={styles.field}>
                    <span>{t('IMAGES_SORT_LABEL')}</span>
                    <select
                        value={ordering.sort}
                        onChange={(event) => onSortChange(event.target.value as ImageLibrarySort)}
                    >
                        {sortOptions.map((option) => (
                            <option key={option.sort} value={option.sort}>
                                {t(option.labelKey)}
                            </option>
                        ))}
                    </select>
                </label>
                <button
                    type="button"
                    className={styles.chip}
                    onClick={onSortOrderToggle}
                    aria-label={orderLabel}
                >
                    {isDescending ? (
                        <ArrowDownWideNarrow size={14} />
                    ) : (
                        <ArrowUpNarrowWide size={14} />
                    )}
                    <span>{orderLabel}</span>
                </button>
            </div>
            {hasUserFilters && (
                <button type="button" className={styles.clearButton} onClick={onClearFilters}>
                    {t('IMAGES_FILTER_CLEAR')}
                </button>
            )}
        </section>
    );
}

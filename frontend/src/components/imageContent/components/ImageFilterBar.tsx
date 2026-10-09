import { ArrowDownWideNarrow, ArrowUpNarrowWide } from 'lucide-react';
import { Autocomplete, MenuItem, Select, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    type ImageCameraFacet,
    type ImageFormatFacet,
    type ImageLibraryFilters,
    type ImageLibraryOrdering,
    type ImageLibrarySort,
} from '@/types/imageLibrary';
import styles from './ImageFilterBar.module.css';
import { useImageLibraryFacets } from '../useImageLibraryFacets';

const sortOptions: { sort: ImageLibrarySort; labelKey: string }[] = [
    { sort: 'taken_at', labelKey: 'IMAGES_SORT_TAKEN_AT' },
    { sort: 'name', labelKey: 'IMAGES_SORT_NAME' },
    { sort: 'size', labelKey: 'IMAGES_SORT_SIZE' },
];

const withSelectedCamera = (cameraFacets: ImageCameraFacet[], selectedCamera: string) => {
    const cameraNames = cameraFacets.map((facet) => facet.camera);
    return selectedCamera && !cameraNames.includes(selectedCamera)
        ? [selectedCamera, ...cameraNames]
        : cameraNames;
};

const withSelectedFormats = (
    formatFacets: ImageFormatFacet[],
    selectedFormats: string[]
): ImageFormatFacet[] => {
    const listedFormats = formatFacets.map((facet) => facet.format);
    const missingSelected = selectedFormats
        .filter((format) => !listedFormats.includes(format))
        .map((format) => ({ format, count: 0 }));
    return [...formatFacets, ...missingSelected];
};

const toFormatList = (selectedValue: string | string[]) =>
    typeof selectedValue === 'string' ? selectedValue.split(',') : selectedValue;

const findToggledFormat = (previousFormats: string[], nextFormats: string[]) => [
    ...nextFormats.filter((format) => !previousFormats.includes(format)),
    ...previousFormats.filter((format) => !nextFormats.includes(format)),
];

type ImageFilterBarProps = {
    filters: ImageLibraryFilters;
    ordering: ImageLibraryOrdering;
    hasUserFilters: boolean;
    onTakenFromChange: (takenFrom: string) => void;
    onTakenToChange: (takenTo: string) => void;
    onFormatToggle: (format: string) => void;
    onCameraChange: (camera: string) => void;
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
    onCameraChange,
    onSortChange,
    onSortOrderToggle,
    onClearFilters,
}: ImageFilterBarProps) {
    const { t } = useI18n();
    const { cameraFacets, formatFacets } = useImageLibraryFacets(filters);
    const cameraOptions = withSelectedCamera(cameraFacets, filters.camera);
    const cameraCounts = new Map(cameraFacets.map((facet) => [facet.camera, facet.count]));
    const formatOptions = withSelectedFormats(formatFacets, filters.formats);
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
            <div className={styles.group}>
                <Autocomplete
                    size="small"
                    className={styles.cameraField}
                    options={cameraOptions}
                    value={filters.camera || null}
                    onChange={(_, selectedCamera) => onCameraChange(selectedCamera ?? '')}
                    renderOption={(optionProps, camera) => (
                        <li {...optionProps} key={camera}>
                            {camera} ({cameraCounts.get(camera) ?? 0})
                        </li>
                    )}
                    noOptionsText={t('IMAGES_FILTER_CAMERA_EMPTY')}
                    clearText={t('IMAGES_FILTER_CAMERA_CLEAR')}
                    openText={t('IMAGES_FILTER_CAMERA_OPEN')}
                    closeText={t('IMAGES_FILTER_CAMERA_CLOSE')}
                    renderInput={(params) => (
                        <TextField {...params} label={t('IMAGES_FILTER_CAMERA')} />
                    )}
                />
                <Select
                    multiple
                    displayEmpty
                    size="small"
                    className={styles.formatField}
                    value={filters.formats}
                    inputProps={{ 'aria-label': t('IMAGES_FILTER_FORMATS_ARIA') }}
                    renderValue={(selectedFormats) =>
                        selectedFormats.length === 0
                            ? t('IMAGES_FILTER_FORMATS_ALL')
                            : selectedFormats.map((format) => format.toUpperCase()).join(', ')
                    }
                    onChange={(event) =>
                        findToggledFormat(
                            filters.formats,
                            toFormatList(event.target.value)
                        ).forEach((format) => onFormatToggle(format))
                    }
                >
                    {formatOptions.map((facet) => (
                        <MenuItem key={facet.format} value={facet.format}>
                            {facet.format.toUpperCase()} ({facet.count})
                        </MenuItem>
                    ))}
                </Select>
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

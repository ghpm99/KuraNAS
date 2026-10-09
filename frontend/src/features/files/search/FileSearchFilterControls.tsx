import {
    Button,
    Checkbox,
    Chip,
    FormControlLabel,
    IconButton,
    MenuItem,
    Radio,
    RadioGroup,
    Select,
    TextField,
    Tooltip,
} from '@mui/material';
import { ArrowDown, ArrowUp, Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { fileSearchKinds, type FileSearchKind } from '@/types/fileSearch';
import FilterChipPopover from './FilterChipPopover';
import {
    emptyFileSearchFilters,
    fileSearchSizePresets,
    fileSearchSorts,
    fileSearchTiers,
    type FileSearchFilters,
    type FileSearchSizePreset,
} from './fileSearchFilters';
import styles from './fileSearch.module.css';

interface FileSearchFilterControlsProps {
    filters?: FileSearchFilters;
    onChange?: (filters: FileSearchFilters) => void;
}

const kindLabelKeys: Record<FileSearchKind, string> = {
    folder: 'FILES_SEARCH_KIND_FOLDER',
    document: 'FILES_SEARCH_KIND_DOCUMENT',
    image: 'FILES_SEARCH_KIND_IMAGE',
    audio: 'FILES_SEARCH_KIND_AUDIO',
    video: 'FILES_SEARCH_KIND_VIDEO',
    archive: 'FILES_SEARCH_KIND_ARCHIVE',
    other: 'FILES_SEARCH_KIND_OTHER',
};

const sizeLabelKeys: Record<FileSearchSizePreset, string> = {
    small: 'FILES_SEARCH_SIZE_SMALL',
    medium: 'FILES_SEARCH_SIZE_MEDIUM',
    large: 'FILES_SEARCH_SIZE_LARGE',
    huge: 'FILES_SEARCH_SIZE_HUGE',
};

const tierLabelKeys = {
    hot: 'FILES_SEARCH_TIER_HOT',
    cold: 'FILES_SEARCH_TIER_COLD',
};

const sortLabelKeys = {
    relevance: 'FILES_SEARCH_SORT_RELEVANCE',
    name: 'FILES_SORT_NAME',
    size: 'FILES_SORT_SIZE',
    modified: 'FILES_SORT_UPDATED_AT',
};

const defaultOrderBySort = { relevance: 'desc', name: 'asc', size: 'desc', modified: 'desc' };

const FileSearchFilterControls = ({
    filters = emptyFileSearchFilters,
    onChange,
}: FileSearchFilterControlsProps) => {
    const { t } = useI18n();
    const update = (changes: Partial<FileSearchFilters>) => onChange?.({ ...filters, ...changes });

    const toggleKind = (kind: FileSearchKind, isSelected: boolean) =>
        update({
            kinds: fileSearchKinds.filter((candidate) =>
                candidate === kind ? isSelected : filters.kinds.includes(candidate)
            ),
        });

    const effectiveOrder = filters.order || defaultOrderBySort[filters.sort];
    const isDescending = effectiveOrder === 'desc';
    const orderLabel = t(isDescending ? 'FILES_SORT_ORDER_DESCENDING' : 'FILES_SORT_ORDER_ASCENDING');
    const kindsLabel =
        filters.kinds.length > 0
            ? `${t('FILES_SEARCH_FILTER_KIND')} (${filters.kinds.length})`
            : t('FILES_SEARCH_FILTER_KIND');

    return (
        <div className={styles.filterControls}>
            <FilterChipPopover label={kindsLabel} isActive={filters.kinds.length > 0}>
                {fileSearchKinds.map((kind) => (
                    <FormControlLabel
                        key={kind}
                        label={t(kindLabelKeys[kind])}
                        control={
                            <Checkbox
                                size="small"
                                checked={filters.kinds.includes(kind)}
                                onChange={(event) => toggleKind(kind, event.target.checked)}
                            />
                        }
                    />
                ))}
            </FilterChipPopover>

            <FilterChipPopover
                label={t('FILES_SEARCH_FILTER_PERIOD')}
                isActive={filters.modifiedFrom !== '' || filters.modifiedTo !== ''}
            >
                <TextField
                    type="date"
                    size="small"
                    label={t('FILES_SEARCH_PERIOD_FROM')}
                    value={filters.modifiedFrom}
                    onChange={(event) => update({ modifiedFrom: event.target.value })}
                    slotProps={{
                        inputLabel: { shrink: true },
                        htmlInput: { max: filters.modifiedTo || undefined },
                    }}
                />
                <TextField
                    type="date"
                    size="small"
                    label={t('FILES_SEARCH_PERIOD_TO')}
                    value={filters.modifiedTo}
                    onChange={(event) => update({ modifiedTo: event.target.value })}
                    slotProps={{
                        inputLabel: { shrink: true },
                        htmlInput: { min: filters.modifiedFrom || undefined },
                    }}
                />
            </FilterChipPopover>

            <FilterChipPopover
                label={t('FILES_SEARCH_FILTER_SIZE')}
                isActive={filters.sizePreset !== ''}
            >
                <RadioGroup
                    value={filters.sizePreset}
                    onChange={(event) =>
                        update({ sizePreset: event.target.value as FileSearchSizePreset })
                    }
                >
                    {fileSearchSizePresets.map((preset) => (
                        <FormControlLabel
                            key={preset}
                            value={preset}
                            label={t(sizeLabelKeys[preset])}
                            control={<Radio size="small" />}
                        />
                    ))}
                </RadioGroup>
                {filters.sizePreset ? (
                    <Button size="small" onClick={() => update({ sizePreset: '' })}>
                        {t('FILES_SEARCH_FILTERS_CLEAR')}
                    </Button>
                ) : null}
            </FilterChipPopover>

            <FilterChipPopover label={t('FILES_SEARCH_FILTER_TIER')} isActive={filters.tier !== ''}>
                <RadioGroup
                    value={filters.tier}
                    onChange={(event) => update({ tier: event.target.value as 'hot' | 'cold' })}
                >
                    {fileSearchTiers.map((tier) => (
                        <FormControlLabel
                            key={tier}
                            value={tier}
                            label={t(tierLabelKeys[tier])}
                            control={<Radio size="small" />}
                        />
                    ))}
                </RadioGroup>
                {filters.tier ? (
                    <Button size="small" onClick={() => update({ tier: '' })}>
                        {t('FILES_SEARCH_FILTERS_CLEAR')}
                    </Button>
                ) : null}
            </FilterChipPopover>

            <Chip
                label={t('FILES_SEARCH_FILTER_STARRED')}
                size="small"
                icon={<Star size={14} />}
                color={filters.onlyStarred ? 'primary' : 'default'}
                variant={filters.onlyStarred ? 'filled' : 'outlined'}
                aria-pressed={filters.onlyStarred}
                onClick={() => update({ onlyStarred: !filters.onlyStarred })}
            />

            <div className={styles.sortControls}>
                <Select
                    size="small"
                    value={filters.sort}
                    onChange={(event) =>
                        update({ sort: event.target.value as FileSearchFilters['sort'], order: '' })
                    }
                    inputProps={{ 'aria-label': t('FILES_SORT_LABEL') }}
                >
                    {fileSearchSorts.map((sort) => (
                        <MenuItem key={sort} value={sort}>
                            {t(sortLabelKeys[sort])}
                        </MenuItem>
                    ))}
                </Select>
                <Tooltip title={orderLabel}>
                    <span>
                        <IconButton
                            size="small"
                            aria-label={orderLabel}
                            disabled={filters.sort === 'relevance'}
                            onClick={() => update({ order: isDescending ? 'asc' : 'desc' })}
                        >
                            {isDescending ? <ArrowDown size={16} /> : <ArrowUp size={16} />}
                        </IconButton>
                    </span>
                </Tooltip>
            </div>
        </div>
    );
};

export default FileSearchFilterControls;

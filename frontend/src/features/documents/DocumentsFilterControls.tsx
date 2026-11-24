import { Button, Chip, FormControlLabel, Radio, RadioGroup, TextField } from '@mui/material';
import { Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import FilterChipPopover from '@/features/files/search/FilterChipPopover';
import type { DocumentFormat, DocumentIndexStateName } from '@/service/documents';
import {
    documentFormats,
    documentIndexStates,
    emptyDocumentsFilters,
    type DocumentsFilterState,
} from './documentsFilters';
import styles from './DocumentsFilters.module.css';

interface DocumentsFilterControlsProps {
    filters?: DocumentsFilterState;
    onChange?: (filters: DocumentsFilterState) => void;
}

const formatLabelKeys: Record<DocumentFormat, string> = {
    pdf: 'DOCUMENTS_FORMAT_PDF',
    text: 'DOCUMENTS_FORMAT_TEXT',
    spreadsheet: 'DOCUMENTS_FORMAT_SPREADSHEET',
    presentation: 'DOCUMENTS_FORMAT_PRESENTATION',
    image: 'DOCUMENTS_FORMAT_IMAGE',
    other: 'DOCUMENTS_FORMAT_OTHER',
};

const indexStateLabelKeys: Record<DocumentIndexStateName, string> = {
    indexed: 'DOCUMENTS_INDEX_STATE_INDEXED',
    partial: 'DOCUMENTS_INDEX_STATE_PARTIAL',
    name_only: 'DOCUMENTS_INDEX_STATE_NAME_ONLY',
    queued: 'DOCUMENTS_INDEX_STATE_QUEUED',
    failed: 'DOCUMENTS_INDEX_STATE_FAILED',
};

const DocumentsFilterControls = ({
    filters = emptyDocumentsFilters,
    onChange,
}: DocumentsFilterControlsProps) => {
    const { t } = useI18n();
    const update = (changes: Partial<DocumentsFilterState>) =>
        onChange?.({ ...filters, ...changes });

    return (
        <div className={styles.filterControls}>
            <FilterChipPopover
                label={t('DOCUMENTS_FILTER_FORMAT')}
                isActive={filters.format !== ''}
            >
                <RadioGroup
                    value={filters.format}
                    onChange={(event) => update({ format: event.target.value as DocumentFormat })}
                >
                    {documentFormats.map((format) => (
                        <FormControlLabel
                            key={format}
                            value={format}
                            label={t(formatLabelKeys[format])}
                            control={<Radio size="small" />}
                        />
                    ))}
                </RadioGroup>
                {filters.format ? (
                    <Button size="small" onClick={() => update({ format: '' })}>
                        {t('DOCUMENTS_FILTERS_CLEAR')}
                    </Button>
                ) : null}
            </FilterChipPopover>

            <FilterChipPopover
                label={t('DOCUMENTS_FILTER_PERIOD')}
                isActive={filters.modifiedFrom !== '' || filters.modifiedTo !== ''}
            >
                <TextField
                    type="date"
                    size="small"
                    label={t('DOCUMENTS_FILTER_PERIOD_FROM')}
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
                    label={t('DOCUMENTS_FILTER_PERIOD_TO')}
                    value={filters.modifiedTo}
                    onChange={(event) => update({ modifiedTo: event.target.value })}
                    slotProps={{
                        inputLabel: { shrink: true },
                        htmlInput: { min: filters.modifiedFrom || undefined },
                    }}
                />
            </FilterChipPopover>

            <FilterChipPopover
                label={t('DOCUMENTS_FILTER_INDEX_STATE')}
                isActive={filters.indexState !== ''}
            >
                <RadioGroup
                    value={filters.indexState}
                    onChange={(event) =>
                        update({ indexState: event.target.value as DocumentIndexStateName })
                    }
                >
                    {documentIndexStates.map((indexState) => (
                        <FormControlLabel
                            key={indexState}
                            value={indexState}
                            label={t(indexStateLabelKeys[indexState])}
                            control={<Radio size="small" />}
                        />
                    ))}
                </RadioGroup>
                {filters.indexState ? (
                    <Button size="small" onClick={() => update({ indexState: '' })}>
                        {t('DOCUMENTS_FILTERS_CLEAR')}
                    </Button>
                ) : null}
            </FilterChipPopover>

            <Chip
                label={t('DOCUMENTS_FILTER_FAVORITES')}
                size="small"
                icon={<Star size={14} />}
                color={filters.favoritesOnly ? 'primary' : 'default'}
                variant={filters.favoritesOnly ? 'filled' : 'outlined'}
                aria-pressed={filters.favoritesOnly}
                onClick={() => update({ favoritesOnly: !filters.favoritesOnly })}
            />
        </div>
    );
};

export default DocumentsFilterControls;

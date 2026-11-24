import { IconButton, MenuItem, Select, Tooltip } from '@mui/material';
import { ArrowDown, ArrowUp } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { DocumentSortField } from '@/service/documents';
import {
    defaultDocumentsSort,
    defaultOrderFor,
    documentSortFields,
    type DocumentsSortState,
} from './documentsFilters';
import styles from './DocumentsFilters.module.css';

interface DocumentsSortControlProps {
    sortState?: DocumentsSortState;
    onChange?: (sortState: DocumentsSortState) => void;
}

const sortLabelKeys: Record<DocumentSortField, string> = {
    modified_at: 'DOCUMENTS_SORT_MODIFIED_AT',
    name: 'DOCUMENTS_SORT_NAME',
    size: 'DOCUMENTS_SORT_SIZE',
    created_at: 'DOCUMENTS_SORT_CREATED_AT',
};

const DocumentsSortControl = ({
    sortState = defaultDocumentsSort,
    onChange,
}: DocumentsSortControlProps) => {
    const { t } = useI18n();
    const isDescending = sortState.order === 'desc';
    const orderLabel = t(
        isDescending ? 'DOCUMENTS_SORT_ORDER_DESCENDING' : 'DOCUMENTS_SORT_ORDER_ASCENDING'
    );

    return (
        <div className={styles.sortControl}>
            <Select
                size="small"
                value={sortState.sort}
                onChange={(event) => {
                    const nextSort = event.target.value as DocumentSortField;
                    onChange?.({ sort: nextSort, order: defaultOrderFor(nextSort) });
                }}
                inputProps={{ 'aria-label': t('DOCUMENTS_SORT_LABEL') }}
            >
                {documentSortFields.map((sortField) => (
                    <MenuItem key={sortField} value={sortField}>
                        {t(sortLabelKeys[sortField])}
                    </MenuItem>
                ))}
            </Select>
            <Tooltip title={orderLabel}>
                <IconButton
                    size="small"
                    aria-label={orderLabel}
                    onClick={() =>
                        onChange?.({ ...sortState, order: isDescending ? 'asc' : 'desc' })
                    }
                >
                    {isDescending ? <ArrowDown size={16} /> : <ArrowUp size={16} />}
                </IconButton>
            </Tooltip>
        </div>
    );
};

export default DocumentsSortControl;

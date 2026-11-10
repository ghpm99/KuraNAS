import useI18n from '@/components/i18n/provider/i18nContext';
import type { TrashSortKey } from '@/types/trash';
import { IconButton, MenuItem, Select, TextField, Tooltip } from '@mui/material';
import { ArrowDown, ArrowUp } from 'lucide-react';
import styles from './TrashScreen.module.css';
import type { TrashSort } from './useTrashScreen';

type TrashListingControlsProps = {
    sort: TrashSort;
    onSortChange: (sort: TrashSort) => void;
    searchInput: string;
    onSearchInputChange: (typedText: string) => void;
};

const sortKeyLabelKeys: Record<TrashSortKey, string> = {
    deleted_at: 'TRASH_SORT_DELETED_AT',
    name: 'TRASH_SORT_NAME',
    size: 'TRASH_SORT_SIZE',
};

const sortKeys = Object.keys(sortKeyLabelKeys) as TrashSortKey[];

const TrashListingControls = ({
    sort,
    onSortChange,
    searchInput,
    onSearchInputChange,
}: TrashListingControlsProps) => {
    const { t } = useI18n();
    const isDescending = sort.order === 'desc';
    const orderLabel = t(
        isDescending ? 'TRASH_SORT_ORDER_DESCENDING' : 'TRASH_SORT_ORDER_ASCENDING'
    );

    return (
        <div className={styles.listingControls}>
            <TextField
                size="small"
                type="search"
                className={styles.searchField}
                label={t('TRASH_SEARCH_LABEL')}
                value={searchInput}
                onChange={(event) => onSearchInputChange(event.target.value)}
            />
            <div className={styles.sortControl}>
                <Select
                    size="small"
                    value={sort.sortKey}
                    onChange={(event) =>
                        onSortChange({ ...sort, sortKey: event.target.value as TrashSortKey })
                    }
                    inputProps={{ 'aria-label': t('TRASH_SORT_LABEL') }}
                    className={styles.sortSelect}
                >
                    {sortKeys.map((sortKey) => (
                        <MenuItem key={sortKey} value={sortKey}>
                            {t(sortKeyLabelKeys[sortKey])}
                        </MenuItem>
                    ))}
                </Select>
                <Tooltip title={orderLabel}>
                    <IconButton
                        size="small"
                        aria-label={orderLabel}
                        onClick={() =>
                            onSortChange({ ...sort, order: isDescending ? 'asc' : 'desc' })
                        }
                    >
                        {isDescending ? <ArrowDown size={16} /> : <ArrowUp size={16} />}
                    </IconButton>
                </Tooltip>
            </div>
        </div>
    );
};

export default TrashListingControls;

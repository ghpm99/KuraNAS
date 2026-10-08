import useI18n from '@/components/i18n/provider/i18nContext';
import { IconButton, MenuItem, Select, Tooltip } from '@mui/material';
import { ArrowDown, ArrowUp } from 'lucide-react';
import type { FilesSort, FilesSortKey } from '@/features/files/providers/fileProvider/fileContext';
import { defaultFilesSort } from '@/features/files/providers/fileProvider/filesSortPreference';
import styles from './filesSortControl.module.css';

type FilesSortControlProps = {
    sort?: FilesSort;
    onChange?: (sort: FilesSort) => void;
};

const sortKeyLabelKeys: Record<FilesSortKey, string> = {
    name: 'FILES_SORT_NAME',
    size: 'FILES_SORT_SIZE',
    updated_at: 'FILES_SORT_UPDATED_AT',
    created_at: 'FILES_SORT_CREATED_AT',
};

const sortKeys = Object.keys(sortKeyLabelKeys) as FilesSortKey[];

const FilesSortControl = ({ sort = defaultFilesSort, onChange }: FilesSortControlProps) => {
    const { t } = useI18n();
    const isDescending = sort.order === 'desc';
    const orderLabel = t(
        isDescending ? 'FILES_SORT_ORDER_DESCENDING' : 'FILES_SORT_ORDER_ASCENDING'
    );

    return (
        <div className={styles.sortControl}>
            <Select
                size="small"
                value={sort.key}
                onChange={(event) =>
                    onChange?.({ ...sort, key: event.target.value as FilesSortKey })
                }
                inputProps={{ 'aria-label': t('FILES_SORT_LABEL') }}
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
                    onClick={() => onChange?.({ ...sort, order: isDescending ? 'asc' : 'desc' })}
                >
                    {isDescending ? <ArrowDown size={16} /> : <ArrowUp size={16} />}
                </IconButton>
            </Tooltip>
        </div>
    );
};

export default FilesSortControl;

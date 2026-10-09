import { ArrowDown, ArrowUp } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FilesSort, FilesSortKey } from '@/features/files/providers/fileProvider/fileContext';
import { defaultFilesSort } from '@/features/files/providers/fileProvider/filesSortPreference';
import styles from './fileListHeader.module.css';

type FileListHeaderProps = {
    sort?: FilesSort;
    onSortChange?: (sort: FilesSort) => void;
};

type SortableColumn = { sortKey: FilesSortKey; labelKey: string; className: string };

const sortableColumns: SortableColumn[] = [
    { sortKey: 'name', labelKey: 'NAME', className: styles.nameColumn! },
    { sortKey: 'size', labelKey: 'SIZE', className: styles.sizeColumn! },
    { sortKey: 'updated_at', labelKey: 'MODIFIED', className: styles.modifiedColumn! },
];

const FileListHeader = ({ sort = defaultFilesSort, onSortChange }: FileListHeaderProps) => {
    const { t } = useI18n();
    const isDescending = sort.order === 'desc';

    const requestSort = (sortKey: FilesSortKey) => {
        if (sortKey !== sort.key) {
            onSortChange?.({ key: sortKey, order: 'asc' });
            return;
        }
        onSortChange?.({ key: sortKey, order: isDescending ? 'asc' : 'desc' });
    };

    const ariaSortFor = (sortKey: FilesSortKey) => {
        if (sortKey !== sort.key) return 'none';
        return isDescending ? 'descending' : 'ascending';
    };

    return (
        <div role="row" className={styles.header}>
            <span aria-hidden="true" />
            {sortableColumns.map((column) => (
                <div
                    key={column.sortKey}
                    role="columnheader"
                    aria-sort={ariaSortFor(column.sortKey)}
                    className={column.className}
                >
                    <button
                        type="button"
                        className={styles.sortButton}
                        onClick={() => requestSort(column.sortKey)}
                    >
                        {t(column.labelKey)}
                        {column.sortKey === sort.key ? (
                            <span className={styles.sortArrow} data-testid="sort-arrow">
                                {isDescending ? (
                                    <ArrowDown size={14} aria-hidden="true" />
                                ) : (
                                    <ArrowUp size={14} aria-hidden="true" />
                                )}
                            </span>
                        ) : null}
                    </button>
                </div>
            ))}
            <div role="columnheader" className={styles.typeColumn}>
                {t('TYPE')}
            </div>
            <span aria-hidden="true" />
        </div>
    );
};

export default FileListHeader;

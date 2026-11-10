import useI18n from '@/components/i18n/provider/i18nContext';
import { Button } from '@mui/material';
import { ArchiveRestore, Trash2 } from 'lucide-react';
import styles from './TrashScreen.module.css';

type TrashSelectionToolbarProps = {
    selectedCount: number;
    isBusy: boolean;
    onSelectAll: () => void;
    onClearSelection: () => void;
    onRestoreSelected: () => void;
    onDeleteSelected: () => void;
};

const TrashSelectionToolbar = ({
    selectedCount,
    isBusy,
    onSelectAll,
    onClearSelection,
    onRestoreSelected,
    onDeleteSelected,
}: TrashSelectionToolbarProps) => {
    const { t } = useI18n();
    const hasSelection = selectedCount > 0;

    return (
        <div
            className={styles.selectionToolbar}
            role="toolbar"
            aria-label={t('TRASH_SELECTION_TOOLBAR')}
        >
            <Button size="small" onClick={onSelectAll}>
                {t('TRASH_SELECT_ALL')}
            </Button>
            {hasSelection && (
                <>
                    <span className={styles.selectionCount}>
                        {t('TRASH_SELECTION_COUNT', { count: String(selectedCount) })}
                    </span>
                    <Button size="small" onClick={onClearSelection}>
                        {t('TRASH_CLEAR_SELECTION')}
                    </Button>
                    <Button
                        size="small"
                        variant="contained"
                        disabled={isBusy}
                        onClick={onRestoreSelected}
                        startIcon={<ArchiveRestore size={16} />}
                    >
                        {t('TRASH_RESTORE_SELECTED_BUTTON')}
                    </Button>
                    <Button
                        size="small"
                        variant="outlined"
                        color="error"
                        disabled={isBusy}
                        onClick={onDeleteSelected}
                        startIcon={<Trash2 size={16} />}
                    >
                        {t('TRASH_DELETE_SELECTED_BUTTON')}
                    </Button>
                </>
            )}
        </div>
    );
};

export default TrashSelectionToolbar;

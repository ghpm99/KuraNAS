import { useState } from 'react';
import { Badge, Button, Drawer, useMediaQuery } from '@mui/material';
import { SlidersHorizontal } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { viewportMediaQueries } from '@/theme/visualTokens';
import FileSearchFilterControls from './FileSearchFilterControls';
import {
    countActiveFileSearchFilters,
    emptyFileSearchFilters,
    type FileSearchFilters,
} from './fileSearchFilters';
import styles from './fileSearch.module.css';

interface FileSearchFilterBarProps {
    filters?: FileSearchFilters;
    onChange?: (filters: FileSearchFilters) => void;
    onReset?: () => void;
}

const FileSearchFilterBar = ({
    filters = emptyFileSearchFilters,
    onChange,
    onReset,
}: FileSearchFilterBarProps) => {
    const { t } = useI18n();
    const isPhone = useMediaQuery(viewportMediaQueries.belowPhone);
    const [isDrawerOpen, setIsDrawerOpen] = useState(false);
    const activeFilterCount = countActiveFileSearchFilters(filters);
    const clearButton =
        activeFilterCount > 0 ? (
            <Button size="small" onClick={onReset}>
                {t('FILES_SEARCH_FILTERS_CLEAR')}
            </Button>
        ) : null;

    if (!isPhone) {
        return (
            <div className={styles.filterBar}>
                <FileSearchFilterControls filters={filters} onChange={onChange} />
                {clearButton}
            </div>
        );
    }

    return (
        <div className={styles.filterBar}>
            <Badge badgeContent={activeFilterCount} color="primary">
                <Button
                    variant="outlined"
                    size="small"
                    startIcon={<SlidersHorizontal size={16} />}
                    onClick={() => setIsDrawerOpen(true)}
                >
                    {t('FILES_SEARCH_FILTERS_BUTTON')}
                </Button>
            </Badge>
            <Drawer anchor="bottom" open={isDrawerOpen} onClose={() => setIsDrawerOpen(false)}>
                <div className={styles.filterDrawer}>
                    <p className={styles.filterDrawerTitle}>{t('FILES_SEARCH_FILTERS_TITLE')}</p>
                    <FileSearchFilterControls filters={filters} onChange={onChange} />
                    <div className={styles.filterDrawerActions}>
                        {clearButton}
                        <Button size="small" onClick={() => setIsDrawerOpen(false)}>
                            {t('FILES_SEARCH_FILTERS_CLOSE')}
                        </Button>
                    </div>
                </div>
            </Drawer>
        </div>
    );
};

export default FileSearchFilterBar;

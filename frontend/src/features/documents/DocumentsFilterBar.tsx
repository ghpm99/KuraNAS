import { useState } from 'react';
import { Badge, Button, Drawer, useMediaQuery } from '@mui/material';
import { SlidersHorizontal } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { viewportMediaQueries } from '@/theme/visualTokens';
import DocumentsFilterControls from './DocumentsFilterControls';
import {
    countActiveDocumentsFilters,
    emptyDocumentsFilters,
    type DocumentsFilterState,
} from './documentsFilters';
import styles from './DocumentsFilters.module.css';

interface DocumentsFilterBarProps {
    filters?: DocumentsFilterState;
    onChange?: (filters: DocumentsFilterState) => void;
    onReset?: () => void;
}

const DocumentsFilterBar = ({
    filters = emptyDocumentsFilters,
    onChange,
    onReset,
}: DocumentsFilterBarProps) => {
    const { t } = useI18n();
    const isPhone = useMediaQuery(viewportMediaQueries.belowPhone);
    const [isDrawerOpen, setIsDrawerOpen] = useState(false);
    const activeFilterCount = countActiveDocumentsFilters(filters);
    const clearButton =
        activeFilterCount > 0 ? (
            <Button size="small" onClick={onReset}>
                {t('DOCUMENTS_FILTERS_CLEAR')}
            </Button>
        ) : null;

    if (!isPhone) {
        return (
            <div className={styles.filterBar}>
                <DocumentsFilterControls filters={filters} onChange={onChange} />
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
                    {t('DOCUMENTS_FILTERS_BUTTON')}
                </Button>
            </Badge>
            <Drawer anchor="bottom" open={isDrawerOpen} onClose={() => setIsDrawerOpen(false)}>
                <div className={styles.filterDrawer}>
                    <p className={styles.filterDrawerTitle}>{t('DOCUMENTS_FILTERS_TITLE')}</p>
                    <DocumentsFilterControls filters={filters} onChange={onChange} />
                    <div className={styles.filterDrawerActions}>
                        {clearButton}
                        <Button size="small" onClick={() => setIsDrawerOpen(false)}>
                            {t('DOCUMENTS_FILTERS_CLOSE')}
                        </Button>
                    </div>
                </div>
            </Drawer>
        </div>
    );
};

export default DocumentsFilterBar;

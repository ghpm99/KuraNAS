import { Skeleton } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { DocumentsViewMode } from './documentsViewModePreference';
import styles from './DocumentsScreen.module.css';

const skeletonRowCount = 8;
const skeletonCardCount = 8;

type DocumentsListSkeletonProps = {
    viewMode?: DocumentsViewMode;
};

const DocumentsListSkeleton = ({ viewMode = 'list' }: DocumentsListSkeletonProps) => {
    const { t } = useI18n();

    if (viewMode === 'grid') {
        return (
            <div role="status" aria-label={t('LOADING')} className={styles.grid}>
                {Array.from({ length: skeletonCardCount }, (_, cardIndex) => (
                    <Skeleton key={cardIndex} variant="rounded" height={180} />
                ))}
            </div>
        );
    }

    return (
        <div role="status" aria-label={t('LOADING')}>
            {Array.from({ length: skeletonRowCount }, (_, rowIndex) => (
                <Skeleton key={rowIndex} variant="rounded" height={40} sx={{ mb: 1 }} />
            ))}
        </div>
    );
};

export default DocumentsListSkeleton;

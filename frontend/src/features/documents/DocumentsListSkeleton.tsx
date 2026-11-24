import { Skeleton } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

const skeletonRowCount = 8;

const DocumentsListSkeleton = () => {
    const { t } = useI18n();

    return (
        <div role="status" aria-label={t('LOADING')}>
            {Array.from({ length: skeletonRowCount }, (_, rowIndex) => (
                <Skeleton key={rowIndex} variant="rounded" height={40} sx={{ mb: 1 }} />
            ))}
        </div>
    );
};

export default DocumentsListSkeleton;

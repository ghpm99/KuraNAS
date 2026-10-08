import { Box, Skeleton } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './fileContent.module.css';

type FileListingSkeletonProps = {
    viewMode: 'grid' | 'list';
    placeholderCount?: number;
};

const defaultPlaceholderCount = 12;

const visuallyHiddenStyle = {
    position: 'absolute',
    width: 1,
    height: 1,
    overflow: 'hidden',
    clip: 'rect(0 0 0 0)',
    whiteSpace: 'nowrap',
} as const;

const GridPlaceholder = () => (
    <Box>
        <Skeleton variant="rounded" height={140} />
        <Skeleton variant="text" sx={{ mt: 1 }} />
        <Skeleton variant="text" width="60%" />
    </Box>
);

const ListPlaceholder = () => (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, py: 0.5 }}>
        <Skeleton variant="rounded" width={40} height={40} sx={{ flexShrink: 0 }} />
        <Skeleton variant="text" sx={{ flex: 1 }} />
        <Skeleton variant="text" width={72} />
    </Box>
);

const FileListingSkeleton = ({
    viewMode,
    placeholderCount = defaultPlaceholderCount,
}: FileListingSkeletonProps) => {
    const { t } = useI18n();
    const placeholderIndexes = Array.from({ length: placeholderCount }, (_, index) => index);
    const isGrid = viewMode === 'grid';

    return (
        <div className={styles.fileContent} role="status" aria-busy="true">
            <span style={visuallyHiddenStyle}>{t('LOADING')}</span>
            <div className={isGrid ? styles.fileGrid : styles.fileList} aria-hidden="true">
                {placeholderIndexes.map((placeholderIndex) =>
                    isGrid ? (
                        <GridPlaceholder key={placeholderIndex} />
                    ) : (
                        <ListPlaceholder key={placeholderIndex} />
                    )
                )}
            </div>
        </div>
    );
};

export default FileListingSkeleton;

import { Box, Button, CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import useInfiniteScrollSentinel from '@/components/hooks/useInfiniteScrollSentinel/useInfiniteScrollSentinel';

interface LoadMoreSentinelProps {
    hasNextPage?: boolean;
    isFetchingNextPage?: boolean;
    fetchNextPage?: () => void;
}

const doNothing = () => undefined;

const LoadMoreSentinel = ({
    hasNextPage = false,
    isFetchingNextPage = false,
    fetchNextPage = doNothing,
}: LoadMoreSentinelProps) => {
    const { t } = useI18n();
    const { sentinelRef, loadMore } = useInfiniteScrollSentinel({
        hasNextPage,
        isFetchingNextPage,
        fetchNextPage,
    });

    if (!hasNextPage) {
        return null;
    }

    return (
        <Box
            ref={sentinelRef}
            sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 1, p: 2 }}
        >
            {isFetchingNextPage ? (
                <CircularProgress size={20} role="progressbar" aria-label={t('LOADING')} />
            ) : null}
            <Button onClick={loadMore} disabled={isFetchingNextPage} size="small">
                {t('LOAD_MORE')}
            </Button>
        </Box>
    );
};

export default LoadMoreSentinel;

import { Box, CircularProgress } from '@mui/material';
import type { ReactNode } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import EmptyState from '@/components/emptyState/emptyState';
import ErrorState from '@/components/errorState/errorState';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';

type MusicCollectionListFeedbackProps = {
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    isEmpty: boolean;
    emptyTitleKey: string;
    errorTitleKey: string;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onRetry: () => void;
    fetchNextPage: () => void;
    children: ReactNode;
};

export default function MusicCollectionListFeedback({
    isLoading,
    isError,
    errorMessage,
    isEmpty,
    emptyTitleKey,
    errorTitleKey,
    hasNextPage,
    isFetchingNextPage,
    onRetry,
    fetchNextPage,
    children,
}: MusicCollectionListFeedbackProps) {
    const { t } = useI18n();

    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                <CircularProgress />
            </Box>
        );
    }

    if (isError && isEmpty) {
        return (
            <Box sx={{ p: 2 }}>
                <ErrorState
                    title={t(errorTitleKey)}
                    backendMessage={errorMessage}
                    onRetry={onRetry}
                />
            </Box>
        );
    }

    if (isEmpty) {
        return (
            <Box sx={{ p: 2 }}>
                <EmptyState title={t(emptyTitleKey)} />
            </Box>
        );
    }

    return (
        <>
            {children}
            {isError && (
                <Box sx={{ p: 2 }}>
                    <ErrorState
                        title={t(errorTitleKey)}
                        backendMessage={errorMessage}
                        onRetry={onRetry}
                    />
                </Box>
            )}
            <LoadMoreSentinel
                hasNextPage={hasNextPage && !isError}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={fetchNextPage}
            />
        </>
    );
}

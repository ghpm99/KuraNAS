import { Box, CircularProgress } from '@mui/material';
import type { ReactNode } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import EmptyState from '@/components/emptyState/emptyState';
import ListFeedback from '@/components/listFeedback/ListFeedback';

type MusicCollectionListFeedbackProps = {
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    isEmpty: boolean;
    emptyTitleKey: string;
    errorTitleKey: string;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    isNextPageError: boolean;
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
    isNextPageError,
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

    return (
        <ListFeedback
            isEmpty={isEmpty}
            isError={isError}
            isNextPageError={isNextPageError}
            errorTitle={t(errorTitleKey)}
            errorMessage={errorMessage}
            onRetry={onRetry}
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onLoadMore={fetchNextPage}
            emptyState={
                <Box sx={{ p: 2 }}>
                    <EmptyState title={t(emptyTitleKey)} />
                </Box>
            }
        >
            {children}
        </ListFeedback>
    );
}

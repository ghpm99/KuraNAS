import { Box } from '@mui/material';
import type { ReactNode } from 'react';
import ErrorState from '@/components/errorState/errorState';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';

type ListFeedbackProps = {
    isEmpty: boolean;
    isError: boolean;
    errorTitle: string;
    errorMessage?: string;
    onRetry: () => void;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onLoadMore: () => void;
    emptyState: ReactNode;
    children: ReactNode;
};

/**
 * Feedback de uma lista paginada já carregada: erro em tela cheia só quando não há itens;
 * com itens, a falha de uma página seguinte aparece abaixo da lista sem esvaziá-la.
 * Não trata o carregamento inicial. Os textos chegam prontos do chamador.
 */
export default function ListFeedback({
    isEmpty,
    isError,
    errorTitle,
    errorMessage,
    onRetry,
    hasNextPage,
    isFetchingNextPage,
    onLoadMore,
    emptyState,
    children,
}: ListFeedbackProps) {
    const errorState = (
        <Box sx={{ p: 2 }}>
            <ErrorState title={errorTitle} backendMessage={errorMessage} onRetry={onRetry} />
        </Box>
    );

    if (isError && isEmpty) {
        return errorState;
    }

    if (isEmpty) {
        return <>{emptyState}</>;
    }

    return (
        <>
            {children}
            {isError && errorState}
            <LoadMoreSentinel
                hasNextPage={hasNextPage && !isError}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={onLoadMore}
            />
        </>
    );
}

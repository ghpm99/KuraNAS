import { useCallback } from 'react';
import type { UseInfiniteQueryResult } from '@tanstack/react-query';

type RetryableInfiniteQuery = Pick<
    UseInfiniteQueryResult,
    'isFetchNextPageError' | 'fetchNextPage' | 'refetch'
>;

/**
 * Devolve `retryFailedPage`, estável entre renders: pede de novo a página que falhou
 * (`fetchNextPage` quando a falha foi ao carregar a próxima página) ou recarrega a query
 * (`refetch`) quando a primeira página falhou.
 */
export default function useRetryFailedPage({
    isFetchNextPageError,
    fetchNextPage,
    refetch,
}: RetryableInfiniteQuery) {
    return useCallback(() => {
        if (isFetchNextPageError) {
            void fetchNextPage();
            return;
        }
        void refetch();
    }, [isFetchNextPageError, fetchNextPage, refetch]);
}

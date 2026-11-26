import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { InfiniteData, QueryKey, UseInfiniteQueryResult } from '@tanstack/react-query';

type ReloadableInfiniteQuery = Pick<UseInfiniteQueryResult, 'refetch'>;

export type ReloadFromStartOutcome = {
    isReloadFailed: boolean;
    reloadError: unknown;
};

/**
 * Devolve `reloadFromStart`, estável entre renders: corta o cache da lista para a primeira
 * página, sem limpar a query (os itens atuais continuam na tela), e refaz a busca a partir
 * dela. Resolve com `{ isReloadFailed, reloadError }`; em falha, a primeira página antiga
 * fica em cache e o chamador avisa o erro.
 */
export default function useReloadInfiniteListFromStart(
    queryKey: QueryKey,
    { refetch }: ReloadableInfiniteQuery
) {
    const queryClient = useQueryClient();

    return useCallback(async (): Promise<ReloadFromStartOutcome> => {
        queryClient.setQueryData<InfiniteData<unknown>>(queryKey, (cachedList) =>
            cachedList
                ? {
                      pages: cachedList.pages.slice(0, 1),
                      pageParams: cachedList.pageParams.slice(0, 1),
                  }
                : cachedList
        );
        const refetchResult = await refetch();
        return { isReloadFailed: refetchResult.isError, reloadError: refetchResult.error };
    }, [queryClient, queryKey, refetch]);
}

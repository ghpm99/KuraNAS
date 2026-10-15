import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { setFileStarred } from '@/service/files';
import type { GlobalSearchResponse } from '@/service/search';
import { invalidateStarredViews } from '@/shared/queryKeys/invalidateStarredViews';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { applySearchStarToCache } from './applySearchStarToCache';

type StarToggleVariables = { fileId: number; wasStarred: boolean };

const searchQueryKeyRoots = [['global-search'], ['global-search-ai']] as const;

export const useSearchStarToggle = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const toggleMutation = useMutation({
        mutationFn: ({ fileId, wasStarred }: StarToggleVariables) =>
            setFileStarred(fileId, !wasStarred),
        onMutate: async ({ fileId, wasStarred }) => {
            await Promise.all(
                searchQueryKeyRoots.map((queryKey) => queryClient.cancelQueries({ queryKey }))
            );
            const snapshots = searchQueryKeyRoots.flatMap((queryKey) =>
                queryClient.getQueriesData<GlobalSearchResponse>({ queryKey })
            );
            snapshots.forEach(([cachedKey, cachedResponse]) => {
                queryClient.setQueryData(
                    cachedKey,
                    applySearchStarToCache(cachedResponse, fileId, !wasStarred)
                );
            });
            return { snapshots };
        },
        onError: (error, _variables, context) => {
            context?.snapshots.forEach(([cachedKey, cachedResponse]) => {
                queryClient.setQueryData(cachedKey, cachedResponse);
            });
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('ERROR_FAVORITE_FAILED'), {
                variant: 'error',
            });
        },
        onSettled: () => {
            void invalidateStarredViews(queryClient);
        },
    });

    return {
        toggleStar: (fileId: number, wasStarred: boolean) =>
            toggleMutation.mutate({ fileId, wasStarred }),
    };
};

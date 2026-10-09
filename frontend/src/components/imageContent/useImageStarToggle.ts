import { useMutation, useQueryClient, type InfiniteData } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { toggleStarredFile } from '@/service/files';
import type { ImageLibraryFilters, ImageLibraryPage } from '@/types/imageLibrary';

type ImageLibraryCache = InfiniteData<ImageLibraryPage, unknown>;
type StarToggleVariables = { fileId: number; wasStarred: boolean };

const libraryQueryKey = ['images', 'library'] as const;

const applyStarToCache = (
    cache: ImageLibraryCache | undefined,
    fileId: number,
    isStarred: boolean,
    shouldDropUnstarred: boolean
): ImageLibraryCache | undefined => {
    if (!cache) {
        return cache;
    }
    return {
        ...cache,
        pages: cache.pages.map((page) => ({
            ...page,
            items: (page.items ?? [])
                .map((item) => (item.file_id === fileId ? { ...item, starred: isStarred } : item))
                .filter((item) => !(shouldDropUnstarred && !item.starred)),
        })),
    };
};

export const useImageStarToggle = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const toggleMutation = useMutation({
        mutationFn: ({ fileId }: StarToggleVariables) => toggleStarredFile(fileId),
        onMutate: async ({ fileId, wasStarred }) => {
            await queryClient.cancelQueries({ queryKey: libraryQueryKey });
            const snapshots = queryClient.getQueriesData<ImageLibraryCache>({
                queryKey: libraryQueryKey,
            });
            snapshots.forEach(([queryKey, cache]) => {
                const filters = queryKey[2] as ImageLibraryFilters | undefined;
                queryClient.setQueryData(
                    queryKey,
                    applyStarToCache(cache, fileId, !wasStarred, Boolean(filters?.isStarredOnly))
                );
            });
            return { snapshots };
        },
        onSuccess: (_, { wasStarred }) => {
            enqueueSnackbar(
                wasStarred
                    ? t('IMAGES_VIEWER_FAVORITE_REMOVED')
                    : t('IMAGES_VIEWER_FAVORITE_ADDED'),
                { variant: 'success' }
            );
        },
        onError: (_error, _variables, context) => {
            context?.snapshots.forEach(([queryKey, cache]) => {
                queryClient.setQueryData(queryKey, cache);
            });
            enqueueSnackbar(t('IMAGES_VIEWER_FAVORITE_ERROR'), { variant: 'error' });
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: libraryQueryKey, refetchType: 'none' });
            queryClient.invalidateQueries({ queryKey: ['images', 'count'] });
            queryClient.invalidateQueries({ queryKey: ['images', 'timeline'] });
        },
    });

    return {
        toggleStar: (fileId: number, wasStarred: boolean) =>
            toggleMutation.mutate({ fileId, wasStarred }),
        isStarTogglePending: toggleMutation.isPending,
    };
};

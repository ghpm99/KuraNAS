import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { setFileStarred } from '@/service/files';
import { invalidateStarredViews } from '@/shared/queryKeys/invalidateStarredViews';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { applyVideoStarToCache } from './applyVideoStarToCache';

type StarToggleVariables = { videoId: number; wasStarred: boolean };

type UseVideoStarToggleOptions = {
    onStarredChange?: (videoId: number, isStarred: boolean) => void;
};

const videoQueryKeyRoots = [['video'], ['videos']] as const;

export const useVideoStarToggle = ({ onStarredChange }: UseVideoStarToggleOptions = {}) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const writeStarToVideoCaches = (videoId: number, isStarred: boolean) => {
        videoQueryKeyRoots.forEach((queryKey) => {
            queryClient.getQueriesData({ queryKey }).forEach(([cachedKey, cachedData]) => {
                if (cachedData === undefined) {
                    return;
                }
                queryClient.setQueryData(
                    cachedKey,
                    applyVideoStarToCache(cachedData, videoId, isStarred)
                );
            });
        });
    };

    const toggleMutation = useMutation({
        mutationFn: ({ videoId, wasStarred }: StarToggleVariables) =>
            setFileStarred(videoId, !wasStarred),
        onMutate: async ({ videoId, wasStarred }) => {
            await Promise.all(
                videoQueryKeyRoots.map((queryKey) => queryClient.cancelQueries({ queryKey }))
            );
            const snapshots = videoQueryKeyRoots.flatMap((queryKey) =>
                queryClient.getQueriesData({ queryKey })
            );
            writeStarToVideoCaches(videoId, !wasStarred);
            onStarredChange?.(videoId, !wasStarred);
            return { snapshots };
        },
        onError: (error, { videoId, wasStarred }, context) => {
            context?.snapshots.forEach(([cachedKey, cachedData]) => {
                queryClient.setQueryData(cachedKey, cachedData);
            });
            onStarredChange?.(videoId, wasStarred);
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('ERROR_FAVORITE_FAILED'), {
                variant: 'error',
            });
        },
        onSettled: () => {
            void invalidateStarredViews(queryClient);
        },
    });

    return {
        toggleStar: (videoId: number, wasStarred: boolean) =>
            toggleMutation.mutate({ videoId, wasStarred }),
        isStarTogglePending: toggleMutation.isPending,
    };
};

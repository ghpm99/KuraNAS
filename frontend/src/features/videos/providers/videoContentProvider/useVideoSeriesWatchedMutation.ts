import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { setVideoPlaylistWatched } from '@/service/videoPlayback';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { invalidateVideoViews } from './invalidateVideoViews';

type SeriesWatchedVariables = { watched: boolean; season?: number };

export const useVideoSeriesWatchedMutation = (playlistId: number) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const mutation = useMutation({
        mutationFn: ({ watched, season }: SeriesWatchedVariables) =>
            setVideoPlaylistWatched(playlistId, watched, season),
        onSuccess: async () => {
            await invalidateVideoViews(queryClient);
            enqueueSnackbar(t('VIDEO_SERIES_WATCHED_UPDATED'), { variant: 'success' });
        },
        onError: (error) => {
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('VIDEO_SERIES_WATCHED_ERROR'), {
                variant: 'error',
            });
        },
    });

    return {
        markWatched: (variables: SeriesWatchedVariables) => mutation.mutate(variables),
        isMarkingWatched: mutation.isPending,
    };
};

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { setVideoEpisode } from '@/service/videoPlayback';
import { extractBackendErrorField } from '@/shared/utils/extractBackendErrorField';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { invalidateVideoViews } from './invalidateVideoViews';

type EpisodeCorrection = { season: number | null; episode: number | null };

export const useVideoEpisodeCorrection = (fileId: number, onCorrected: () => void) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const mutation = useMutation({
        mutationFn: ({ season, episode }: EpisodeCorrection) =>
            setVideoEpisode(fileId, season, episode),
        onSuccess: async () => {
            await invalidateVideoViews(queryClient);
            enqueueSnackbar(t('VIDEO_CORRECTION_EPISODE_SAVED'), { variant: 'success' });
            onCorrected();
        },
    });

    return {
        correct: (correction: EpisodeCorrection) => mutation.mutate(correction),
        isCorrecting: mutation.isPending,
        errorMessage: mutation.isError
            ? (extractBackendErrorMessage(mutation.error) ?? t('VIDEO_CORRECTION_EPISODE_ERROR'))
            : undefined,
        errorField: mutation.isError ? extractBackendErrorField(mutation.error) : undefined,
        resetError: mutation.reset,
    };
};

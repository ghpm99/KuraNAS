import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { setVideoClassification, type VideoManualClassification } from '@/service/videoPlayback';
import { extractBackendErrorField } from '@/shared/utils/extractBackendErrorField';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { invalidateVideoViews } from './invalidateVideoViews';

export const useVideoClassificationCorrection = (fileId: number, onCorrected: () => void) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const mutation = useMutation({
        mutationFn: (classification: VideoManualClassification | null) =>
            setVideoClassification(fileId, classification),
        onSuccess: async () => {
            await invalidateVideoViews(queryClient);
            enqueueSnackbar(t('VIDEO_CORRECTION_CLASSIFICATION_SAVED'), { variant: 'success' });
            onCorrected();
        },
    });

    return {
        correct: (classification: VideoManualClassification | null) =>
            mutation.mutate(classification),
        isCorrecting: mutation.isPending,
        errorMessage: mutation.isError
            ? (extractBackendErrorMessage(mutation.error) ??
              t('VIDEO_CORRECTION_CLASSIFICATION_ERROR'))
            : undefined,
        errorField: mutation.isError ? extractBackendErrorField(mutation.error) : undefined,
        resetError: mutation.reset,
    };
};

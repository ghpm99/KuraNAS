import useI18n from '@/components/i18n/provider/i18nContext';
import {
    getPendingImageClassificationCount,
    startImageClassificationBackfill,
} from '@/service/files';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';

const extractBackendError = (error: unknown): string | undefined => {
    if (typeof error === 'object' && error !== null && 'response' in error) {
        const response = (error as { response?: { data?: { error?: string } } }).response;
        return response?.data?.error;
    }
    return undefined;
};

const useImageClassificationBackfill = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const pendingQuery = useQuery({
        queryKey: ['image-classification-pending-count'],
        queryFn: getPendingImageClassificationCount,
        retry: false,
    });

    const backfillMutation = useMutation({
        mutationFn: startImageClassificationBackfill,
        onSuccess: () => {
            enqueueSnackbar(t('IMAGE_CLASSIFY_BACKFILL_TOAST_SUCCESS'), { variant: 'success' });
            queryClient.invalidateQueries({
                queryKey: ['image-classification-pending-count'],
            });
        },
        onError: (error: unknown) => {
            enqueueSnackbar(
                extractBackendError(error) ?? t('IMAGE_CLASSIFY_BACKFILL_TOAST_ERROR'),
                { variant: 'error' }
            );
        },
    });

    return {
        t,
        pendingCount: pendingQuery.data ?? 0,
        isLoading: pendingQuery.isLoading,
        hasError: pendingQuery.isError,
        isStarting: backfillMutation.isPending,
        startBackfill: () => backfillMutation.mutate(),
    };
};

export default useImageClassificationBackfill;

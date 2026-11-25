import useI18n from '@/components/i18n/provider/i18nContext';
import { captureThumbnailHref, deleteCapture, getCaptures } from '@/service/captures';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { formatSize } from '@/shared/utils/formatSize';
import type { Capture, CaptureStatus } from '@/types/captures';
import type { Pagination } from '@/types/pagination';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useMemo, useState } from 'react';

export const captureStatusLabelKey = (status: CaptureStatus): string => {
    switch (status) {
        case 'uploaded':
            return 'CAPTURES_STATUS_UPLOADED';
        case 'promoting':
            return 'CAPTURES_STATUS_PROMOTING';
        case 'promoted':
            return 'CAPTURES_STATUS_PROMOTED';
        case 'failed':
            return 'CAPTURES_STATUS_FAILED';
        default:
            return 'CAPTURES_STATUS_UPLOADED';
    }
};

export const useCapturesScreen = () => {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const { enqueueSnackbar } = useSnackbar();
    const [deletingId, setDeletingId] = useState<number | null>(null);
    const [captureAwaitingDeletion, setCaptureAwaitingDeletion] = useState<Capture | null>(null);

    const { data, isLoading, isError, error, refetch } = useQuery<Pagination<Capture>>({
        queryKey: ['captures'],
        queryFn: () => getCaptures(),
        refetchOnWindowFocus: false,
    });

    const removeMutation = useMutation({
        mutationFn: (id: number) => deleteCapture(id),
        onMutate: (id: number) => setDeletingId(id),
        onSuccess: (deletion) => {
            const messageKey = deletion?.library_file_kept
                ? 'CAPTURE_DELETED_LIBRARY_FILE_KEPT'
                : 'CAPTURE_DELETED';
            enqueueSnackbar(t(messageKey), { variant: 'success' });
        },
        onError: (deletionError) => {
            enqueueSnackbar(extractBackendErrorMessage(deletionError) ?? t('CAPTURE_DELETE_ERROR'), {
                variant: 'error',
            });
        },
        onSettled: () => {
            setDeletingId(null);
            queryClient.invalidateQueries({ queryKey: ['captures'] });
        },
    });

    const items = useMemo(() => data?.items ?? [], [data]);

    const confirmRemoval = () => {
        if (!captureAwaitingDeletion) return;
        removeMutation.mutate(captureAwaitingDeletion.id);
        setCaptureAwaitingDeletion(null);
    };

    return {
        t,
        items,
        isLoading,
        isError,
        errorMessage: extractBackendErrorMessage(error) ?? t('CONNECTION_BANNER_UNAVAILABLE'),
        retryLoading: () => {
            refetch();
        },
        isEmpty: !isLoading && !isError && items.length === 0,
        deletingId,
        captureAwaitingDeletion,
        requestRemoval: setCaptureAwaitingDeletion,
        cancelRemoval: () => setCaptureAwaitingDeletion(null),
        confirmRemoval,
        formatSize,
        captureThumbnailHref,
        captureStatusLabelKey,
    };
};

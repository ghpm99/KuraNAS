import useI18n from '@/components/i18n/provider/i18nContext';
import { captureThumbnailHref, deleteCapture, getCaptures } from '@/service/captures';
import { formatSize } from '@/shared/utils/formatSize';
import type { Capture, CaptureStatus } from '@/types/captures';
import type { Pagination } from '@/types/pagination';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
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
    const [deletingId, setDeletingId] = useState<number | null>(null);

    const { data, isLoading, isError } = useQuery<Pagination<Capture>>({
        queryKey: ['captures'],
        queryFn: () => getCaptures(),
        refetchOnWindowFocus: false,
    });

    const removeMutation = useMutation({
        mutationFn: (id: number) => deleteCapture(id),
        onMutate: (id: number) => setDeletingId(id),
        onSettled: () => {
            setDeletingId(null);
            queryClient.invalidateQueries({ queryKey: ['captures'] });
        },
    });

    const items = useMemo(() => data?.items ?? [], [data]);

    return {
        t,
        items,
        isLoading,
        isError,
        isEmpty: !isLoading && !isError && items.length === 0,
        deletingId,
        removeCapture: (id: number) => removeMutation.mutate(id),
        formatSize,
        captureThumbnailHref,
        captureStatusLabelKey,
    };
};

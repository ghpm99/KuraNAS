import useI18n from '@/components/i18n/provider/i18nContext';
import {
    deleteTrashItem,
    emptyTrash,
    getTrashItems,
    getTrashRetention,
    restoreTrashItem,
    updateTrashRetention,
} from '@/service/trash';
import type { TrashPage } from '@/types/trash';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useState } from 'react';

const PAGE_SIZE = 15;

const extractBackendError = (error: unknown): string | undefined => {
    if (typeof error === 'object' && error !== null && 'response' in error) {
        const response = (error as { response?: { data?: { error?: string } } }).response;
        return response?.data?.error;
    }
    return undefined;
};

export const useTrashScreen = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();
    const [page, setPage] = useState(1);

    const trashQuery = useQuery<TrashPage>({
        queryKey: ['trash', page],
        queryFn: () => getTrashItems(page, PAGE_SIZE),
        retry: false,
    });

    const retentionQuery = useQuery({
        queryKey: ['trash', 'retention'],
        queryFn: getTrashRetention,
        retry: false,
    });

    const invalidate = useCallback(() => {
        void queryClient.invalidateQueries({ queryKey: ['trash'] });
        // Restored entries reappear in the files tree right away.
        void queryClient.invalidateQueries({ queryKey: ['files'] });
    }, [queryClient]);

    const notifyError = useCallback(
        (error: unknown) => {
            // Backend errors arrive already translated — show them verbatim.
            const backendMessage = extractBackendError(error);
            enqueueSnackbar(backendMessage ?? t('TRASH_LOAD_ERROR'), { variant: 'error' });
        },
        [enqueueSnackbar, t]
    );

    const restoreMutation = useMutation({ mutationFn: restoreTrashItem, onSuccess: invalidate });
    const deleteMutation = useMutation({ mutationFn: deleteTrashItem, onSuccess: invalidate });
    const emptyMutation = useMutation({ mutationFn: emptyTrash, onSuccess: invalidate });
    const retentionMutation = useMutation({
        mutationFn: updateTrashRetention,
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: ['trash', 'retention'] });
        },
    });

    const handleRestore = useCallback(
        async (id: number) => {
            try {
                await restoreMutation.mutateAsync(id);
                enqueueSnackbar(t('TRASH_RESTORED_TOAST'), { variant: 'success' });
            } catch (error) {
                notifyError(error);
            }
        },
        [enqueueSnackbar, notifyError, restoreMutation, t]
    );

    const handleDeleteForever = useCallback(
        async (id: number) => {
            if (!window.confirm(t('TRASH_DELETE_CONFIRM'))) {
                return;
            }
            try {
                await deleteMutation.mutateAsync(id);
                enqueueSnackbar(t('TRASH_DELETED_TOAST'), { variant: 'success' });
            } catch (error) {
                notifyError(error);
            }
        },
        [deleteMutation, enqueueSnackbar, notifyError, t]
    );

    const handleEmptyTrash = useCallback(async () => {
        if (!window.confirm(t('TRASH_EMPTY_CONFIRM'))) {
            return;
        }
        try {
            await emptyMutation.mutateAsync();
            enqueueSnackbar(t('TRASH_EMPTIED_TOAST'), { variant: 'success' });
        } catch (error) {
            notifyError(error);
        }
    }, [emptyMutation, enqueueSnackbar, notifyError, t]);

    const handleRetentionChange = useCallback(
        async (days: number) => {
            if (!Number.isFinite(days) || days <= 0) {
                return;
            }
            try {
                await retentionMutation.mutateAsync(days);
                enqueueSnackbar(t('TRASH_RETENTION_SAVED'), { variant: 'success' });
            } catch (error) {
                notifyError(error);
            }
        },
        [enqueueSnackbar, notifyError, retentionMutation, t]
    );

    const items = trashQuery.data?.items ?? [];
    const pagination = trashQuery.data?.pagination;

    return {
        t,
        items,
        page,
        hasNext: pagination?.has_next ?? false,
        hasPrev: pagination?.has_prev ?? false,
        goToNextPage: () => setPage((current) => current + 1),
        goToPrevPage: () => setPage((current) => Math.max(1, current - 1)),
        isLoading: trashQuery.isLoading,
        isError: trashQuery.isError,
        isEmpty: !trashQuery.isLoading && !trashQuery.isError && items.length === 0,
        isMutating:
            restoreMutation.isPending || deleteMutation.isPending || emptyMutation.isPending,
        retentionDays: retentionQuery.data?.days,
        handleRestore,
        handleDeleteForever,
        handleEmptyTrash,
        handleRetentionChange,
    };
};

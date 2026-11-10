import useI18n from '@/components/i18n/provider/i18nContext';
import {
    deleteTrashItem,
    emptyTrash,
    getTrashItems,
    getTrashRetention,
    restoreTrashItem,
    updateTrashRetention,
} from '@/service/trash';
import { trashQueryKey } from '@/shared/queryKeys/fileQueryKeys';
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
        queryKey: [trashQueryKey, page],
        queryFn: () => getTrashItems(page, PAGE_SIZE),
        retry: false,
    });

    const retentionQuery = useQuery({
        queryKey: [trashQueryKey, 'retention'],
        queryFn: getTrashRetention,
        retry: false,
    });

    const items = trashQuery.data?.items ?? [];
    const pagination = trashQuery.data?.pagination;

    const leavePageEmptiedBy = useCallback(
        (removedItemCount: number) => {
            const isPageEmptied = removedItemCount >= items.length;
            if (isPageEmptied) {
                setPage((current) => Math.max(1, current - 1));
            }
        },
        [items.length]
    );

    const refreshTrashAndFilesTree = useCallback(() => {
        void queryClient.invalidateQueries({ queryKey: [trashQueryKey] });
        void queryClient.invalidateQueries({ queryKey: ['files'] });
    }, [queryClient]);

    const notifyActionError = useCallback(
        (error: unknown) => {
            const translatedBackendMessage = extractBackendError(error);
            enqueueSnackbar(translatedBackendMessage ?? t('TRASH_ACTION_ERROR'), {
                variant: 'error',
            });
        },
        [enqueueSnackbar, t]
    );

    const restoreMutation = useMutation({
        mutationFn: restoreTrashItem,
        onSuccess: refreshTrashAndFilesTree,
    });
    const deleteMutation = useMutation({
        mutationFn: deleteTrashItem,
        onSuccess: refreshTrashAndFilesTree,
    });
    const emptyMutation = useMutation({
        mutationFn: emptyTrash,
        onSuccess: refreshTrashAndFilesTree,
    });
    const retentionMutation = useMutation({
        mutationFn: updateTrashRetention,
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: [trashQueryKey, 'retention'] });
        },
    });

    const handleRestore = useCallback(
        async (id: number) => {
            try {
                await restoreMutation.mutateAsync(id);
                leavePageEmptiedBy(1);
                enqueueSnackbar(t('TRASH_RESTORED_TOAST'), { variant: 'success' });
            } catch (error) {
                notifyActionError(error);
            }
        },
        [enqueueSnackbar, leavePageEmptiedBy, notifyActionError, restoreMutation, t]
    );

    const handleDeleteForever = useCallback(
        async (id: number) => {
            if (!window.confirm(t('TRASH_DELETE_CONFIRM'))) {
                return;
            }
            try {
                await deleteMutation.mutateAsync(id);
                leavePageEmptiedBy(1);
                enqueueSnackbar(t('TRASH_DELETED_TOAST'), { variant: 'success' });
            } catch (error) {
                notifyActionError(error);
            }
        },
        [deleteMutation, enqueueSnackbar, leavePageEmptiedBy, notifyActionError, t]
    );

    const handleEmptyTrash = useCallback(async () => {
        if (!window.confirm(t('TRASH_EMPTY_CONFIRM'))) {
            return;
        }
        try {
            await emptyMutation.mutateAsync();
            setPage(1);
            enqueueSnackbar(t('TRASH_EMPTIED_TOAST'), { variant: 'success' });
        } catch (error) {
            notifyActionError(error);
        }
    }, [emptyMutation, enqueueSnackbar, notifyActionError, t]);

    const handleRetentionChange = useCallback(
        async (days: number) => {
            if (!Number.isFinite(days) || days <= 0) {
                return;
            }
            try {
                await retentionMutation.mutateAsync(days);
                enqueueSnackbar(t('TRASH_RETENTION_SAVED'), { variant: 'success' });
            } catch (error) {
                notifyActionError(error);
            }
        },
        [enqueueSnackbar, notifyActionError, retentionMutation, t]
    );

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
        isEmpty: trashQuery.isSuccess && items.length === 0,
        isMutating:
            restoreMutation.isPending || deleteMutation.isPending || emptyMutation.isPending,
        retentionDays: retentionQuery.data?.days,
        handleRestore,
        handleDeleteForever,
        handleEmptyTrash,
        handleRetentionChange,
    };
};

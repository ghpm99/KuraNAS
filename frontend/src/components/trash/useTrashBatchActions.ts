import useI18n from '@/components/i18n/provider/i18nContext';
import { deleteTrashItems, restoreTrashItems } from '@/service/trash';
import { useBulkOutcomeNotifier, type OutcomeMessages } from '@/shared/bulk/useBulkOutcomeNotifier';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { TrashBatchFailure, TrashItem } from '@/types/trash';
import { useMutation } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback } from 'react';

type BatchOutcome = {
    succeededIds: number[];
    failures: TrashBatchFailure[];
};

type TrashBatchActionsOptions = {
    onItemsLeftTrash: (departedItemIds: number[]) => void;
};

const restoreMessages: OutcomeMessages = {
    singleSuccessKey: 'TRASH_RESTORED_TOAST',
    multipleSuccessKey: 'TRASH_BULK_RESTORED_TOAST',
    failureKey: 'TRASH_ACTION_ERROR',
};

const deleteMessages: OutcomeMessages = {
    singleSuccessKey: 'TRASH_DELETED_TOAST',
    multipleSuccessKey: 'TRASH_BULK_DELETED_TOAST',
    failureKey: 'TRASH_ACTION_ERROR',
};

const restoreInBatch = async (itemIds: number[]): Promise<BatchOutcome> => {
    const restoreOutcome = await restoreTrashItems(itemIds);
    return {
        succeededIds: restoreOutcome?.restored_ids ?? [],
        failures: restoreOutcome?.failures ?? [],
    };
};

const deleteInBatch = async (itemIds: number[]): Promise<BatchOutcome> => {
    const deleteOutcome = await deleteTrashItems(itemIds);
    return {
        succeededIds: deleteOutcome?.deleted_ids ?? [],
        failures: deleteOutcome?.failures ?? [],
    };
};

export const useTrashBatchActions = ({ onItemsLeftTrash }: TrashBatchActionsOptions) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const notifyBulkOutcome = useBulkOutcomeNotifier();

    const restoreBatchMutation = useMutation({ mutationFn: restoreInBatch });
    const deleteBatchMutation = useMutation({ mutationFn: deleteInBatch });

    const runBatch = useCallback(
        async (
            selectedItems: TrashItem[],
            sendBatch: (itemIds: number[]) => Promise<BatchOutcome>,
            messages: OutcomeMessages
        ) => {
            if (selectedItems.length === 0) {
                return;
            }
            try {
                const batchOutcome = await sendBatch(selectedItems.map((item) => item.id));
                notifyBulkOutcome(
                    {
                        succeededCount: batchOutcome.succeededIds.length,
                        failedCount: batchOutcome.failures.length,
                        firstFailureMessage: batchOutcome.failures[0]?.error || undefined,
                    },
                    messages
                );
                onItemsLeftTrash(batchOutcome.succeededIds);
            } catch (error) {
                enqueueSnackbar(extractBackendErrorMessage(error) ?? t('TRASH_ACTION_ERROR'), {
                    variant: 'error',
                });
            }
        },
        [enqueueSnackbar, notifyBulkOutcome, onItemsLeftTrash, t]
    );

    const restoreSelected = useCallback(
        (selectedItems: TrashItem[]) =>
            runBatch(selectedItems, restoreBatchMutation.mutateAsync, restoreMessages),
        [restoreBatchMutation.mutateAsync, runBatch]
    );

    const deleteSelectedForever = useCallback(
        (selectedItems: TrashItem[]) =>
            runBatch(selectedItems, deleteBatchMutation.mutateAsync, deleteMessages),
        [deleteBatchMutation.mutateAsync, runBatch]
    );

    return {
        restoreSelected,
        deleteSelectedForever,
        isBatchRunning: restoreBatchMutation.isPending || deleteBatchMutation.isPending,
    };
};

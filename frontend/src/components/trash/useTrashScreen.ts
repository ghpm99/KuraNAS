import useI18n from '@/components/i18n/provider/i18nContext';
import { deleteTrashItem, emptyTrash, getTrashItems, restoreTrashItem } from '@/service/trash';
import { trashQueryKey } from '@/shared/queryKeys/fileQueryKeys';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';
import { useItemSelection } from '@/shared/selection/useItemSelection';
import type {
    TrashItem,
    TrashPage,
    TrashRestoreResult,
    TrashSortKey,
    TrashSortOrder,
} from '@/types/trash';
import { useTrashBatchActions } from './useTrashBatchActions';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { createElement, useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { buildFilesUrl } from '@/app/routes';
import TrashRestoredToastAction from './TrashRestoredToastAction';
import { parentFolderOf } from './trashItemPresentation';

const PAGE_SIZE = 15;
const SEARCH_DEBOUNCE_MS = 300;

export type TrashSort = { sortKey: TrashSortKey; order: TrashSortOrder };

const newestDeletionsFirst: TrashSort = { sortKey: 'deleted_at', order: 'desc' };

const extractBackendError = (error: unknown): string | undefined => {
    if (typeof error === 'object' && error !== null && 'response' in error) {
        const response = (error as { response?: { data?: { error?: string } } }).response;
        return response?.data?.error;
    }
    return undefined;
};

type PendingConfirmation =
    | { kind: 'deleteForever'; itemId: number }
    | { kind: 'deleteSelected' }
    | { kind: 'emptyTrash' };

type ConfirmationText = {
    titleKey: string;
    messageKey: string;
    messageVariables?: Record<string, string>;
};

const explorerFolderOf = (restoredPath: string): string => {
    const restoredFolder = parentFolderOf(restoredPath);
    return restoredFolder === '/' ? '' : restoredFolder;
};

const readTrashItemId = (item: TrashItem) => item.id;

const describeConfirmation = (
    pendingConfirmation: PendingConfirmation | null,
    selectedCount: number
): ConfirmationText => {
    if (pendingConfirmation?.kind === 'emptyTrash') {
        return { titleKey: 'TRASH_EMPTY_BUTTON', messageKey: 'TRASH_EMPTY_CONFIRM' };
    }
    if (pendingConfirmation?.kind === 'deleteSelected') {
        return {
            titleKey: 'TRASH_DELETE_SELECTED_BUTTON',
            messageKey: 'TRASH_DELETE_SELECTED_CONFIRM',
            messageVariables: { count: String(selectedCount) },
        };
    }
    return { titleKey: 'TRASH_DELETE_FOREVER_BUTTON', messageKey: 'TRASH_DELETE_CONFIRM' };
};

export const useTrashScreen = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();
    const navigate = useNavigate();
    const [page, setPage] = useState(1);
    const [pendingConfirmation, setPendingConfirmation] = useState<PendingConfirmation | null>(
        null
    );

    const [sort, setSort] = useState<TrashSort>(newestDeletionsFirst);
    const [searchInput, setSearchInput] = useState('');
    const searchText = useDebouncedValue(searchInput, SEARCH_DEBOUNCE_MS).trim();

    const trashQuery = useQuery<TrashPage>({
        queryKey: [trashQueryKey, page, sort.sortKey, sort.order, searchText],
        queryFn: () => getTrashItems(page, PAGE_SIZE, { ...sort, searchText }),
        retry: false,
    });

    const changeSort = useCallback((requestedSort: TrashSort) => {
        setSort(requestedSort);
        setPage(1);
    }, []);

    const changeSearchInput = useCallback((typedText: string) => {
        setSearchInput(typedText);
        setPage(1);
    }, []);

    const items = trashQuery.data?.items ?? [];
    const pagination = trashQuery.data?.pagination;

    const isPastLastPage =
        page > 1 && trashQuery.isSuccess && !trashQuery.isFetching && items.length === 0;
    if (isPastLastPage) {
        setPage(page - 1);
    }

    const refreshTrashAndFilesTree = useCallback(() => {
        void queryClient.invalidateQueries({ queryKey: [trashQueryKey] });
        void queryClient.invalidateQueries({ queryKey: ['files'] });
    }, [queryClient]);

    const selection = useItemSelection<TrashItem>(
        [page, sort.sortKey, sort.order, searchText].join('|'),
        readTrashItemId
    );
    const { deselect: deselectItems, selectedItems } = selection;

    const leaveTrashBehind = useCallback(
        (departedItemIds: number[]) => {
            const departedIds = new Set(departedItemIds);
            deselectItems(selectedItems.filter((selectedItem) => departedIds.has(selectedItem.id)));
            refreshTrashAndFilesTree();
        },
        [deselectItems, refreshTrashAndFilesTree, selectedItems]
    );

    const { restoreSelected, deleteSelectedForever, isBatchRunning } = useTrashBatchActions({
        onItemsLeftTrash: leaveTrashBehind,
    });

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
    const [busyItemIds, setBusyItemIds] = useState<ReadonlySet<number>>(new Set());

    const runItemAction = useCallback(
        async <TActionResult>(
            itemId: number,
            sendItemAction: (itemId: number) => Promise<TActionResult>,
            announceSuccess: (actionResult: TActionResult) => void
        ) => {
            setBusyItemIds((currentBusyIds) => new Set(currentBusyIds).add(itemId));
            try {
                announceSuccess(await sendItemAction(itemId));
            } catch (error) {
                notifyActionError(error);
            } finally {
                setBusyItemIds((currentBusyIds) => {
                    const remainingBusyIds = new Set(currentBusyIds);
                    remainingBusyIds.delete(itemId);
                    return remainingBusyIds;
                });
            }
        },
        [notifyActionError]
    );

    const { mutateAsync: sendRestore } = restoreMutation;
    const { mutateAsync: sendDeleteForever } = deleteMutation;

    const announceRestore = useCallback(
        (restoreResult: TrashRestoreResult | undefined) => {
            const restoredPath = restoreResult?.path;
            if (!restoredPath) {
                enqueueSnackbar(t('TRASH_RESTORED_TOAST'), { variant: 'success' });
                return;
            }
            const restoredFolderUrl = buildFilesUrl(explorerFolderOf(restoredPath));
            enqueueSnackbar(t('TRASH_RESTORED_TOAST'), {
                variant: 'success',
                action: () =>
                    createElement(TrashRestoredToastAction, {
                        onViewInFolder: () => navigate(restoredFolderUrl),
                    }),
            });
        },
        [enqueueSnackbar, navigate, t]
    );

    const handleRestore = useCallback(
        (itemId: number) => runItemAction(itemId, sendRestore, announceRestore),
        [announceRestore, runItemAction, sendRestore]
    );

    const deleteForever = useCallback(
        (itemId: number) =>
            runItemAction(itemId, sendDeleteForever, () =>
                enqueueSnackbar(t('TRASH_DELETED_TOAST'), { variant: 'success' })
            ),
        [enqueueSnackbar, runItemAction, sendDeleteForever, t]
    );

    const emptyWholeTrash = useCallback(async () => {
        try {
            await emptyMutation.mutateAsync();
            setPage(1);
            enqueueSnackbar(t('TRASH_EMPTIED_TOAST'), { variant: 'success' });
        } catch (error) {
            notifyActionError(error);
        }
    }, [emptyMutation, enqueueSnackbar, notifyActionError, t]);

    const confirmPendingAction = useCallback(async () => {
        const confirmedAction = pendingConfirmation;
        setPendingConfirmation(null);
        if (confirmedAction?.kind === 'deleteForever') {
            await deleteForever(confirmedAction.itemId);
        }
        if (confirmedAction?.kind === 'deleteSelected') {
            await deleteSelectedForever(selectedItems);
        }
        if (confirmedAction?.kind === 'emptyTrash') {
            await emptyWholeTrash();
        }
    }, [deleteForever, deleteSelectedForever, emptyWholeTrash, pendingConfirmation, selectedItems]);

    const isWholeTrashBusy = emptyMutation.isPending;

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
        isSearching: searchText !== '',
        sort,
        changeSort,
        searchInput,
        changeSearchInput,
        isItemBusy: (itemId: number) =>
            isWholeTrashBusy ||
            busyItemIds.has(itemId) ||
            (isBatchRunning && selection.isSelected(itemId)),
        isWholeTrashBusy,
        isAnyActionRunning: isWholeTrashBusy || isBatchRunning || busyItemIds.size > 0,
        isBatchRunning,
        selectedCount: selection.selectedCount,
        isItemSelected: selection.isSelected,
        toggleItemSelection: selection.toggle,
        selectWholePage: () => selection.selectAll(items),
        clearSelection: selection.clear,
        restoreSelectedItems: () => void restoreSelected(selectedItems),
        requestDeleteSelected: () => setPendingConfirmation({ kind: 'deleteSelected' }),
        confirmationText: describeConfirmation(pendingConfirmation, selection.selectedCount),
        handleRestore,
        pendingConfirmation,
        requestDeleteForever: (itemId: number) =>
            setPendingConfirmation({ kind: 'deleteForever', itemId }),
        requestEmptyTrash: () => setPendingConfirmation({ kind: 'emptyTrash' }),
        cancelPendingAction: () => setPendingConfirmation(null),
        confirmPendingAction,
    };
};

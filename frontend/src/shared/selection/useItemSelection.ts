import { useCallback, useMemo, useState } from 'react';

type SelectionState<TItem> = {
    scopeKey: string;
    selectedItems: TItem[];
    anchorItemId: number | null;
};

export type ItemSelection<TItem> = {
    selectedItems: TItem[];
    selectedCount: number;
    hasSelection: boolean;
    isSelected: (itemId: number) => boolean;
    toggle: (item: TItem) => void;
    selectRange: (targetItem: TItem, orderedItems: TItem[]) => void;
    selectAll: (items: TItem[]) => void;
    select: (items: TItem[]) => void;
    deselect: (items: TItem[]) => void;
    clear: () => void;
};

const createEmptyState = <TItem>(scopeKey: string): SelectionState<TItem> => ({
    scopeKey,
    selectedItems: [],
    anchorItemId: null,
});

export const useItemSelection = <TItem>(
    scopeKey: string,
    getItemId: (item: TItem) => number
): ItemSelection<TItem> => {
    const [storedState, setStoredState] = useState<SelectionState<TItem>>(() =>
        createEmptyState<TItem>(scopeKey)
    );

    const state = storedState.scopeKey === scopeKey ? storedState : createEmptyState<TItem>(scopeKey);

    const updateState = useCallback(
        (update: (currentState: SelectionState<TItem>) => SelectionState<TItem>) => {
            setStoredState((previousState) =>
                update(
                    previousState.scopeKey === scopeKey
                        ? previousState
                        : createEmptyState<TItem>(scopeKey)
                )
            );
        },
        [scopeKey]
    );

    const toggle = useCallback(
        (item: TItem) => {
            const itemId = getItemId(item);
            updateState((currentState) => {
                const isAlreadySelected = currentState.selectedItems.some(
                    (selectedItem) => getItemId(selectedItem) === itemId
                );
                return {
                    ...currentState,
                    selectedItems: isAlreadySelected
                        ? currentState.selectedItems.filter(
                              (selectedItem) => getItemId(selectedItem) !== itemId
                          )
                        : [...currentState.selectedItems, item],
                    anchorItemId: itemId,
                };
            });
        },
        [updateState, getItemId]
    );

    const selectRange = useCallback(
        (targetItem: TItem, orderedItems: TItem[]) => {
            const targetId = getItemId(targetItem);
            updateState((currentState) => {
                const anchorIndex = orderedItems.findIndex(
                    (orderedItem) => getItemId(orderedItem) === currentState.anchorItemId
                );
                const targetIndex = orderedItems.findIndex(
                    (orderedItem) => getItemId(orderedItem) === targetId
                );
                const alreadySelectedIds = new Set(currentState.selectedItems.map(getItemId));
                if (anchorIndex === -1 || targetIndex === -1) {
                    return {
                        ...currentState,
                        selectedItems: alreadySelectedIds.has(targetId)
                            ? currentState.selectedItems
                            : [...currentState.selectedItems, targetItem],
                        anchorItemId: targetId,
                    };
                }
                const rangeItems = orderedItems.slice(
                    Math.min(anchorIndex, targetIndex),
                    Math.max(anchorIndex, targetIndex) + 1
                );
                return {
                    ...currentState,
                    selectedItems: [
                        ...currentState.selectedItems,
                        ...rangeItems.filter((rangeItem) => !alreadySelectedIds.has(getItemId(rangeItem))),
                    ],
                };
            });
        },
        [updateState, getItemId]
    );

    const selectAll = useCallback(
        (items: TItem[]) => {
            updateState((currentState) => ({ ...currentState, selectedItems: [...items] }));
        },
        [updateState]
    );

    const select = useCallback(
        (items: TItem[]) => {
            updateState((currentState) => {
                const alreadySelectedIds = new Set(currentState.selectedItems.map(getItemId));
                return {
                    ...currentState,
                    selectedItems: [
                        ...currentState.selectedItems,
                        ...items.filter((item) => !alreadySelectedIds.has(getItemId(item))),
                    ],
                };
            });
        },
        [updateState, getItemId]
    );

    const deselect = useCallback(
        (items: TItem[]) => {
            const deselectedIds = new Set(items.map(getItemId));
            updateState((currentState) => ({
                ...currentState,
                selectedItems: currentState.selectedItems.filter(
                    (selectedItem) => !deselectedIds.has(getItemId(selectedItem))
                ),
            }));
        },
        [updateState, getItemId]
    );

    const clear = useCallback(() => {
        updateState((currentState) => createEmptyState<TItem>(currentState.scopeKey));
    }, [updateState]);

    const selectedIds = useMemo(
        () => new Set(state.selectedItems.map(getItemId)),
        [state.selectedItems, getItemId]
    );

    const isSelected = useCallback((itemId: number) => selectedIds.has(itemId), [selectedIds]);

    return {
        selectedItems: state.selectedItems,
        selectedCount: state.selectedItems.length,
        hasSelection: state.selectedItems.length > 0,
        isSelected,
        toggle,
        selectRange,
        selectAll,
        select,
        deselect,
        clear,
    };
};

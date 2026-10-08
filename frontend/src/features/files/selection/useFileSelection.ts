import { useCallback, useMemo, useState } from 'react';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

type SelectionState = {
    scopeKey: string;
    selectedFiles: FileData[];
    anchorFileId: number | null;
};

export type FileSelection = {
    selectedFiles: FileData[];
    selectedCount: number;
    hasSelection: boolean;
    isSelected: (fileId: number) => boolean;
    toggle: (file: FileData) => void;
    selectRange: (targetFile: FileData, orderedFiles: FileData[]) => void;
    selectAll: (files: FileData[]) => void;
    deselect: (files: FileData[]) => void;
    clear: () => void;
};

const createEmptyState = (scopeKey: string): SelectionState => ({
    scopeKey,
    selectedFiles: [],
    anchorFileId: null,
});

export const useFileSelection = (scopeKey: string): FileSelection => {
    const [storedState, setStoredState] = useState<SelectionState>(() => createEmptyState(scopeKey));

    const state = storedState.scopeKey === scopeKey ? storedState : createEmptyState(scopeKey);

    const updateState = useCallback(
        (update: (currentState: SelectionState) => SelectionState) => {
            setStoredState((previousState) =>
                update(
                    previousState.scopeKey === scopeKey
                        ? previousState
                        : createEmptyState(scopeKey)
                )
            );
        },
        [scopeKey]
    );

    const toggle = useCallback(
        (file: FileData) => {
            updateState((currentState) => {
                const isAlreadySelected = currentState.selectedFiles.some(
                    (selectedFile) => selectedFile.id === file.id
                );
                return {
                    ...currentState,
                    selectedFiles: isAlreadySelected
                        ? currentState.selectedFiles.filter(
                              (selectedFile) => selectedFile.id !== file.id
                          )
                        : [...currentState.selectedFiles, file],
                    anchorFileId: file.id,
                };
            });
        },
        [updateState]
    );

    const selectRange = useCallback(
        (targetFile: FileData, orderedFiles: FileData[]) => {
            updateState((currentState) => {
                const anchorIndex = orderedFiles.findIndex(
                    (orderedFile) => orderedFile.id === currentState.anchorFileId
                );
                const targetIndex = orderedFiles.findIndex(
                    (orderedFile) => orderedFile.id === targetFile.id
                );
                if (anchorIndex === -1 || targetIndex === -1) {
                    const isAlreadySelected = currentState.selectedFiles.some(
                        (selectedFile) => selectedFile.id === targetFile.id
                    );
                    return {
                        ...currentState,
                        selectedFiles: isAlreadySelected
                            ? currentState.selectedFiles
                            : [...currentState.selectedFiles, targetFile],
                        anchorFileId: targetFile.id,
                    };
                }
                const rangeFiles = orderedFiles.slice(
                    Math.min(anchorIndex, targetIndex),
                    Math.max(anchorIndex, targetIndex) + 1
                );
                const alreadySelectedIds = new Set(
                    currentState.selectedFiles.map((selectedFile) => selectedFile.id)
                );
                return {
                    ...currentState,
                    selectedFiles: [
                        ...currentState.selectedFiles,
                        ...rangeFiles.filter((rangeFile) => !alreadySelectedIds.has(rangeFile.id)),
                    ],
                };
            });
        },
        [updateState]
    );

    const selectAll = useCallback(
        (files: FileData[]) => {
            updateState((currentState) => ({ ...currentState, selectedFiles: [...files] }));
        },
        [updateState]
    );

    const deselect = useCallback(
        (files: FileData[]) => {
            const deselectedIds = new Set(files.map((file) => file.id));
            updateState((currentState) => ({
                ...currentState,
                selectedFiles: currentState.selectedFiles.filter(
                    (selectedFile) => !deselectedIds.has(selectedFile.id)
                ),
            }));
        },
        [updateState]
    );

    const clear = useCallback(() => {
        updateState((currentState) => createEmptyState(currentState.scopeKey));
    }, [updateState]);

    const selectedIds = useMemo(
        () => new Set(state.selectedFiles.map((selectedFile) => selectedFile.id)),
        [state.selectedFiles]
    );

    const isSelected = useCallback((fileId: number) => selectedIds.has(fileId), [selectedIds]);

    return {
        selectedFiles: state.selectedFiles,
        selectedCount: state.selectedFiles.length,
        hasSelection: state.selectedFiles.length > 0,
        isSelected,
        toggle,
        selectRange,
        selectAll,
        deselect,
        clear,
    };
};

export default useFileSelection;

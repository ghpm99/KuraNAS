import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { useItemSelection } from '@/shared/selection/useItemSelection';

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

const readFileId = (file: FileData) => file.id;

export const useFileSelection = (scopeKey: string): FileSelection => {
    const { selectedItems, selectedCount, hasSelection, isSelected, toggle, selectRange, selectAll, deselect, clear } =
        useItemSelection<FileData>(scopeKey, readFileId);

    return {
        selectedFiles: selectedItems,
        selectedCount,
        hasSelection,
        isSelected,
        toggle,
        selectRange,
        selectAll,
        deselect,
        clear,
    };
};

export default useFileSelection;

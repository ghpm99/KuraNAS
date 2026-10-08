import { createContext, useContext } from 'react';
import type { FileSelection } from './useFileSelection';

const noop = () => undefined;

const inertFileSelection: FileSelection = {
    selectedFiles: [],
    selectedCount: 0,
    hasSelection: false,
    isSelected: () => false,
    toggle: noop,
    selectRange: noop,
    selectAll: noop,
    deselect: noop,
    clear: noop,
};

export const FileSelectionContext = createContext<FileSelection>(inertFileSelection);

export const useFileSelectionContext = (): FileSelection => useContext(FileSelectionContext);

export default useFileSelectionContext;

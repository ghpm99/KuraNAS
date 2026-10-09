import type { ReactNode } from 'react';
import {
    FileContextProvider,
    type FileContextType,
    type FileData,
} from '@/features/files/providers/fileProvider/fileContext';
import { FileSelectionContext } from './fileSelectionContext';
import { useFileSelection } from './useFileSelection';

export const SelectionTestHarness = ({
    fileContext,
    seedFiles,
    children,
}: {
    fileContext: FileContextType;
    seedFiles: FileData[];
    children: ReactNode;
}) => {
    const fileSelection = useFileSelection('test-scope');
    return (
        <FileContextProvider value={fileContext}>
            <FileSelectionContext.Provider value={fileSelection}>
                <button onClick={() => fileSelection.selectAll(seedFiles)}>seed-selection</button>
                <span data-testid="selected-count">{fileSelection.selectedCount}</span>
                {children}
            </FileSelectionContext.Provider>
        </FileContextProvider>
    );
};

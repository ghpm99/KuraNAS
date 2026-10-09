import { useEffect, type ReactNode } from 'react';
import { FileSelectionContext } from './fileSelectionContext';
import { useFileSelection } from './useFileSelection';

const FileSelectionProvider = ({
    scopeKey,
    children,
}: {
    scopeKey: string;
    children: ReactNode;
}) => {
    const fileSelection = useFileSelection(scopeKey);
    const { hasSelection, clear } = fileSelection;

    useEffect(() => {
        if (!hasSelection) return undefined;
        const clearOnEscape = (event: KeyboardEvent) => {
            if (event.key === 'Escape') clear();
        };
        document.addEventListener('keydown', clearOnEscape);
        return () => document.removeEventListener('keydown', clearOnEscape);
    }, [hasSelection, clear]);

    return (
        <FileSelectionContext.Provider value={fileSelection}>
            {children}
        </FileSelectionContext.Provider>
    );
};

export default FileSelectionProvider;

import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import { findItemInTree } from '@/features/files/providers/fileProvider/fileProviderUtils';
import { useCallback, useMemo, useState, type ReactNode } from 'react';
import { FileDetailsContext } from './fileDetailsContext';

type OpenedDetails = {
    file: FileData;
    openedWhileViewingItemId: number | null;
};

export const FileDetailsProvider = ({ children }: { children: ReactNode }) => {
    const { files, selectedItem } = useFile();
    const viewedItemId = selectedItem?.id ?? null;
    const [openedDetails, setOpenedDetails] = useState<OpenedDetails | null>(null);

    const openDetails = useCallback(
        (file: FileData) => setOpenedDetails({ file, openedWhileViewingItemId: viewedItemId }),
        [viewedItemId]
    );
    const closeDetails = useCallback(() => setOpenedDetails(null), []);

    const snapshotTarget =
        openedDetails?.openedWhileViewingItemId === viewedItemId ? openedDetails.file : null;
    const explicitTarget = snapshotTarget
        ? (findItemInTree(
              [...(files ?? []), ...(selectedItem ? [selectedItem] : [])],
              snapshotTarget.id
          ) ?? snapshotTarget)
        : null;

    const contextValue = useMemo(
        () => ({ isAvailable: true, explicitTarget, openDetails, closeDetails }),
        [explicitTarget, openDetails, closeDetails]
    );

    return (
        <FileDetailsContext.Provider value={contextValue}>{children}</FileDetailsContext.Provider>
    );
};

export default FileDetailsProvider;

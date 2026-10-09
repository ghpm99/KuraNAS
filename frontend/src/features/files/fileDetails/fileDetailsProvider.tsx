import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import { useCallback, useMemo, useState, type ReactNode } from 'react';
import { FileDetailsContext } from './fileDetailsContext';

type OpenedDetails = {
    file: FileData;
    openedWhileViewingItemId: number | null;
};

export const FileDetailsProvider = ({ children }: { children: ReactNode }) => {
    const { selectedItem } = useFile();
    const viewedItemId = selectedItem?.id ?? null;
    const [openedDetails, setOpenedDetails] = useState<OpenedDetails | null>(null);

    const openDetails = useCallback(
        (file: FileData) => setOpenedDetails({ file, openedWhileViewingItemId: viewedItemId }),
        [viewedItemId]
    );
    const closeDetails = useCallback(() => setOpenedDetails(null), []);

    const explicitTarget =
        openedDetails?.openedWhileViewingItemId === viewedItemId ? openedDetails.file : null;

    const contextValue = useMemo(
        () => ({ isAvailable: true, explicitTarget, openDetails, closeDetails }),
        [explicitTarget, openDetails, closeDetails]
    );

    return (
        <FileDetailsContext.Provider value={contextValue}>{children}</FileDetailsContext.Provider>
    );
};

export default FileDetailsProvider;

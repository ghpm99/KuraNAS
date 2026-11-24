import useDocumentsLibraryRoot, { type DocumentsLibraryRootStatus } from './useDocumentsLibraryRoot';

const useDocumentsDestinationFolder = (selectedFolderId: number | undefined) => {
    const { status, root } = useDocumentsLibraryRoot();
    const destinationFolderId = selectedFolderId ?? root?.id;
    const isReady = status === 'ready' && destinationFolderId !== undefined;

    return {
        status: status as DocumentsLibraryRootStatus,
        destinationFolderId,
        isReady,
    };
};

export default useDocumentsDestinationFolder;

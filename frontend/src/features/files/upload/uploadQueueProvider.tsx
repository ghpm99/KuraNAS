import { useEffect, useMemo, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { createThrottledListingInvalidation } from './invalidateListingThrottled';
import { isActiveItem } from './uploadQueueTypes';
import { UploadQueueContextProvider } from './uploadQueueContext';
import UploadConflictDialog from './uploadConflictDialog';
import UploadQueuePanel from './uploadQueuePanel';
import useConfirmLeaveWhileUploading from './useConfirmLeaveWhileUploading';
import useUploadQueue from './useUploadQueue';

const UploadQueueProvider = ({ children }: { children: ReactNode }) => {
    const queryClient = useQueryClient();
    const listingInvalidation = useMemo(
        () => createThrottledListingInvalidation(queryClient),
        [queryClient]
    );
    useEffect(() => listingInvalidation.cancel, [listingInvalidation]);

    const uploadQueue = useUploadQueue({ onFileFinished: listingInvalidation.request });
    useConfirmLeaveWhileUploading(uploadQueue.items.some(isActiveItem));

    return (
        <UploadQueueContextProvider value={uploadQueue}>
            {children}
            <UploadQueuePanel />
            <UploadConflictDialog />
        </UploadQueueContextProvider>
    );
};

export default UploadQueueProvider;

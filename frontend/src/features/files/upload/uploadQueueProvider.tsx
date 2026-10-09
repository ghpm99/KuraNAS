import { useEffect, useMemo, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { createThrottledListingInvalidation } from './invalidateListingThrottled';
import { UploadQueueContextProvider } from './uploadQueueContext';
import UploadQueuePanel from './uploadQueuePanel';
import useUploadQueue from './useUploadQueue';

const UploadQueueProvider = ({ children }: { children: ReactNode }) => {
    const queryClient = useQueryClient();
    const listingInvalidation = useMemo(
        () => createThrottledListingInvalidation(queryClient),
        [queryClient]
    );
    useEffect(() => listingInvalidation.cancel, [listingInvalidation]);

    const uploadQueue = useUploadQueue({ onFileFinished: listingInvalidation.request });

    return (
        <UploadQueueContextProvider value={uploadQueue}>
            {children}
            <UploadQueuePanel />
        </UploadQueueContextProvider>
    );
};

export default UploadQueueProvider;

import { useEffect } from 'react';

const confirmLeave = (event: BeforeUnloadEvent): void => {
    event.preventDefault();
    event.returnValue = '';
};

export const useConfirmLeaveWhileUploading = (hasActiveUploads: boolean): void => {
    useEffect(() => {
        if (!hasActiveUploads) return undefined;
        window.addEventListener('beforeunload', confirmLeave);
        return () => window.removeEventListener('beforeunload', confirmLeave);
    }, [hasActiveUploads]);
};

export default useConfirmLeaveWhileUploading;

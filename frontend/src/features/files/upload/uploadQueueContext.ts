import { createContext, useContext } from 'react';
import { defaultConflictPolicy } from './conflictPolicyPreference';
import type { UploadQueue } from './uploadQueueTypes';

const emptyUploadQueue: UploadQueue = {
    items: [],
    conflictPolicy: defaultConflictPolicy,
    setConflictPolicy: () => undefined,
    enqueue: () => undefined,
    cancel: () => undefined,
    cancelAll: () => undefined,
    retry: () => undefined,
    retryFailed: () => undefined,
    clearFinished: () => undefined,
};

const UploadQueueContext = createContext<UploadQueue>(emptyUploadQueue);

export const UploadQueueContextProvider = UploadQueueContext.Provider;

export const useUploadQueueContext = (): UploadQueue => useContext(UploadQueueContext);

export default useUploadQueueContext;

import type { QueryClient } from '@tanstack/react-query';
import { fileQueryKeys } from '../providers/fileProvider/fileProviderUtils';

export const listingInvalidationIntervalMs = 1500;

export type ThrottledInvalidation = {
    request: () => void;
    cancel: () => void;
};

export const createThrottledListingInvalidation = (
    queryClient: QueryClient,
    intervalMs: number = listingInvalidationIntervalMs
): ThrottledInvalidation => {
    let pendingTimer: ReturnType<typeof setTimeout> | null = null;

    const flush = () => {
        pendingTimer = null;
        fileQueryKeys.forEach((queryKey) => {
            void queryClient.invalidateQueries({ queryKey: [queryKey] });
        });
    };

    return {
        request: () => {
            if (pendingTimer !== null) return;
            pendingTimer = setTimeout(flush, intervalMs);
        },
        cancel: () => {
            if (pendingTimer === null) return;
            clearTimeout(pendingTimer);
            pendingTimer = null;
        },
    };
};

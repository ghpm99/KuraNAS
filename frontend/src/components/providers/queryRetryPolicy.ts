import { isClientError } from '@/service/networkError';

const MAX_RETRIES = 1;

export const shouldRetryQuery = (failureCount: number, error: unknown): boolean => {
    if (isClientError(error)) return false;
    return failureCount < MAX_RETRIES;
};

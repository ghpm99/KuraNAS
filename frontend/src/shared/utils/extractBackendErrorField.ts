export const extractBackendErrorField = (error: unknown): string | undefined => {
    const field = (error as { response?: { data?: { field?: unknown } } } | null)?.response?.data
        ?.field;
    return typeof field === 'string' && field !== '' ? field : undefined;
};

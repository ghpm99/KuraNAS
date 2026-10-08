export const extractBackendErrorMessage = (error: unknown): string | undefined => {
    const backendMessage = (error as { response?: { data?: { error?: unknown } } } | null)
        ?.response?.data?.error;
    return typeof backendMessage === 'string' && backendMessage !== '' ? backendMessage : undefined;
};

export const extractBackendMessage = (error: unknown): string | undefined => {
    const backendMessage = (error as { response?: { data?: { error?: unknown } } })?.response?.data
        ?.error;
    return typeof backendMessage === 'string' && backendMessage !== '' ? backendMessage : undefined;
};

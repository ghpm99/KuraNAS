import type { QueryClient } from '@tanstack/react-query';

export const invalidateVideoViews = async (queryClient: QueryClient): Promise<void> => {
    await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['video'] }),
        queryClient.invalidateQueries({ queryKey: ['videos'] }),
    ]);
};

import type { QueryClient } from '@tanstack/react-query';
import { isMusicMetadataDependentQuery } from './isMusicMetadataDependentQuery';

const globalSearchQueryRoots = ['global-search', 'global-search-ai'];

export const invalidateMusicMetadataViews = (queryClient: QueryClient): Promise<unknown> =>
    Promise.all([
        queryClient.invalidateQueries({ predicate: isMusicMetadataDependentQuery }),
        ...globalSearchQueryRoots.map((queryRoot) =>
            queryClient.invalidateQueries({ queryKey: [queryRoot] })
        ),
    ]);

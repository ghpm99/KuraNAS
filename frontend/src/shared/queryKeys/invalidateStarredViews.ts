import type { InvalidateQueryFilters, QueryClient } from '@tanstack/react-query';
import { homeFavoritesQueryKey, searchQueryKey } from './fileQueryKeys';

export const starredViewInvalidations: InvalidateQueryFilters[] = [
    { queryKey: homeFavoritesQueryKey },
    { queryKey: ['files'] },
    { queryKey: [searchQueryKey] },
    { queryKey: ['images', 'library'], refetchType: 'none' },
    { queryKey: ['images', 'count'] },
    { queryKey: ['images', 'timeline'] },
];

export const invalidateStarredViews = (queryClient: QueryClient): Promise<unknown> =>
    Promise.all(starredViewInvalidations.map((filters) => queryClient.invalidateQueries(filters)));

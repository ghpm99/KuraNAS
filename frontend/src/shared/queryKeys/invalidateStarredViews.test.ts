import { QueryClient } from '@tanstack/react-query';
import { homeFavoritesQueryKey, searchQueryKey } from './fileQueryKeys';
import { invalidateStarredViews } from './invalidateStarredViews';

describe('invalidateStarredViews', () => {
    it('invalidates exactly the starred-dependent views', async () => {
        const queryClient = new QueryClient();
        const invalidateQueries = jest.spyOn(queryClient, 'invalidateQueries');

        await invalidateStarredViews(queryClient);

        expect(invalidateQueries.mock.calls.map(([filters]) => filters)).toEqual([
            { queryKey: homeFavoritesQueryKey },
            { queryKey: ['files'] },
            { queryKey: [searchQueryKey] },
            { queryKey: ['images', 'library'], refetchType: 'none' },
            { queryKey: ['images', 'count'] },
            { queryKey: ['images', 'timeline'] },
        ]);
    });

    it('marks the home favorites query as invalidated', async () => {
        const queryClient = new QueryClient();
        queryClient.setQueryData([...homeFavoritesQueryKey], { items: [] });
        queryClient.setQueryData(['unrelated'], 1);

        await invalidateStarredViews(queryClient);

        expect(queryClient.getQueryState([...homeFavoritesQueryKey])?.isInvalidated).toBe(true);
        expect(queryClient.getQueryState(['unrelated'])?.isInvalidated).toBe(false);
    });
});

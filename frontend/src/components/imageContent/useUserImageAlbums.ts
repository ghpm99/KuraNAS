import { useInfiniteQuery } from '@tanstack/react-query';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getImageAlbums } from '@/service/imageAlbum';
import type { ImageAlbum } from '@/types/imageAlbum';

export const userImageAlbumsQueryKey = ['images', 'albums', 'list'] as const;

export const useUserImageAlbums = (isEnabled: boolean) => {
    const albumsQuery = useInfiniteQuery({
        queryKey: userImageAlbumsQueryKey,
        queryFn: ({ pageParam }) => getImageAlbums(pageParam),
        initialPageParam: 1,
        getNextPageParam: (lastPage, loadedPages) =>
            lastPage.pagination?.has_next ? loadedPages.length + 1 : undefined,
        enabled: isEnabled,
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    const albums: ImageAlbum[] = albumsQuery.data?.pages.flatMap((page) => page.items ?? []) ?? [];

    return {
        albums,
        status: albumsQuery.status,
        error: albumsQuery.error,
        hasNextPage: albumsQuery.hasNextPage,
        isFetchingNextPage: albumsQuery.isFetchingNextPage,
        fetchNextPage: albumsQuery.fetchNextPage,
        refetch: albumsQuery.refetch,
    };
};

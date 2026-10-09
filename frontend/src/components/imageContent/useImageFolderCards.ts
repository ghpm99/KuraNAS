import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getImageLibraryFolders } from '@/service/image';
import type { ImageLibraryFolder } from '@/types/imageLibrary';
import type { ImageCollectionCard } from './components/ImageCollectionsPanel';

export const imageFolderPageSize = 48;

const toFolderCard = (folder: ImageLibraryFolder): ImageCollectionCard => ({
    id: folder.path,
    title: folder.name,
    description: folder.path,
    imageCount: folder.image_count,
    coverImageId: folder.cover_file_id,
});

export const useImageFolderCards = (parentPath: string, isEnabled: boolean) => {
    const query = useInfiniteQuery({
        queryKey: ['images', 'folders', parentPath],
        queryFn: ({ pageParam }) =>
            getImageLibraryFolders(parentPath, pageParam, imageFolderPageSize),
        initialPageParam: 1,
        getNextPageParam: (lastPage, loadedPages) =>
            lastPage.pagination?.has_next ? loadedPages.length + 1 : undefined,
        enabled: isEnabled,
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    const cards = useMemo(
        () => query.data?.pages.flatMap((page) => (page.items ?? []).map(toFolderCard)) ?? [],
        [query.data]
    );

    return {
        cards,
        status: query.status,
        error: query.error,
        isFetchNextPageError: query.isFetchNextPageError,
        hasNextPage: query.hasNextPage,
        isFetchingNextPage: query.isFetchingNextPage,
        fetchNextPage: query.fetchNextPage,
        refetch: query.refetch,
    };
};

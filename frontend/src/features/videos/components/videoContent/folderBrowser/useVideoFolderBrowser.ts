import { useInfiniteQuery } from '@tanstack/react-query';
import { useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getVideoLibraryFolders, getVideoLibraryFolderVideos } from '@/service/videoPlayback';

export const videoFolderSearchParam = 'folder';
export const videoFolderPageSize = 48;
export const videoFolderVideosPageSize = 24;

const nextPageNumber = (
    lastPage: { pagination?: { has_next?: boolean } } | undefined,
    loadedPages: unknown[]
) => (lastPage?.pagination?.has_next ? loadedPages.length + 1 : undefined);

export const useVideoFolderBrowser = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedFolder = searchParams.get(videoFolderSearchParam) ?? '';

    const selectFolder = useCallback(
        (folderPath: string | null) =>
            setSearchParams((currentParams) => {
                const nextParams = new URLSearchParams(currentParams);
                if (folderPath) {
                    nextParams.set(videoFolderSearchParam, folderPath);
                } else {
                    nextParams.delete(videoFolderSearchParam);
                }
                return nextParams;
            }),
        [setSearchParams]
    );

    const folderQuery = useInfiniteQuery({
        queryKey: ['videos', 'library-folders', selectedFolder],
        queryFn: ({ pageParam }) =>
            getVideoLibraryFolders(selectedFolder, pageParam, videoFolderPageSize),
        initialPageParam: 1,
        getNextPageParam: nextPageNumber,
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    const videoQuery = useInfiniteQuery({
        queryKey: ['videos', 'library-folder-videos', selectedFolder],
        queryFn: ({ pageParam }) =>
            getVideoLibraryFolderVideos(selectedFolder, pageParam, videoFolderVideosPageSize),
        initialPageParam: 1,
        getNextPageParam: nextPageNumber,
        enabled: selectedFolder !== '',
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    return {
        selectedFolder,
        selectFolder,
        folderQuery,
        videoQuery,
        folders: folderQuery.data?.pages.flatMap((page) => page.items ?? []) ?? [],
        videos: videoQuery.data?.pages.flatMap((page) => page.items ?? []) ?? [],
    };
};

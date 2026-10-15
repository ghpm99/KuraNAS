import {
    getHiddenVideoPlaylists,
    getVideoPlaylistById,
    getVideoPlaylistsByVideo,
    getVideoPlaylistsBySection,
    searchVideoPlaylists,
    type VideoLibraryFilters,
    type VideoLibrarySort,
    type VideoPlaylistSection,
} from '@/service/videoPlayback';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';

export const videoQueryKeys = {
    playlists: ['video', 'playlists'] as const,
    sectionPlaylists: (section: VideoPlaylistSection, pageSize: number) =>
        ['video', 'playlists', 'section', section, pageSize] as const,
    playlistDetail: (playlistId?: number) => ['video', 'playlist-detail', playlistId] as const,
    playlistItems: (playlistId?: number) => ['video', 'playlist-items', playlistId] as const,
    hiddenPlaylists: (pageSize: number) => ['video', 'playlists', 'hidden', pageSize] as const,
    playlistSearch: (query: string) => ['video', 'playlists', 'search', query] as const,
    homeCatalog: ['video', 'home-catalog'] as const,
    libraryFiles: (search: string, sort: VideoLibrarySort, filters: VideoLibraryFilters) =>
        [
            'video',
            'library-files',
            search,
            sort.key,
            sort.order,
            filters.formats.join(','),
            filters.resolution ?? '',
            filters.duration ?? '',
        ] as const,
    playbackState: ['video', 'playback-state'] as const,
    continueWatching: ['video', 'continue-watching'] as const,
    playlistsOfVideo: (videoId: number) => ['video', 'playlists-of-video', videoId] as const,
};

export const useVideoPlaylistSearch = (query: string, pageSize: number, isEnabled: boolean) => {
    return useInfiniteQuery({
        queryKey: videoQueryKeys.playlistSearch(query),
        queryFn: ({ pageParam }) => searchVideoPlaylists(query, pageParam, pageSize),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
        enabled: isEnabled,
    });
};

export const useHiddenVideoPlaylists = (pageSize: number, isEnabled: boolean) => {
    return useInfiniteQuery({
        queryKey: videoQueryKeys.hiddenPlaylists(pageSize),
        enabled: isEnabled,
        queryFn: ({ pageParam }) => getHiddenVideoPlaylists(pageParam, pageSize),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage?.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
    });
};

export const useVideoPlaylistsOfVideo = (videoId: number, isEnabled: boolean) => {
    return useQuery({
        queryKey: videoQueryKeys.playlistsOfVideo(videoId),
        queryFn: () => getVideoPlaylistsByVideo(videoId),
        enabled: isEnabled,
    });
};

export const useVideoPlaylistDetail = (playlistId?: number) => {
    return useQuery({
        queryKey: videoQueryKeys.playlistDetail(playlistId),
        queryFn: () => getVideoPlaylistById(playlistId as number),
        enabled: Boolean(playlistId),
    });
};

export const useVideoSectionPlaylists = (
    section: VideoPlaylistSection,
    pageSize: number,
    isEnabled: boolean
) => {
    return useInfiniteQuery({
        queryKey: videoQueryKeys.sectionPlaylists(section, pageSize),
        enabled: isEnabled,
        queryFn: ({ pageParam }) => getVideoPlaylistsBySection(section, pageParam, pageSize),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage?.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
    });
};

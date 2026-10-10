import {
    getAllVideoFiles,
    getVideoHomeCatalog,
    getVideoPlaylistById,
    getVideoPlaylistsByVideo,
    getVideoPlaylists,
    getVideoPlaylistsBySection,
    getVideosWithoutPlaylist,
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
    unassigned: ['video', 'unassigned'] as const,
    allFiles: ['video', 'all-files'] as const,
    homeCatalog: ['video', 'home-catalog'] as const,
    libraryFiles: (search: string, sort: VideoLibrarySort) =>
        ['video', 'library-files', search, sort.key, sort.order] as const,
    playbackState: ['video', 'playback-state'] as const,
    continueWatching: ['video', 'continue-watching'] as const,
    playlistsOfVideo: (videoId: number) => ['video', 'playlists-of-video', videoId] as const,
};

export const useVideoPlaylists = () => {
    return useQuery({
        queryKey: videoQueryKeys.playlists,
        queryFn: () => getVideoPlaylists(false),
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

export const useVideosWithoutPlaylist = () => {
    return useQuery({
        queryKey: videoQueryKeys.unassigned,
        queryFn: () => getVideosWithoutPlaylist(2000),
    });
};

export const useAllVideoFiles = () => {
    return useQuery({
        queryKey: videoQueryKeys.allFiles,
        queryFn: () => getAllVideoFiles(3000),
    });
};

export const useVideoHomeCatalog = () => {
    return useQuery({
        queryKey: videoQueryKeys.homeCatalog,
        queryFn: () => getVideoHomeCatalog(24),
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

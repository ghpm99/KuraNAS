import {
    getAllVideoFiles,
    getVideoHomeCatalog,
    getVideoPlaylistById,
    getVideoPlaylists,
    getVideosWithoutPlaylist,
} from '@/service/videoPlayback';
import { useQuery } from '@tanstack/react-query';

export const videoQueryKeys = {
    playlists: ['video', 'playlists'] as const,
    playlistDetail: (playlistId?: number) => ['video', 'playlist-detail', playlistId] as const,
    unassigned: ['video', 'unassigned'] as const,
    allFiles: ['video', 'all-files'] as const,
    homeCatalog: ['video', 'home-catalog'] as const,
    libraryFiles: (search: string) => ['video', 'library-files', search] as const,
    playbackState: ['video', 'playback-state'] as const,
    playlistMembership: (key: string) => ['video', 'playlist-membership', key] as const,
};

export const useVideoPlaylists = () => {
    return useQuery({
        queryKey: videoQueryKeys.playlists,
        queryFn: () => getVideoPlaylists(false),
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

import { getAllVideoFiles, getVideoPlaylistById, getVideoPlaylists, getVideosWithoutPlaylist } from '@/service/videoPlayback';
import { useQuery } from '@tanstack/react-query';

export const useVideoPlaylists = () => {
	return useQuery({
		queryKey: ['video-playlists'],
		queryFn: () => getVideoPlaylists(false),
	});
};

export const useVideoPlaylistDetail = (playlistId?: number) => {
	return useQuery({
		queryKey: ['video-playlist', playlistId],
		queryFn: () => getVideoPlaylistById(playlistId as number),
		enabled: Boolean(playlistId),
	});
};

export const useVideosWithoutPlaylist = () => {
	return useQuery({
		queryKey: ['video-unassigned'],
		queryFn: () => getVideosWithoutPlaylist(2000),
	});
};

export const useAllVideoFiles = () => {
	return useQuery({
		queryKey: ['all-video-files'],
		queryFn: () => getAllVideoFiles(3000),
	});
};

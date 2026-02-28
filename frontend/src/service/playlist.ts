import { Pagination } from '@/types/pagination';
import { Playlist, PlaylistTrack, CreatePlaylistRequest, UpdatePlaylistRequest } from '@/types/playlist';
import { apiBase } from '.';

export const getPlaylists = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<Playlist>>('/music/playlists/', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getPlaylistById = async (id: number) => {
	const response = await apiBase.get<Playlist>(`/music/playlists/${id}`);
	return response.data;
};

export const createPlaylist = async (req: CreatePlaylistRequest) => {
	const response = await apiBase.post<Playlist>('/music/playlists/', req);
	return response.data;
};

export const updatePlaylist = async (id: number, req: UpdatePlaylistRequest) => {
	const response = await apiBase.put<Playlist>(`/music/playlists/${id}`, req);
	return response.data;
};

export const deletePlaylist = async (id: number) => {
	await apiBase.delete(`/music/playlists/${id}`);
};

export const getPlaylistTracks = async (id: number, page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<PlaylistTrack>>(`/music/playlists/${id}/tracks`, {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const addTrackToPlaylist = async (playlistId: number, fileId: number) => {
	const response = await apiBase.post<PlaylistTrack>(`/music/playlists/${playlistId}/tracks`, {
		file_id: fileId,
	});
	return response.data;
};

export const removeTrackFromPlaylist = async (playlistId: number, fileId: number) => {
	await apiBase.delete(`/music/playlists/${playlistId}/tracks/${fileId}`);
};

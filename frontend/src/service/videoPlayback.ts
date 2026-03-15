import { apiBase } from '.';
import type { Pagination } from '@/types/pagination';

export interface VideoFileDto {
	id: number;
	name: string;
	path: string;
	parent_path: string;
	format: string;
	size: number;
	created_at?: string;
	updated_at?: string;
}

export interface VideoPlaylistItemDto {
	id: number;
	order_index: number;
	source_kind: 'auto' | 'manual';
	video: VideoFileDto;
	status: 'not_started' | 'in_progress' | 'completed';
	progress_pct: number;
}

export interface VideoPlaylistDto {
	id: number;
	type: 'folder' | 'series' | 'movie' | 'custom';
	source_path: string;
	name: string;
	is_hidden: boolean;
	is_auto: boolean;
	group_mode: 'folder' | 'prefix' | 'single';
	classification: 'anime' | 'series' | 'movie' | 'personal' | 'clip' | 'program';
	item_count: number;
	cover_video_id: number | null;
	created_at: string;
	updated_at: string;
	last_played_at: string | null;
	items: VideoPlaylistItemDto[];
}

export interface VideoPlaybackStateDto {
	id: number;
	client_id: string;
	playlist_id: number | null;
	video_id: number | null;
	current_time: number;
	duration: number;
	is_paused: boolean;
	completed: boolean;
	last_update: string;
}

export interface VideoPlaybackSessionDto {
	playlist: VideoPlaylistDto;
	playback_state: VideoPlaybackStateDto;
}

export interface VideoCatalogItemDto {
	video: VideoFileDto;
	status: 'not_started' | 'in_progress' | 'completed';
	progress_pct: number;
}

export interface VideoCatalogSectionDto {
	key: 'continue' | 'series' | 'movies' | 'personal' | 'recent';
	title: string;
	items: VideoCatalogItemDto[];
}

export interface VideoHomeCatalogDto {
	sections: VideoCatalogSectionDto[];
}

interface PaginationResponse<T> {
	items: T[];
}

export interface VideoPlaylistMembershipDto {
	playlist_id: number;
	video_id: number;
}

export interface UpdateVideoPlaybackStateRequest {
	playlist_id?: number | null;
	video_id?: number | null;
	current_time?: number;
	duration?: number;
	is_paused?: boolean;
	completed?: boolean;
}

export interface ReorderVideoPlaylistItemRequest {
	video_id: number;
	order_index: number;
}

export const startVideoPlayback = async (
	videoId: number,
	playlistId?: number | null,
): Promise<VideoPlaybackSessionDto> => {
	const response = await apiBase.post<VideoPlaybackSessionDto>('/video/playback/start', {
		video_id: videoId,
		playlist_id: playlistId ?? null,
	});
	return response.data;
};

export const getVideoPlaybackState = async (): Promise<VideoPlaybackSessionDto> => {
	const response = await apiBase.get<VideoPlaybackSessionDto>('/video/playback/state');
	return response.data;
};

export const updateVideoPlaybackState = async (
	state: UpdateVideoPlaybackStateRequest,
): Promise<VideoPlaybackStateDto> => {
	const response = await apiBase.put<VideoPlaybackStateDto>('/video/playback/state', state);
	return response.data;
};

export const nextVideoPlayback = async (): Promise<VideoPlaybackSessionDto> => {
	const response = await apiBase.post<VideoPlaybackSessionDto>('/video/playback/next');
	return response.data;
};

export const previousVideoPlayback = async (): Promise<VideoPlaybackSessionDto> => {
	const response = await apiBase.post<VideoPlaybackSessionDto>('/video/playback/previous');
	return response.data;
};

export const getVideoHomeCatalog = async (limit = 24): Promise<VideoHomeCatalogDto> => {
	const response = await apiBase.get<VideoHomeCatalogDto>('/video/catalog/home', { params: { limit } });
	return response.data;
};

export const getVideoPlaylists = async (includeHidden = false): Promise<VideoPlaylistDto[]> => {
	const response = await apiBase.get<VideoPlaylistDto[]>('/video/playlists/', {
		params: { include_hidden: includeHidden },
	});
	return response.data;
};

export const getVideoPlaylistMemberships = async (includeHidden = false): Promise<VideoPlaylistMembershipDto[]> => {
	const response = await apiBase.get<VideoPlaylistMembershipDto[]>('/video/playlists/memberships', {
		params: { include_hidden: includeHidden },
	});
	return response.data;
};

export const getVideoPlaylistById = async (playlistId: number): Promise<VideoPlaylistDto> => {
	const response = await apiBase.get<VideoPlaylistDto>(`/video/playlists/${playlistId}`);
	return response.data;
};

export const setVideoPlaylistHidden = async (playlistId: number, hidden: boolean): Promise<void> => {
	await apiBase.put(`/video/playlists/${playlistId}/hidden`, { hidden });
};

export const addVideoToPlaylist = async (playlistId: number, videoId: number): Promise<void> => {
	await apiBase.post(`/video/playlists/${playlistId}/videos`, { video_id: videoId });
};

export const removeVideoFromPlaylist = async (playlistId: number, videoId: number): Promise<void> => {
	await apiBase.delete(`/video/playlists/${playlistId}/videos/${videoId}`);
};

export const reorderVideoPlaylist = async (
	playlistId: number,
	items: ReorderVideoPlaylistItemRequest[],
): Promise<void> => {
	await apiBase.put(`/video/playlists/${playlistId}/reorder`, { items });
};

export const updateVideoPlaylistName = async (playlistId: number, name: string): Promise<void> => {
	await apiBase.put(`/video/playlists/${playlistId}`, { name });
};

export const getVideosWithoutPlaylist = async (limit = 2000): Promise<VideoFileDto[]> => {
	const response = await apiBase.get<VideoFileDto[]>('/video/playlists/unassigned', { params: { limit } });
	return response.data;
};

export const getAllVideoFiles = async (limit = 2000): Promise<VideoFileDto[]> => {
	const response = await apiBase.get<PaginationResponse<VideoFileDto>>('/files/videos', {
		params: { page: 1, page_size: limit },
	});
	return response.data.items ?? [];
};

export const getVideoLibraryFiles = async (
	page: number,
	pageSize: number,
	searchQuery = '',
): Promise<Pagination<VideoFileDto>> => {
	const response = await apiBase.get<Pagination<VideoFileDto>>('/video/library/files', {
		params: {
			page,
			page_size: pageSize,
			query: searchQuery,
		},
	});
	return response.data;
};

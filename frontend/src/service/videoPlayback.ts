import { apiBase } from '.';

export interface VideoFileDto {
	id: number;
	name: string;
	path: string;
	parent_path: string;
	format: string;
	size: number;
}

export interface VideoPlaylistItemDto {
	id: number;
	order_index: number;
	video: VideoFileDto;
	status: 'not_started' | 'in_progress' | 'completed';
}

export interface VideoPlaylistDto {
	id: number;
	type: 'folder' | 'series' | 'movie' | 'custom';
	source_path: string;
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

export interface UpdateVideoPlaybackStateRequest {
	playlist_id?: number | null;
	video_id?: number | null;
	current_time?: number;
	duration?: number;
	is_paused?: boolean;
	completed?: boolean;
}

export const startVideoPlayback = async (videoId: number): Promise<VideoPlaybackSessionDto> => {
	const response = await apiBase.post<VideoPlaybackSessionDto>('/video/playback/start', { video_id: videoId });
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

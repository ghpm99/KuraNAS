import { apiBase } from '.';

export interface GlobalSearchFileResult {
	id: number;
	name: string;
	path: string;
	parent_path: string;
	format: string;
	starred: boolean;
}

export interface GlobalSearchFolderResult {
	id: number;
	name: string;
	path: string;
	parent_path: string;
	starred: boolean;
}

export interface GlobalSearchArtistResult {
	key: string;
	artist: string;
	track_count: number;
	album_count: number;
}

export interface GlobalSearchAlbumResult {
	key: string;
	artist: string;
	album: string;
	year: string;
	track_count: number;
}

export interface GlobalSearchPlaylistResult {
	scope: 'music' | 'video';
	id: number;
	name: string;
	description: string;
	count: number;
	classification: string;
	source_path: string;
	is_auto: boolean;
}

export interface GlobalSearchVideoResult {
	id: number;
	name: string;
	path: string;
	parent_path: string;
	format: string;
}

export interface GlobalSearchImageResult {
	id: number;
	name: string;
	path: string;
	parent_path: string;
	format: string;
	category: string;
	context: string;
}

export interface GlobalSearchResponse {
	query: string;
	files: GlobalSearchFileResult[];
	folders: GlobalSearchFolderResult[];
	artists: GlobalSearchArtistResult[];
	albums: GlobalSearchAlbumResult[];
	playlists: GlobalSearchPlaylistResult[];
	videos: GlobalSearchVideoResult[];
	images: GlobalSearchImageResult[];
}

export const searchGlobal = async (query: string, limit = 6): Promise<GlobalSearchResponse> => {
	const response = await apiBase.get<GlobalSearchResponse>('/search/global', {
		params: {
			q: query,
			limit,
		},
	});
	return response.data;
};

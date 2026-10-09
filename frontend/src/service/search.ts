import { apiBase } from '.';

export interface GlobalSearchFileResult {
    id: number;
    name: string;
    path: string;
    parent_path: string;
    format: string;
    starred: boolean;
    size: number;
    updated_at: string;
    tier: 'hot' | 'cold';
}

export interface GlobalSearchFolderResult {
    id: number;
    name: string;
    path: string;
    parent_path: string;
    starred: boolean;
    size: number;
    updated_at: string;
    tier: 'hot' | 'cold';
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
    updated_at: string;
}

export interface GlobalSearchImageResult {
    id: number;
    name: string;
    path: string;
    parent_path: string;
    format: string;
    updated_at: string;
    category: string;
    context: string;
}

export interface GlobalSearchTrackResult {
    file_id: number;
    title: string;
    artist: string;
    album: string;
    album_key: string;
    duration: number;
    path: string;
}

export interface GlobalSearchResponse {
    query: string;
    suggestion?: string;
    files: GlobalSearchFileResult[];
    folders: GlobalSearchFolderResult[];
    artists: GlobalSearchArtistResult[];
    albums: GlobalSearchAlbumResult[];
    playlists: GlobalSearchPlaylistResult[];
    videos: GlobalSearchVideoResult[];
    images: GlobalSearchImageResult[];
    tracks?: GlobalSearchTrackResult[];
}

export const searchGlobal = async (
    query: string,
    limit = 6,
    signal?: AbortSignal
): Promise<GlobalSearchResponse> => {
    const response = await apiBase.get<GlobalSearchResponse>('/search/global', {
        params: {
            q: query,
            limit,
        },
        signal,
    });
    return response.data;
};

export const searchGlobalWithAI = async (
    query: string,
    limit = 6,
    signal?: AbortSignal
): Promise<GlobalSearchResponse> => {
    const response = await apiBase.get<GlobalSearchResponse>('/search/global/ai', {
        params: {
            q: query,
            limit,
        },
        signal,
    });
    return response.data;
};

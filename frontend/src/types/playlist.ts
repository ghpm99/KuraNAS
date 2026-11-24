import { IMusicData } from '@/types/music';

export type PlaylistRuleMatch = 'any' | 'all';

export interface PlaylistRules {
    match: PlaylistRuleMatch;
    genres: string[];
    artists: string[];
    albums: string[];
    folders: string[];
    year_from: number | null;
    year_to: number | null;
    favorites_only: boolean;
}

export interface PlaylistRulePreview {
    track_count: number;
}

export interface Playlist {
    id: number;
    name: string;
    description: string;
    is_system: boolean;
    is_auto: boolean;
    is_ai_generated?: boolean;
    kind: string;
    source_key: string;
    created_at: string;
    updated_at: string;
    track_count: number;
    total_length_seconds?: number;
    rules?: PlaylistRules | null;
}

export interface PlaylistTrack {
    id: number;
    position: number;
    added_at: string;
    file: IMusicData;
}

export interface CreatePlaylistRequest {
    name: string;
    description?: string;
    rules?: PlaylistRules;
}

export interface UpdatePlaylistRequest {
    name: string;
    description?: string;
    rules?: PlaylistRules;
}

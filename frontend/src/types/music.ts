export interface IMusicData {
    id: number;
    name: string;
    path: string;
    type: number;
    format: string;
    size: number;
    updated_at: string;
    created_at: string;
    deleted_at: string;
    last_interaction: string;
    last_backup: string;
    check_sum: string;
    directory_content_count: number;
    starred: boolean;
    metadata?: IMusicMetadata;
}

export interface IMusicMetadata {
    mime?: string;
    length?: number;
    bitrate?: number;
    sample_rate?: number;
    channels?: number;
    bitrate_mode?: number;
    encoder_info?: string;
    bit_depth?: number;
    title?: string;
    artist?: string;
    album?: string;
    album_artist?: string;
    track_number?: string;
    genre?: string;
    composer?: string;
    year?: string;
    recording_date?: string;
    encoder?: string;
    publisher?: string;
    original_release_date?: string;
    original_artist?: string;
    lyricist?: string;
    lyrics?: string;
    disc_number?: string;
}

export interface MusicArtist {
    key: string;
    artist: string;
    track_count: number;
    album_count: number;
}

export interface MusicAlbum {
    key: string;
    album: string;
    artist: string;
    year: string;
    track_count: number;
}

export interface MusicGenre {
    key: string;
    genre: string;
    track_count: number;
}

export interface MusicFolder {
    folder: string;
    track_count: number;
}

export interface MusicAlbumSummary {
    key: string;
    name: string;
    artist: string;
    year: string;
    track_count: number;
    total_length_seconds: number;
    disc_count: number;
}

export interface MusicArtistSummary {
    key: string;
    name: string;
    track_count: number;
    album_count: number;
    total_length_seconds: number;
}

export interface MusicGroupSummary {
    key: string;
    name: string;
    track_count: number;
    total_length_seconds: number;
}

export interface MusicPlayedTrack {
    track: IMusicData;
    play_count: number;
    last_played_at: string;
}

export type MusicMostPlayedPeriod = 'all' | '30d';

export interface MusicHomeCatalog {
    summary: {
        total_tracks: number;
        total_artists: number;
        total_albums: number;
        total_genres: number;
        total_folders: number;
    };
    playlists: Array<{
        id: number;
        name: string;
        description: string;
        is_system: boolean;
        is_auto: boolean;
        kind: string;
        source_key: string;
        created_at: string;
        updated_at: string;
        track_count: number;
    }>;
    artists: MusicArtist[];
    albums: MusicAlbum[];
}

export type MusicListSortField = 'tracks' | 'name' | 'recent' | 'year';

export type MusicListSortOrder = 'asc' | 'desc';

export interface MusicListSort {
    sort: MusicListSortField;
    order: MusicListSortOrder;
}

export interface MusicQueueEntry {
    file_id: number;
    name: string;
    path: string;
    format: string;
    title: string;
    artist: string;
    album: string;
    length: number;
}

export interface MusicQueue {
    items: MusicQueueEntry[];
    truncated: boolean;
}

export interface PlayerQueue {
    items: MusicQueueEntry[];
    current_index: number;
}

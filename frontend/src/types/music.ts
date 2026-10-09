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

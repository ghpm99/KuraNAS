export type MusicView = 'all' | 'artists' | 'albums' | 'genres' | 'folders';

export interface MusicArtist {
	artist: string;
	track_count: number;
	album_count: number;
}

export interface MusicAlbum {
	album: string;
	artist: string;
	year: string;
	track_count: number;
}

export interface MusicGenre {
	genre: string;
	track_count: number;
}

export interface MusicFolder {
	folder: string;
	track_count: number;
}

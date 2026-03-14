import { IMusicData } from '@/components/providers/musicProvider/musicProvider';

export interface Playlist {
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
}

export interface UpdatePlaylistRequest {
	name: string;
	description?: string;
}

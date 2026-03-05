import { apiBase } from '.';

export interface PlayerStateDto {
	id: number;
	client_id: string;
	playlist_id: number | null;
	current_file_id: number | null;
	current_position: number;
	volume: number;
	shuffle: boolean;
	repeat_mode: string;
	updated_at: string;
}

export interface UpdatePlayerStateRequest {
	playlist_id?: number | null;
	current_file_id?: number | null;
	current_position?: number;
	volume?: number;
	shuffle?: boolean;
	repeat_mode?: string;
}

export const getPlayerState = async (): Promise<PlayerStateDto> => {
	const response = await apiBase.get<PlayerStateDto>('/music/player-state/');
	return response.data;
};

export const updatePlayerState = async (state: UpdatePlayerStateRequest): Promise<PlayerStateDto> => {
	const response = await apiBase.put<PlayerStateDto>('/music/player-state/', state);
	return response.data;
};

import { apiBase } from '.';
import type { PlayerQueue } from '@/types/music';

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

export const updatePlayerState = async (
    state: UpdatePlayerStateRequest
): Promise<PlayerStateDto> => {
    const response = await apiBase.put<PlayerStateDto>('/music/player-state/', state);
    return response.data;
};

export interface ReplacePlayerQueueRequest {
    file_ids: number[];
    current_index: number;
}

export const getPlayerQueue = async (): Promise<PlayerQueue> => {
    const response = await apiBase.get<PlayerQueue>('/music/player-state/queue');
    return response.data;
};

export const replacePlayerQueue = async (request: ReplacePlayerQueueRequest): Promise<void> => {
    await apiBase.put('/music/player-state/queue', request);
};

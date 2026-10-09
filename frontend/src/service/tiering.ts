import { apiBase } from '@/service';
import type {
	TieringFileLocation,
	TieringSettings,
	TieringStatus,
	TieringUsage,
} from '@/types/tiering';

export const getTieringSettings = async (): Promise<TieringSettings> => {
	const response = await apiBase.get<TieringSettings>('/tiering/settings');
	return response.data;
};

export const updateTieringSettings = async (settings: TieringSettings): Promise<TieringSettings> => {
	const response = await apiBase.put<TieringSettings>('/tiering/settings', settings);
	return response.data;
};

export const getTieringStatus = async (): Promise<TieringStatus> => {
	const response = await apiBase.get<TieringStatus>('/tiering/status');
	return response.data;
};

export const getTieringUsage = async (): Promise<TieringUsage> => {
	const response = await apiBase.get<TieringUsage>('/tiering/usage');
	return response.data;
};

export const promoteFileToHot = async (fileId: number): Promise<TieringFileLocation> => {
	const response = await apiBase.post<TieringFileLocation>(`/tiering/promote/${fileId}`);
	return response.data;
};

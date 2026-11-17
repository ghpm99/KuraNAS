import { apiBase } from '@/service';
import type {
	AutoShutdownSettings,
	AutoShutdownStatus,
	SuggestedShutdownTime,
} from '@/types/autoShutdown';

export const getAutoShutdownSettings = async (): Promise<AutoShutdownSettings> => {
	const response = await apiBase.get<AutoShutdownSettings>('/auto-shutdown/settings');
	return response.data;
};

export const updateAutoShutdownSettings = async (
	settings: AutoShutdownSettings
): Promise<AutoShutdownSettings> => {
	const response = await apiBase.put<AutoShutdownSettings>('/auto-shutdown/settings', settings);
	return response.data;
};

export const getSuggestedShutdownTime = async (): Promise<SuggestedShutdownTime> => {
	const response = await apiBase.get<SuggestedShutdownTime>('/auto-shutdown/suggested-time');
	return response.data;
};

export const getAutoShutdownStatus = async (): Promise<AutoShutdownStatus> => {
	const response = await apiBase.get<AutoShutdownStatus>('/auto-shutdown/status');
	return response.data;
};

export const cancelAutoShutdownToday = async (): Promise<AutoShutdownStatus> => {
	const response = await apiBase.post<AutoShutdownStatus>('/auto-shutdown/cancel');
	return response.data;
};

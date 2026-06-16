import { apiBase } from '@/service';
import type { AutoShutdownSettings, SuggestedShutdownTime } from '@/types/autoShutdown';

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

import { AboutContextType } from '@/components/providers/aboutProvider/AboutContext';
import { apiBase } from '.';

export const getAboutConfiguration = async (): Promise<AboutContextType> => {
	const response = await apiBase.get<AboutContextType>('/configuration/about');
	return response.data;
};

export const getTranslations = async (): Promise<Record<string, string>> => {
	const response = await apiBase.get<Record<string, string>>('/configuration/translation');
	return response.data;
};

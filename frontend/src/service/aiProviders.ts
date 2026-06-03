import { apiBase } from '@/service';
import type { AIProviderDto, AIProviderName, UpdateAIProviderRequest } from '@/types/aiProviders';

export const getAIProviders = async (): Promise<AIProviderDto[]> => {
	const response = await apiBase.get<AIProviderDto[]>('/ai/providers');
	return response.data;
};

export const updateAIProvider = async (
	name: AIProviderName,
	request: UpdateAIProviderRequest
): Promise<AIProviderDto> => {
	const response = await apiBase.put<AIProviderDto>(
		`/ai/providers/${encodeURIComponent(name)}`,
		request
	);
	return response.data;
};

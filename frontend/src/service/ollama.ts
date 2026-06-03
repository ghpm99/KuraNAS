import { apiBase } from '@/service';
import type { OllamaModel, OllamaStatus, PullModelResponse } from '@/types/ollama';

export const getOllamaStatus = async (): Promise<OllamaStatus> => {
	const response = await apiBase.get<OllamaStatus>('/ai/ollama/status');
	return response.data;
};

export const listOllamaModels = async (): Promise<OllamaModel[]> => {
	const response = await apiBase.get<OllamaModel[]>('/ai/ollama/models');
	return response.data;
};

export const pullOllamaModel = async (model: string): Promise<PullModelResponse> => {
	const response = await apiBase.post<PullModelResponse>('/ai/ollama/models/pull', { model });
	return response.data;
};

export const deleteOllamaModel = async (name: string): Promise<void> => {
	await apiBase.delete(`/ai/ollama/models/${encodeURIComponent(name)}`);
};

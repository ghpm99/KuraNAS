import { apiBase } from '@/service';
import type { Job } from '@/types/jobs';

export const getJob = async (id: number): Promise<Job> => {
	const response = await apiBase.get<Job>(`/jobs/${id}`);
	return response.data;
};

export const cancelJob = async (id: number): Promise<void> => {
	await apiBase.post(`/jobs/${id}/cancel`);
};

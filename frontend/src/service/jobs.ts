import { apiBase } from '@/service';
import type { Job } from '@/types/jobs';

export const getJob = async (id: number): Promise<Job> => {
	const response = await apiBase.get<Job>(`/jobs/${id}`);
	return response.data;
};

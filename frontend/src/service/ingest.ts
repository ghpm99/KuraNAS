import { apiBase } from '@/service';
import type { YtDlpStatus } from '@/types/ingest';

export const getYtDlpStatus = async (): Promise<YtDlpStatus> => {
	const response = await apiBase.get<YtDlpStatus>('/ingest/ytdlp/status');
	return response.data;
};

export const updateYtDlp = async (): Promise<void> => {
	await apiBase.post('/ingest/ytdlp/update');
};

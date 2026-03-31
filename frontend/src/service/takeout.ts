import { apiBase } from '@/service';
import type { CompleteTakeoutUploadResponse, InitTakeoutUploadResponse } from '@/types/takeout';

export const initTakeoutUpload = async (
	fileName: string,
	size: number
): Promise<InitTakeoutUploadResponse> => {
	const response = await apiBase.post<InitTakeoutUploadResponse>('/takeout/upload/init', {
		file_name: fileName,
		size,
	});
	return response.data;
};

export const uploadTakeoutChunk = async (
	uploadId: string,
	chunk: Blob,
	offset: number
): Promise<{ received: boolean }> => {
	const formData = new FormData();
	formData.append('chunk', chunk);
	formData.append('upload_id', uploadId);
	formData.append('offset', String(offset));

	const response = await apiBase.post<{ received: boolean }>('/takeout/upload/chunk', formData, {
		headers: {
			'Content-Type': 'multipart/form-data',
		},
	});
	return response.data;
};

export const completeTakeoutUpload = async (
	uploadId: string
): Promise<CompleteTakeoutUploadResponse> => {
	const response = await apiBase.post<CompleteTakeoutUploadResponse>('/takeout/upload/complete', {
		upload_id: uploadId,
	});
	return response.data;
};

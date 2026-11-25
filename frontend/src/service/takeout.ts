import { apiBase } from '@/service';
import type { Pagination } from '@/types/pagination';
import type {
	CompleteTakeoutUploadResponse,
	InitTakeoutUploadResponse,
	ListTakeoutImportsParams,
	TakeoutImport,
	TakeoutUploadStatus,
} from '@/types/takeout';

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

export const getTakeoutUploadStatus = async (uploadId: string): Promise<TakeoutUploadStatus> => {
	const response = await apiBase.get<TakeoutUploadStatus>(`/takeout/upload/${uploadId}`);
	return response.data;
};

export const cancelTakeoutUpload = async (uploadId: string): Promise<void> => {
	await apiBase.delete(`/takeout/upload/${uploadId}`);
};

export const listTakeoutImports = async ({
	page = 1,
	pageSize = 20,
}: ListTakeoutImportsParams = {}): Promise<Pagination<TakeoutImport>> => {
	const response = await apiBase.get<Pagination<TakeoutImport>>('/takeout/imports', {
		params: { page, page_size: pageSize },
	});
	if (!Array.isArray(response.data?.items)) {
		throw new Error('invalid takeout imports response');
	}
	return response.data;
};

export const getTakeoutImport = async (importId: number): Promise<TakeoutImport> => {
	const response = await apiBase.get<TakeoutImport>(`/takeout/imports/${importId}`);
	return response.data;
};

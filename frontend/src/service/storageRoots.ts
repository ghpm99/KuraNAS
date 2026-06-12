import { apiBase } from '@/service';
import type {
	CreateStorageRootRequest,
	StorageRootDto,
	UpdateStorageRootRequest,
} from '@/types/storageRoots';

export const getStorageRoots = async (): Promise<StorageRootDto[]> => {
	const response = await apiBase.get<StorageRootDto[]>('/storage-roots');
	return response.data;
};

export const createStorageRoot = async (
	request: CreateStorageRootRequest
): Promise<StorageRootDto> => {
	const response = await apiBase.post<StorageRootDto>('/storage-roots', request);
	return response.data;
};

export const updateStorageRoot = async (
	id: number,
	request: UpdateStorageRootRequest
): Promise<StorageRootDto> => {
	const response = await apiBase.put<StorageRootDto>(`/storage-roots/${id}`, request);
	return response.data;
};

export const deleteStorageRoot = async (id: number): Promise<void> => {
	await apiBase.delete(`/storage-roots/${id}`);
};

import { apiBase } from '@/service';
import type {
	AllowedIPDto,
	ClientIPDto,
	CreateAllowedIPRequest,
	UpdateAllowedIPRequest,
} from '@/types/accessControl';

export const getAllowedIPs = async (): Promise<AllowedIPDto[]> => {
	const response = await apiBase.get<AllowedIPDto[]>('/access-control/ips');
	return response.data;
};

export const createAllowedIP = async (request: CreateAllowedIPRequest): Promise<AllowedIPDto> => {
	const response = await apiBase.post<AllowedIPDto>('/access-control/ips', request);
	return response.data;
};

export const updateAllowedIP = async (
	id: number,
	request: UpdateAllowedIPRequest
): Promise<AllowedIPDto> => {
	const response = await apiBase.put<AllowedIPDto>(`/access-control/ips/${id}`, request);
	return response.data;
};

export const deleteAllowedIP = async (id: number): Promise<void> => {
	await apiBase.delete(`/access-control/ips/${id}`);
};

export const getClientIP = async (): Promise<ClientIPDto> => {
	const response = await apiBase.get<ClientIPDto>('/access-control/client-ip');
	return response.data;
};

import { UpdateStatus } from '@/types/update';
import { apiBase } from '.';

export const getUpdateStatus = async () => {
    const response = await apiBase.get<UpdateStatus>('/update/status');
    return response.data;
};

export const applyUpdate = async () => {
    const response = await apiBase.post('/update/apply');
    return response.data;
};

import { apiBase } from '@/service';
import type { TrashPage, TrashRetention } from '@/types/trash';

export const getTrashItems = async (page: number, pageSize: number): Promise<TrashPage> => {
    const response = await apiBase.get<TrashPage>('/trash', {
        params: { page, page_size: pageSize },
    });
    return response.data;
};

export const restoreTrashItem = async (id: number): Promise<void> => {
    await apiBase.post(`/trash/${id}/restore`);
};

export const deleteTrashItem = async (id: number): Promise<void> => {
    await apiBase.delete(`/trash/${id}`);
};

export const emptyTrash = async (): Promise<void> => {
    await apiBase.delete('/trash');
};

export const getTrashRetention = async (): Promise<TrashRetention> => {
    const response = await apiBase.get<TrashRetention>('/trash/retention');
    return response.data;
};

export const updateTrashRetention = async (days: number): Promise<TrashRetention> => {
    const response = await apiBase.put<TrashRetention>('/trash/retention', { days });
    return response.data;
};

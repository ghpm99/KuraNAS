import type { ImageAlbum, ImageAlbumItemsChange, ImageAlbumUpdate } from '@/types/imageAlbum';
import type { Pagination } from '@/types/pagination';
import { apiBase } from '.';

export const imageAlbumsPageSize = 60;

export const getImageAlbums = async (
    page: number,
    pageSize: number = imageAlbumsPageSize
): Promise<Pagination<ImageAlbum>> => {
    const response = await apiBase.get<Pagination<ImageAlbum>>('/image/albums', {
        params: { page, page_size: pageSize },
    });
    return response.data;
};

export const getImageAlbum = async (albumId: number): Promise<ImageAlbum> => {
    const response = await apiBase.get<ImageAlbum>(`/image/albums/${albumId}`);
    return response.data;
};

export const createImageAlbum = async (name: string): Promise<ImageAlbum> => {
    const response = await apiBase.post<ImageAlbum>('/image/albums', { name });
    return response.data;
};

export const updateImageAlbum = async (
    albumId: number,
    update: ImageAlbumUpdate
): Promise<ImageAlbum> => {
    const response = await apiBase.put<ImageAlbum>(`/image/albums/${albumId}`, update);
    return response.data;
};

export const deleteImageAlbum = async (albumId: number): Promise<void> => {
    await apiBase.delete(`/image/albums/${albumId}`);
};

export const addImageAlbumItems = async (
    albumId: number,
    fileIds: number[]
): Promise<ImageAlbumItemsChange> => {
    const response = await apiBase.post<ImageAlbumItemsChange>(`/image/albums/${albumId}/items`, {
        file_ids: fileIds,
    });
    return response.data;
};

export const removeImageAlbumItems = async (
    albumId: number,
    fileIds: number[]
): Promise<ImageAlbumItemsChange> => {
    const response = await apiBase.delete<ImageAlbumItemsChange>(`/image/albums/${albumId}/items`, {
        data: { file_ids: fileIds },
    });
    return response.data;
};

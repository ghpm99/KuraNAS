import type {
    FileData,
    FileListCategoryType,
    PaginationResponse as FilePaginationResponse,
    RecentAccessFile,
} from '@/features/files/providers/fileProvider/fileContext';
import type { IImageData, ImageGroupBy } from '@/components/providers/imageProvider/imageProvider';
import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import { Pagination } from '@/types/pagination';
import { apiBase } from '.';

type FilesTreeParams = {
    page: number;
    pageSize: number;
    fileParent?: number;
    category: FileListCategoryType;
};

export const getFilesTree = async ({
    page,
    pageSize,
    fileParent,
    category,
}: FilesTreeParams): Promise<FilePaginationResponse> => {
    const response = await apiBase.get<FilePaginationResponse>('/files/tree', {
        params: {
            page,
            page_size: pageSize,
            file_parent: fileParent,
            category,
        },
    });
    return response.data;
};

export const getRecentAccessByFileId = async (fileId: number): Promise<RecentAccessFile[]> => {
    const response = await apiBase.get<RecentAccessFile[]>(`/files/recent/${fileId}`);
    return response.data;
};

export const getFileByPath = async (path: string): Promise<FileData | null> => {
    const response = await apiBase.get<FilePaginationResponse>('/files/path', {
        params: { path },
    });

    return response.data.items[0] ?? null;
};

export const toggleStarredFile = async (itemId: number): Promise<void> => {
    await apiBase.post(`/files/starred/${itemId}`);
};

export const rescanFiles = async (): Promise<void> => {
    const formData = new FormData();
    formData.append('data', 'manual-rescan');

    await apiBase.post('/files/update', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
};

export const uploadFiles = async (files: FileList, targetFolderId?: number): Promise<void> => {
    const formData = new FormData();

    for (const file of Array.from(files)) {
        formData.append('files', file);
    }

    if (targetFolderId !== undefined && targetFolderId > 0) {
        formData.append('target_folder_id', String(targetFolderId));
    }

    await apiBase.post('/files/upload', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
};

export const createFolder = async (name: string, parentId?: number): Promise<void> => {
    await apiBase.post('/files/folder', {
        name,
        parent_id: parentId ?? null,
    });
};

export const deleteFile = async (id: number): Promise<void> => {
    await apiBase.delete('/files/path', {
        data: { id },
    });
};

export const renameFile = async (id: number, newName: string): Promise<void> => {
    await apiBase.post('/files/rename', {
        id,
        new_name: newName,
    });
};

export const moveFile = async (
    sourceId: number,
    destinationFolderId?: number,
    destinationPath?: string
): Promise<void> => {
    await apiBase.post('/files/move', {
        source_id: sourceId,
        destination_folder_id: destinationFolderId ?? null,
        destination_path: destinationPath ?? '',
    });
};

export const copyFile = async (
    sourceId: number,
    destinationFolderId?: number,
    destinationPath?: string,
    newName?: string
): Promise<void> => {
    await apiBase.post('/files/copy', {
        source_id: sourceId,
        destination_folder_id: destinationFolderId ?? null,
        destination_path: destinationPath ?? '',
        new_name: newName ?? '',
    });
};

export const downloadFileBlob = async (fileId: number): Promise<Blob> => {
    const response = await apiBase.get<Blob>(`/files/blob/${fileId}`, {
        responseType: 'blob',
    });
    return response.data;
};

export const getMusicFiles = async (
    page: number,
    pageSize: number
): Promise<Pagination<IMusicData>> => {
    const response = await apiBase.get<Pagination<IMusicData>>('/files/music', {
        params: { page, page_size: pageSize },
    });
    return response.data;
};

export const getImageFiles = async (
    page: number,
    pageSize: number,
    groupBy: ImageGroupBy
): Promise<Pagination<IImageData>> => {
    const response = await apiBase.get<Pagination<IImageData>>('/files/images', {
        params: { page, page_size: pageSize, group_by: groupBy },
    });
    return response.data;
};

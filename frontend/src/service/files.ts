import type {
    FileData,
    FilesSort,
    FileListCategoryType,
    PaginationResponse as FilePaginationResponse,
    RecentAccessFile,
} from '@/features/files/providers/fileProvider/fileContext';
import type { IImageData, ImageGroupBy } from '@/components/providers/imageProvider/imageProvider';
import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import type { FileAncestor } from '@/types/fileAncestor';
import type { FileLocation } from '@/types/fileLocation';
import type { FolderStats } from '@/types/folderStats';
import { Pagination } from '@/types/pagination';
import { apiBase } from '.';
import { getApiV1BaseUrl } from './apiUrl';

type FilesTreeParams = {
    page: number;
    pageSize: number;
    fileParent?: number;
    category: FileListCategoryType;
    sort?: FilesSort;
};

export const getFilesTree = async ({
    page,
    pageSize,
    fileParent,
    category,
    sort,
}: FilesTreeParams): Promise<FilePaginationResponse> => {
    const response = await apiBase.get<FilePaginationResponse>('/files/tree', {
        params: {
            page,
            page_size: pageSize,
            file_parent: fileParent,
            category,
            sort: sort?.key,
            order: sort?.order,
        },
    });
    return response.data;
};

type SearchFilesParams = {
    q: string;
    parentId?: number;
    recursive?: boolean;
    page: number;
    pageSize: number;
};

export const searchFiles = async ({
    q,
    parentId,
    recursive,
    page,
    pageSize,
}: SearchFilesParams): Promise<FilePaginationResponse> => {
    const response = await apiBase.get<FilePaginationResponse>('/files/search', {
        params: {
            q,
            parent_id: parentId,
            recursive,
            page,
            page_size: pageSize,
        },
    });
    return response.data;
};

type FilesPageParams = {
    page: number;
    pageSize: number;
};

export const getStarredFiles = async ({
    page,
    pageSize,
}: FilesPageParams): Promise<FilePaginationResponse> => {
    const response = await apiBase.get<FilePaginationResponse>('/files/starred', {
        params: { page, page_size: pageSize },
    });
    return response.data;
};

export const getRecentlyAccessedFiles = async ({
    page,
    pageSize,
}: FilesPageParams): Promise<FilePaginationResponse> => {
    const response = await apiBase.get<FilePaginationResponse>('/files/recent-files', {
        params: { page, page_size: pageSize },
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

export const getFileLocation = async (fileId: number): Promise<FileLocation> => {
    const response = await apiBase.get<FileLocation>(`/files/location/${fileId}`);
    return response.data;
};

export const getFolderStats = async (folderId: number): Promise<FolderStats> => {
    const response = await apiBase.get<FolderStats>(`/files/folder-stats/${folderId}`);
    return response.data;
};

export const getFileAncestors = async (fileId: number): Promise<FileAncestor[]> => {
    const response = await apiBase.get<FileAncestor[]>(`/files/ancestors/${fileId}`);
    return response.data;
};

export const getFileByDiskPath = async (diskPath: string): Promise<FileData> => {
    const response = await apiBase.get<FileData>('/files/by-disk-path', {
        params: { path: diskPath },
    });
    return response.data;
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

export type UploadConflictPolicy = 'rename' | 'replace' | 'skip';

export type UploadOutcomeStatus = 'uploaded' | 'skipped' | 'replaced' | 'renamed' | 'failed';

export type UploadOutcome = {
    status: UploadOutcomeStatus;
    name?: string;
    error?: string;
};

export type UploadSingleFileParams = {
    file: File;
    targetFolderId?: number;
    relativePath?: string;
    onConflict: UploadConflictPolicy;
    signal?: AbortSignal;
    onProgress?: (percent: number) => void;
};

const toProgressPercent = (loaded: number, total?: number): number =>
    total && total > 0 ? Math.min(100, Math.round((loaded / total) * 100)) : 0;

export const uploadSingleFile = async ({
    file,
    targetFolderId,
    relativePath,
    onConflict,
    signal,
    onProgress,
}: UploadSingleFileParams): Promise<UploadOutcome> => {
    const formData = new FormData();
    formData.append('files', file);
    formData.append('on_conflict', onConflict);

    if (targetFolderId !== undefined && targetFolderId > 0) {
        formData.append('target_folder_id', String(targetFolderId));
    }
    if (relativePath) {
        formData.append('relative_paths', relativePath);
    }

    const response = await apiBase.post('/files/upload', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
        signal,
        onUploadProgress: (event: { loaded: number; total?: number }) =>
            onProgress?.(toProgressPercent(event.loaded, event.total)),
    });

    const fileResult = response?.data?.files?.[0];
    if (!fileResult) {
        return { status: 'uploaded' };
    }
    return { status: fileResult.status, name: fileResult.name, error: fileResult.error };
};

export const createFolder = async (name: string, parentId?: number): Promise<void> => {
    await apiBase.post('/files/folder', {
        name,
        parent_id: parentId ?? null,
    });
};

export const deleteFile = async (id: number, permanent = false): Promise<void> => {
    await apiBase.delete('/files/path', {
        data: { id },
        ...(permanent ? { params: { permanent: true } } : {}),
    });
};

export const renameFile = async (id: number, newName: string): Promise<string> => {
    const response = await apiBase.post('/files/rename', {
        id,
        new_name: newName,
    });
    return response?.data?.path ?? '';
};

export const moveFile = async (
    sourceId: number,
    destinationFolderId?: number,
    destinationPath?: string
): Promise<string> => {
    const response = await apiBase.post('/files/move', {
        source_id: sourceId,
        destination_folder_id: destinationFolderId ?? null,
        destination_path: destinationPath ?? '',
    });
    return response?.data?.path ?? '';
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

export const textPreviewMaxBytes = 512 * 1024;

export const getFileTextPreview = async (fileId: number): Promise<string> => {
    const response = await apiBase.get<string>(`/files/blob/${fileId}`, {
        headers: { Range: `bytes=0-${textPreviewMaxBytes - 1}` },
        responseType: 'text',
        transformResponse: (rawBody: string) => rawBody,
    });
    return response.data;
};

export const getFileBlobUrl = (fileId: number): string =>
    `${getApiV1BaseUrl()}/files/blob/${fileId}`;

export const getFileThumbnailUrl = (fileId: number): string =>
    `${getApiV1BaseUrl()}/files/thumbnail/${fileId}`;

export const getFileDownloadUrl = (fileId: number): string =>
    `${getApiV1BaseUrl()}/files/download/${fileId}`;

export const getFilesZipDownloadUrl = (fileIds: number[]): string =>
    `${getApiV1BaseUrl()}/files/download-zip?ids=${fileIds.join(',')}`;

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

export const getPendingImageClassificationCount = async (): Promise<number> => {
    const response = await apiBase.get<{ pending_count: number }>(
        '/files/images/classification/pending-count'
    );
    return response.data.pending_count;
};

export const startImageClassificationBackfill = async (): Promise<number> => {
    const response = await apiBase.post<{ job_id: number; message: string }>(
        '/files/images/classification/backfill'
    );
    return response.data.job_id;
};

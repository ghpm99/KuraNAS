import type { Pagination } from '@/types/pagination';
import { apiBase } from '.';

export type DocumentFormat = 'pdf' | 'text' | 'spreadsheet' | 'presentation' | 'other' | 'image';
export type DocumentIndexStateName = 'indexed' | 'partial' | 'name_only' | 'queued' | 'failed';
export type DocumentSortField = 'modified_at' | 'name' | 'size' | 'created_at';
export type DocumentSortOrder = 'asc' | 'desc';

export interface DocumentFilters {
    format?: DocumentFormat;
    folderId?: number;
    modifiedFrom?: string;
    modifiedTo?: string;
    favoritesOnly?: boolean;
    indexState?: DocumentIndexStateName;
}

export interface DocumentSort {
    sort?: DocumentSortField;
    order?: DocumentSortOrder;
}

export interface DocumentListItem {
    id: number;
    name: string;
    format: string;
    parent_path: string;
    size: number;
    modified_at: string;
    created_at: string;
    starred: boolean;
    index_state: DocumentIndexStateName;
}

export interface DocumentSearchResult {
    file_id: number;
    name: string;
    path: string;
    parent_path: string;
    format: string;
    size: number;
    updated_at: string;
    snippet: string;
    page: number | null;
}

export interface DocumentIndexState {
    file_id: number;
    state: DocumentIndexStateName;
    indexed_bytes: number;
    reason: string;
    error: string;
}

export interface DocumentText {
    text: string;
    is_partial: boolean;
}

export interface DocumentCount {
    total: number;
}

export interface DocumentReindexJob {
    job_id: number;
}

export type DocumentListResponse = Pagination<DocumentListItem>;
export type DocumentSearchResponse = Pagination<DocumentSearchResult>;

type ListDocumentsParams = {
    filters?: DocumentFilters;
    sort?: DocumentSort;
    page: number;
    pageSize: number;
    signal?: AbortSignal;
};

type SearchDocumentsParams = {
    q: string;
    filters?: DocumentFilters;
    page: number;
    pageSize: number;
    signal?: AbortSignal;
};

const buildFilterParams = (filters: DocumentFilters = {}): Record<string, string | number> => {
    const params: Record<string, string | number> = {};
    if (filters.format) params.format = filters.format;
    if (filters.folderId !== undefined) params.folder_id = filters.folderId;
    if (filters.modifiedFrom) params.modified_from = filters.modifiedFrom;
    if (filters.modifiedTo) params.modified_to = filters.modifiedTo;
    if (filters.favoritesOnly) params.favorites_only = 'true';
    if (filters.indexState) params.index_state = filters.indexState;
    return params;
};

const buildSortParams = (sort: DocumentSort = {}): Record<string, string> => {
    const params: Record<string, string> = {};
    if (sort.sort) params.sort = sort.sort;
    if (sort.order) params.order = sort.order;
    return params;
};

export const listDocuments = async ({
    filters,
    sort,
    page,
    pageSize,
    signal,
}: ListDocumentsParams): Promise<DocumentListResponse> => {
    const response = await apiBase.get<DocumentListResponse>('/documents', {
        params: {
            ...buildFilterParams(filters),
            ...buildSortParams(sort),
            page,
            page_size: pageSize,
        },
        signal,
    });
    return response.data;
};

export const countDocuments = async (filters?: DocumentFilters): Promise<DocumentCount> => {
    const response = await apiBase.get<DocumentCount>('/documents/count', {
        params: buildFilterParams(filters),
    });
    return response.data;
};

export const searchDocuments = async ({
    q,
    filters,
    page,
    pageSize,
    signal,
}: SearchDocumentsParams): Promise<DocumentSearchResponse> => {
    const response = await apiBase.get<DocumentSearchResponse>('/documents/search', {
        params: { q, ...buildFilterParams(filters), page, page_size: pageSize },
        signal,
    });
    return response.data;
};

export const getDocumentIndexState = async (fileId: number): Promise<DocumentIndexState> => {
    const response = await apiBase.get<DocumentIndexState>(`/documents/${fileId}/index-state`);
    return response.data;
};

export const reindexDocument = async (fileId: number): Promise<DocumentReindexJob> => {
    const response = await apiBase.post<DocumentReindexJob>(`/documents/${fileId}/reindex`);
    return response.data;
};

export const getDocumentText = async (fileId: number): Promise<DocumentText> => {
    const response = await apiBase.get<DocumentText>(`/documents/${fileId}/text`);
    return response.data;
};

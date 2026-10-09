import { apiBase } from '.';

export interface DocumentSearchResult {
    file_id: number;
    name: string;
    path: string;
    parent_path: string;
    format: string;
    size: number;
    updated_at: string;
    snippet: string;
}

export interface DocumentSearchPagination {
    page: number;
    page_size: number;
    has_next: boolean;
    has_prev: boolean;
}

export interface DocumentSearchResponse {
    items: DocumentSearchResult[];
    pagination: DocumentSearchPagination;
}

type SearchDocumentsParams = {
    q: string;
    page: number;
    pageSize: number;
    signal?: AbortSignal;
};

export const searchDocuments = async ({
    q,
    page,
    pageSize,
    signal,
}: SearchDocumentsParams): Promise<DocumentSearchResponse> => {
    const response = await apiBase.get<DocumentSearchResponse>('/documents/search', {
        params: { q, page, page_size: pageSize },
        signal,
    });
    return response.data;
};

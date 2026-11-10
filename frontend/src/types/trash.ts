import type { Pagination } from '@/types/pagination';

export type TrashItem = {
    id: number;
    name?: string;
    is_directory?: boolean;
    expires_at?: string;
    original_path: string;
    size: number;
    deleted_at: string;
};

export type TrashPage = Pagination<TrashItem>;

export type TrashRetention = {
    days: number;
};

export type TrashSummary = {
    items: number;
    size_bytes: number;
};

export type TrashSortKey = 'deleted_at' | 'name' | 'size';

export type TrashSortOrder = 'asc' | 'desc';

export type TrashListing = {
    sortKey: TrashSortKey;
    order: TrashSortOrder;
    searchText: string;
};

export type TrashBatchFailure = {
    id: number;
    error: string;
    path?: string;
};

export type TrashBatchRestoreOutcome = {
    restored_ids?: number[];
    failures?: TrashBatchFailure[];
};

export type TrashBatchDeleteOutcome = {
    deleted_ids?: number[];
    failures?: TrashBatchFailure[];
};

export type TrashRestoreResult = {
    path?: string;
};

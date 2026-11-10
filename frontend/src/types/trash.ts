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

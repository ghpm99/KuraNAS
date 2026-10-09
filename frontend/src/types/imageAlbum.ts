export type ImageAlbum = {
    id: number;
    name: string;
    cover_file_id: number | null;
    item_count: number;
    created_at: string;
    updated_at: string;
};

export type ImageAlbumUpdate = {
    name?: string;
    cover_file_id?: number;
};

export type ImageAlbumItemsChange = {
    requested: number;
    changed: number;
};

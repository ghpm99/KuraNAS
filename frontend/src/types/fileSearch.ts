export const fileSearchKinds = [
    'folder',
    'document',
    'image',
    'audio',
    'video',
    'archive',
    'other',
] as const;

export type FileSearchKind = (typeof fileSearchKinds)[number];
export type FileSearchTier = 'hot' | 'cold';
export type FileSearchSort = 'relevance' | 'name' | 'size' | 'modified';
export type FileSearchOrder = 'asc' | 'desc';

export type FileSearchRefinements = {
    kinds?: FileSearchKind[];
    modifiedFrom?: string;
    modifiedTo?: string;
    minSize?: number;
    maxSize?: number;
    tier?: FileSearchTier;
    starred?: boolean;
    sort?: FileSearchSort;
    order?: FileSearchOrder;
};

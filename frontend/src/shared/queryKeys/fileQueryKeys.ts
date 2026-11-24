export const documentsQueryKey = 'documents';

export const fileQueryKeys = ['files', 'files-path', 'filesRecent', documentsQueryKey];
export const ancestorsQueryKey = 'files-ancestors';
export const searchQueryKey = 'files-search';

export const trashQueryKey = 'trash';

export const allFileQueryKeys = [
    ...fileQueryKeys,
    ancestorsQueryKey,
    searchQueryKey,
    trashQueryKey,
];

export const homeFavoritesQueryKey = ['home', 'favorites'] as const;
export const starredCountsQueryKey = ['files-starred-counts'] as const;

import type { GlobalSearchResponse } from '@/service/search';

type StarrableSearchEntry = { id: number; starred: boolean };

const withStarredEntry = <Entry extends StarrableSearchEntry>(
    entries: Entry[] | undefined,
    fileId: number,
    isStarred: boolean
): Entry[] | undefined =>
    entries?.map((entry) => (entry.id === fileId ? { ...entry, starred: isStarred } : entry));

export const applySearchStarToCache = (
    cachedResponse: GlobalSearchResponse | undefined,
    fileId: number,
    isStarred: boolean
): GlobalSearchResponse | undefined => {
    if (!cachedResponse) {
        return cachedResponse;
    }
    return {
        ...cachedResponse,
        files: withStarredEntry(cachedResponse.files, fileId, isStarred) ?? cachedResponse.files,
        folders:
            withStarredEntry(cachedResponse.folders, fileId, isStarred) ?? cachedResponse.folders,
    };
};

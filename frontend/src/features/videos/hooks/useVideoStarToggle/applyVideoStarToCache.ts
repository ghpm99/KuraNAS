const isVideoRecord = (candidate: Record<string, unknown>, videoId: number): boolean =>
    candidate.id === videoId &&
    typeof candidate.name === 'string' &&
    typeof candidate.path === 'string' &&
    typeof candidate.format === 'string';

export const applyVideoStarToCache = <CachedData>(
    cachedData: CachedData,
    videoId: number,
    isStarred: boolean
): CachedData => {
    if (Array.isArray(cachedData)) {
        return cachedData.map((entry) =>
            applyVideoStarToCache(entry, videoId, isStarred)
        ) as CachedData;
    }
    if (cachedData === null || typeof cachedData !== 'object') {
        return cachedData;
    }
    const record = cachedData as Record<string, unknown>;
    if (isVideoRecord(record, videoId)) {
        return { ...record, starred: isStarred } as CachedData;
    }
    return Object.fromEntries(
        Object.entries(record).map(([field, fieldValue]) => [
            field,
            applyVideoStarToCache(fieldValue, videoId, isStarred),
        ])
    ) as CachedData;
};

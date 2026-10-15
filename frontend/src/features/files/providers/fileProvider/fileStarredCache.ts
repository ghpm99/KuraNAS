import type { InfiniteData } from '@tanstack/react-query';
import type { FileData, FileListCategoryType, PaginationResponse } from './fileContext';

export type FileListingCache = InfiniteData<PaginationResponse, unknown>;

type FileListingQueryKey = readonly unknown[];

const isStarredRootListing = (queryKey: FileListingQueryKey): boolean => {
    const queryParams = queryKey[1] as { file_parent?: number } | undefined;
    const listingFilter = queryKey[2] as FileListCategoryType | undefined;
    return listingFilter === 'starred' && queryParams?.file_parent === undefined;
};

const withStarredState = (
    file: FileData,
    fileIds: readonly number[],
    isStarred: boolean
): FileData => {
    if (!fileIds.includes(file.id)) return file;
    return { ...file, starred: isStarred };
};

export const applyStarredToListing = (
    cache: FileListingCache | undefined,
    queryKey: FileListingQueryKey,
    fileIds: readonly number[],
    isStarred: boolean
): FileListingCache | undefined => {
    if (!cache?.pages) return cache;
    const shouldDropUnstarred = isStarredRootListing(queryKey);
    return {
        ...cache,
        pages: cache.pages.map((page) => ({
            ...page,
            items: (page?.items ?? [])
                .map((file) => withStarredState(file, fileIds, isStarred))
                .filter((file) => !(shouldDropUnstarred && fileIds.includes(file.id) && !file.starred)),
        })),
    };
};

export const applyStarredToFile = (
    file: FileData | null | undefined,
    fileIds: readonly number[],
    isStarred: boolean
): FileData | null | undefined => {
    if (!file) return file;
    return withStarredState(file, fileIds, isStarred);
};

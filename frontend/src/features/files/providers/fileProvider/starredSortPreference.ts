import type { StarredSort } from '@/service/files';

export const defaultStarredSort: StarredSort = 'name';

export const starredSortStorageKey = 'kuranas.favorites.sort';

const validStarredSorts: StarredSort[] = ['name', 'starred_at'];

export const loadStarredSort = (): StarredSort => {
    try {
        const storedSort = window.localStorage.getItem(starredSortStorageKey);
        return validStarredSorts.find((sort) => sort === storedSort) ?? defaultStarredSort;
    } catch {
        return defaultStarredSort;
    }
};

export const saveStarredSort = (sort: StarredSort): void => {
    try {
        window.localStorage.setItem(starredSortStorageKey, sort);
    } catch {
        return;
    }
};

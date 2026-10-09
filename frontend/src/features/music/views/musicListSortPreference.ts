import type { MusicListSort, MusicListSortField, MusicListSortOrder } from '@/types/music';

export type MusicSortableView = 'artists' | 'albums' | 'genres' | 'folders';

export const musicSortFieldsByView: Record<MusicSortableView, MusicListSortField[]> = {
    artists: ['tracks', 'name', 'recent'],
    albums: ['tracks', 'name', 'recent', 'year'],
    genres: ['tracks', 'name', 'recent'],
    folders: ['tracks', 'name', 'recent'],
};

const validSortOrders: MusicListSortOrder[] = ['asc', 'desc'];

export const getDefaultOrderOfField = (field: MusicListSortField): MusicListSortOrder =>
    field === 'name' ? 'asc' : 'desc';

export const defaultMusicListSort: MusicListSort = { sort: 'tracks', order: 'desc' };

export const getMusicListSortStorageKey = (view: MusicSortableView) => `kuranas.music.${view}.sort`;

const isMusicListSortOfView = (candidate: unknown, view: MusicSortableView): boolean => {
    if (typeof candidate !== 'object' || candidate === null) {
        return false;
    }
    const { sort, order } = candidate as Partial<MusicListSort>;
    return (
        musicSortFieldsByView[view].includes(sort as MusicListSortField) &&
        validSortOrders.includes(order as MusicListSortOrder)
    );
};

export const loadMusicListSort = (view: MusicSortableView): MusicListSort => {
    try {
        const storedSort = window.localStorage.getItem(getMusicListSortStorageKey(view));
        if (!storedSort) {
            return defaultMusicListSort;
        }
        const parsedSort: unknown = JSON.parse(storedSort);
        return isMusicListSortOfView(parsedSort, view)
            ? (parsedSort as MusicListSort)
            : defaultMusicListSort;
    } catch {
        return defaultMusicListSort;
    }
};

export const saveMusicListSort = (view: MusicSortableView, listSort: MusicListSort): void => {
    try {
        window.localStorage.setItem(getMusicListSortStorageKey(view), JSON.stringify(listSort));
    } catch {
        return;
    }
};

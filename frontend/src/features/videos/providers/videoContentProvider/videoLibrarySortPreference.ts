import type {
    VideoLibrarySort,
    VideoLibrarySortKey,
    VideoLibrarySortOrder,
} from '@/service/videoPlayback';

export const defaultVideoLibrarySort: VideoLibrarySort = { key: 'recent', order: 'desc' };

export const videoLibrarySortStorageKey = 'kuranas.videos.librarySort';

export const videoLibrarySortKeys: VideoLibrarySortKey[] = ['recent', 'name', 'size', 'duration'];
const validSortOrders: VideoLibrarySortOrder[] = ['asc', 'desc'];

const isVideoLibrarySort = (candidate: unknown): candidate is VideoLibrarySort => {
    if (typeof candidate !== 'object' || candidate === null) {
        return false;
    }
    const { key, order } = candidate as Partial<VideoLibrarySort>;
    return (
        videoLibrarySortKeys.includes(key as VideoLibrarySortKey) &&
        validSortOrders.includes(order as VideoLibrarySortOrder)
    );
};

export const loadVideoLibrarySort = (): VideoLibrarySort => {
    try {
        const storedSort = window.localStorage.getItem(videoLibrarySortStorageKey);
        if (!storedSort) {
            return defaultVideoLibrarySort;
        }
        const parsedSort: unknown = JSON.parse(storedSort);
        return isVideoLibrarySort(parsedSort) ? parsedSort : defaultVideoLibrarySort;
    } catch {
        return defaultVideoLibrarySort;
    }
};

export const saveVideoLibrarySort = (sort: VideoLibrarySort): void => {
    try {
        window.localStorage.setItem(videoLibrarySortStorageKey, JSON.stringify(sort));
    } catch {
        return;
    }
};

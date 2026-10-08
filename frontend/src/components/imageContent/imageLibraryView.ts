import type { ImageSection } from '@/app/routes';
import {
    imageFormats,
    type ImageCategory,
    type ImageLibraryFilters,
    type ImageLibraryOrdering,
    type ImageLibrarySort,
    type ImageLibrarySortOrder,
} from '@/types/imageLibrary';

export type ImageAlbumPresetId = 'documents' | 'memes' | 'art' | 'landscapes' | 'portraits';

export type ImageAlbumPreset = {
    id: ImageAlbumPresetId;
    categories: ImageCategory[];
    titleKey: string;
    descriptionKey: string;
};

export const imageAlbumPresets: ImageAlbumPreset[] = [
    {
        id: 'documents',
        categories: ['document', 'receipt'],
        titleKey: 'IMAGES_ALBUM_DOCUMENTS',
        descriptionKey: 'IMAGES_ALBUM_DOCUMENTS_DESCRIPTION',
    },
    {
        id: 'memes',
        categories: ['meme'],
        titleKey: 'IMAGES_ALBUM_MEMES',
        descriptionKey: 'IMAGES_ALBUM_MEMES_DESCRIPTION',
    },
    {
        id: 'art',
        categories: ['art'],
        titleKey: 'IMAGES_ALBUM_ART',
        descriptionKey: 'IMAGES_ALBUM_ART_DESCRIPTION',
    },
    {
        id: 'landscapes',
        categories: ['landscape'],
        titleKey: 'IMAGES_ALBUM_LANDSCAPES',
        descriptionKey: 'IMAGES_ALBUM_LANDSCAPES_DESCRIPTION',
    },
    {
        id: 'portraits',
        categories: ['portrait'],
        titleKey: 'IMAGES_ALBUM_PORTRAITS',
        descriptionKey: 'IMAGES_ALBUM_PORTRAITS_DESCRIPTION',
    },
];

export const imageSearchParamNames = {
    query: 'q',
    takenFrom: 'from',
    takenTo: 'to',
    format: 'format',
    sort: 'sort',
    order: 'order',
    jumpBefore: 'before',
    folder: 'folder',
    album: 'album',
} as const;

export type ImageLibraryView = {
    section: ImageSection;
    filters: ImageLibraryFilters;
    ordering: ImageLibraryOrdering;
    takenBefore: string;
    selectedFolder: string;
    selectedAlbum: ImageAlbumPreset | null;
    hasUserFilters: boolean;
    isKeyset: boolean;
};

const recentWindowInDays = 30;
const millisecondsPerDay = 24 * 60 * 60 * 1000;
const dateOnlyPattern = /^\d{4}-\d{2}-\d{2}$/;
const validSorts: ImageLibrarySort[] = ['taken_at', 'name', 'size'];
const validOrders: ImageLibrarySortOrder[] = ['asc', 'desc'];

const sectionCategories: Partial<Record<ImageSection, ImageCategory[]>> = {
    captures: ['capture', 'screenshot_app'],
    photos: ['photo', 'landscape', 'portrait'],
};

export const getDefaultSortOrder = (sort: ImageLibrarySort): ImageLibrarySortOrder =>
    sort === 'name' ? 'asc' : 'desc';

export const isKeysetOrdering = (ordering: ImageLibraryOrdering) =>
    ordering.sort === 'taken_at' && ordering.order === 'desc';

const toDateOnly = (instant: Date) => instant.toISOString().slice(0, 10);

const readDateOnly = (rawValue: string | null) =>
    rawValue && dateOnlyPattern.test(rawValue) ? rawValue : '';

const readOrdering = (searchParams: URLSearchParams): ImageLibraryOrdering => {
    const rawSort = searchParams.get(imageSearchParamNames.sort) as ImageLibrarySort | null;
    const sort = rawSort && validSorts.includes(rawSort) ? rawSort : 'taken_at';
    const rawOrder = searchParams.get(imageSearchParamNames.order) as ImageLibrarySortOrder | null;
    const order = rawOrder && validOrders.includes(rawOrder) ? rawOrder : getDefaultSortOrder(sort);
    return { sort, order };
};

const readFormats = (searchParams: URLSearchParams) =>
    searchParams
        .getAll(imageSearchParamNames.format)
        .filter((format) => (imageFormats as readonly string[]).includes(format));

export const parseImageLibraryView = (
    section: ImageSection,
    searchParams: URLSearchParams,
    now: Date = new Date()
): ImageLibraryView => {
    const nameQuery = searchParams.get(imageSearchParamNames.query)?.trim() ?? '';
    const userTakenFrom = readDateOnly(searchParams.get(imageSearchParamNames.takenFrom));
    const takenTo = readDateOnly(searchParams.get(imageSearchParamNames.takenTo));
    const formats = readFormats(searchParams);
    const ordering = readOrdering(searchParams);
    const isKeyset = isKeysetOrdering(ordering);

    const selectedFolder =
        section === 'folders' ? (searchParams.get(imageSearchParamNames.folder) ?? '') : '';
    const albumId = searchParams.get(imageSearchParamNames.album);
    const selectedAlbum =
        section === 'albums'
            ? (imageAlbumPresets.find((preset) => preset.id === albumId) ?? null)
            : null;

    const recentTakenFrom =
        section === 'recent'
            ? toDateOnly(new Date(now.getTime() - recentWindowInDays * millisecondsPerDay))
            : '';

    const filters: ImageLibraryFilters = {
        nameQuery,
        categories: selectedAlbum?.categories ?? sectionCategories[section] ?? [],
        isStarredOnly: section === 'favorites',
        formats,
        takenFrom: userTakenFrom || recentTakenFrom,
        takenTo,
        folder: selectedFolder,
    };

    return {
        section,
        filters,
        ordering,
        takenBefore: isKeyset
            ? readDateOnly(searchParams.get(imageSearchParamNames.jumpBefore))
            : '',
        selectedFolder,
        selectedAlbum,
        hasUserFilters: Boolean(nameQuery || userTakenFrom || takenTo || formats.length > 0),
        isKeyset,
    };
};

export const buildJumpBeforeDate = (year: number, month: number) =>
    toDateOnly(new Date(Date.UTC(year, month, 1)));

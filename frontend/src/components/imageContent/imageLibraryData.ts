import type { ImageSection } from '@/app/routes';
import type { IImageData } from '@/components/providers/imageProvider/imageProvider';

export type AutomaticAlbumKey = 'travel' | 'documents' | 'wallpapers' | 'memes' | 'others';

export type ImageCollection = {
    id: string;
    cover: IImageData | null;
    images: IImageData[];
    latestDate: Date | null;
};

const recentWindowInMs = 1000 * 60 * 60 * 24 * 30;

const pathKeywordGroups = {
    travel: [
        'travel',
        'trip',
        'vacation',
        'holiday',
        'journey',
        'viagem',
        'ferias',
        'praia',
        'beach',
        'camp',
        'trail',
    ],
    documents: [
        'scan',
        'document',
        'receipt',
        'invoice',
        'contract',
        'doc',
        'boleto',
        'nota',
        'comprovante',
    ],
    memes: ['meme', 'reaction', 'sticker', 'funny', 'shitpost', 'whatsapp', 'discord', 'telegram'],
} as const;

const normalizePath = (value: string) => value.replace(/\\/g, '/').trim();

export const getImageDate = (image: IImageData): Date | null => {
    const candidates = [
        image.metadata?.datetime_original,
        image.metadata?.datetime,
        image.metadata?.createdAt,
        image.updated_at,
        image.created_at,
    ];
    for (const candidate of candidates) {
        if (!candidate) continue;
        const parsed = new Date(candidate);
        if (!Number.isNaN(parsed.getTime())) return parsed;
    }
    return null;
};

export const isRecentImage = (date: Date | null, now = Date.now()): boolean => {
    if (!date) return false;
    return date.getTime() >= now - recentWindowInMs;
};

export const hasPersistedCategory = (image: IImageData, category: 'capture' | 'photo'): boolean =>
    image.metadata?.classification?.category === category;

export const getImageDirectoryPath = (image: IImageData) => {
    const normalizedPath = normalizePath(image.path || '');
    if (!normalizedPath) {
        return '/';
    }

    const segments = normalizedPath.split('/').filter(Boolean);
    const lastSegment = segments.length > 0 ? segments[segments.length - 1]! : '';

    const normalizedName = image.name?.trim();
    if (
        normalizedName &&
        (normalizedPath === normalizedName || normalizedPath.endsWith(`/${normalizedName}`))
    ) {
        const lastSlashIndex = normalizedPath.lastIndexOf('/');
        return lastSlashIndex <= 0 ? '/' : normalizedPath.slice(0, lastSlashIndex);
    }

    if (lastSegment.includes('.')) {
        const lastSlashIndex = normalizedPath.lastIndexOf('/');
        return lastSlashIndex <= 0 ? '/' : normalizedPath.slice(0, lastSlashIndex);
    }

    return normalizedPath;
};

export const getCollectionTitleFromPath = (path: string) => {
    if (path === '/') {
        return '/';
    }

    const normalizedPath = normalizePath(path);
    const segments = normalizedPath.split('/').filter(Boolean);
    return segments.length > 0 ? segments[segments.length - 1]! : normalizedPath;
};

const matchesKeywords = (value: string, keywords: readonly string[]) =>
    keywords.some((keyword) => value.includes(keyword));

const getAlbumHeuristicSource = (image: IImageData) =>
    [
        image.name,
        getImageDirectoryPath(image),
        image.metadata?.image_description,
        image.metadata?.user_comment,
        image.metadata?.software,
    ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();

const isWallpaperCandidate = (image: IImageData) => {
    const width = image.metadata?.width ?? 0;
    const height = image.metadata?.height ?? 0;

    if (width < 1600 || height < 900) {
        return false;
    }

    return width >= height * 1.45;
};

export const getAutomaticAlbumKey = (image: IImageData): AutomaticAlbumKey => {
    const source = getAlbumHeuristicSource(image);

    if (matchesKeywords(source, pathKeywordGroups.memes)) {
        return 'memes';
    }

    if (matchesKeywords(source, pathKeywordGroups.documents)) {
        return 'documents';
    }

    if (matchesKeywords(source, pathKeywordGroups.travel) && hasPersistedCategory(image, 'photo')) {
        return 'travel';
    }

    if (isWallpaperCandidate(image)) {
        return 'wallpapers';
    }

    return 'others';
};

const sortCollections = (collections: ImageCollection[]) =>
    collections.sort((left, right) => {
        const rightDate = right.latestDate?.getTime() ?? 0;
        const leftDate = left.latestDate?.getTime() ?? 0;

        if (rightDate !== leftDate) {
            return rightDate - leftDate;
        }

        return right.id.localeCompare(left.id);
    });

export const buildFolderCollections = (images: IImageData[]): ImageCollection[] => {
    const collections = new Map<string, ImageCollection>();

    for (const image of images) {
        const folderPath = getImageDirectoryPath(image);
        const imageDate = getImageDate(image);
        const existing = collections.get(folderPath);

        if (!existing) {
            collections.set(folderPath, {
                id: folderPath,
                cover: image,
                images: [image],
                latestDate: imageDate,
            });
            continue;
        }

        existing.images.push(image);
        if ((imageDate?.getTime() ?? 0) >= (existing.latestDate?.getTime() ?? 0)) {
            existing.latestDate = imageDate;
            existing.cover = image;
        }
    }

    return sortCollections(Array.from(collections.values()));
};

export const buildAutomaticAlbumCollections = (images: IImageData[]) => {
    const albumOrder: AutomaticAlbumKey[] = [
        'travel',
        'documents',
        'wallpapers',
        'memes',
        'others',
    ];
    const collections = new Map<AutomaticAlbumKey, ImageCollection>(
        albumOrder.map((key) => [
            key,
            {
                id: key,
                cover: null,
                images: [],
                latestDate: null,
            },
        ])
    );

    for (const image of images) {
        const albumKey = getAutomaticAlbumKey(image);
        const imageDate = getImageDate(image);
        const collection = collections.get(albumKey);

        if (!collection) {
            continue;
        }

        collection.images.push(image);
        if (
            !collection.cover ||
            (imageDate?.getTime() ?? 0) >= (collection.latestDate?.getTime() ?? 0)
        ) {
            collection.cover = image;
            collection.latestDate = imageDate;
        }
    }

    return albumOrder.map((key) => collections.get(key)!);
};

export const matchesImageSearch = (image: IImageData, search: string) => {
    const searchValue = search.trim().toLowerCase();
    if (!searchValue) {
        return true;
    }

    const haystack = [
        image.name,
        image.path,
        image.format,
        image.metadata?.make,
        image.metadata?.model,
        image.metadata?.image_description,
    ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();

    return haystack.includes(searchValue);
};

export const matchesImageSection = (
    section: ImageSection,
    image: IImageData,
    date: Date | null
) => {
    switch (section) {
        case 'recent':
            return isRecentImage(date);
        case 'captures':
            return hasPersistedCategory(image, 'capture');
        case 'photos':
            return hasPersistedCategory(image, 'photo');
        case 'library':
        case 'folders':
        case 'albums':
        default:
            return true;
    }
};

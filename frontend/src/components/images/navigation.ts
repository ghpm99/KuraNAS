import { getImageRoute, type ImageSection } from '@/app/routes';

export type ImageNavigationItem = {
    key: ImageSection;
    labelKey: string;
    descriptionKey: string;
};

export const imageNavigationItems: ImageNavigationItem[] = [
    {
        key: 'library',
        labelKey: 'IMAGES_SECTION_LIBRARY',
        descriptionKey: 'IMAGES_SECTION_LIBRARY_DESCRIPTION',
    },
    {
        key: 'recent',
        labelKey: 'IMAGES_SECTION_RECENT',
        descriptionKey: 'IMAGES_SECTION_RECENT_DESCRIPTION',
    },
    {
        key: 'captures',
        labelKey: 'IMAGES_SECTION_CAPTURES',
        descriptionKey: 'IMAGES_SECTION_CAPTURES_DESCRIPTION',
    },
    {
        key: 'photos',
        labelKey: 'IMAGES_SECTION_PHOTOS',
        descriptionKey: 'IMAGES_SECTION_PHOTOS_DESCRIPTION',
    },
    {
        key: 'folders',
        labelKey: 'IMAGES_SECTION_FOLDERS',
        descriptionKey: 'IMAGES_SECTION_FOLDERS_DESCRIPTION',
    },
    {
        key: 'albums',
        labelKey: 'IMAGES_SECTION_ALBUMS',
        descriptionKey: 'IMAGES_SECTION_ALBUMS_DESCRIPTION',
    },
];

const imageRouteEntries = imageNavigationItems.map(
    (item) => [getImageRoute(item.key), item.key] as const
);

export const getImageSectionFromPath = (pathname: string): ImageSection => {
    if (pathname === getImageRoute('library')) {
        return 'library';
    }

    const matchedEntry = imageRouteEntries.find(([route, section]) => {
        if (section === 'library') {
            return false;
        }

        return pathname === route || pathname.startsWith(`${route}/`);
    });

    return matchedEntry?.[1] ?? 'library';
};

export const getImageSectionMeta = (section: ImageSection) => {
    const matchedItem = imageNavigationItems.find((item) => item.key === section);

    if (matchedItem) {
        return matchedItem;
    }

    return imageNavigationItems[0]!;
};

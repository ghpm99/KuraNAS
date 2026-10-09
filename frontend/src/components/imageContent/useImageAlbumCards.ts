import { useQueries } from '@tanstack/react-query';
import useI18n from '@/components/i18n/provider/i18nContext';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getImageLibraryCount, getImageLibraryPage } from '@/service/image';
import type { ImageLibraryFilters, ImageLibraryOrdering } from '@/types/imageLibrary';
import type { ImageCollectionCard } from './components/ImageCollectionsPanel';
import { imageAlbumPresets, type ImageAlbumPreset } from './imageLibraryView';

const newestFirst: ImageLibraryOrdering = { sort: 'taken_at', order: 'desc' };

const buildPresetFilters = (preset: ImageAlbumPreset): ImageLibraryFilters => ({
    nameQuery: '',
    categories: preset.categories,
    isStarredOnly: false,
    formats: [],
    camera: '',
    takenFrom: '',
    takenTo: '',
    folder: '',
});

export const useImageAlbumCards = (isEnabled: boolean): ImageCollectionCard[] => {
    const { t } = useI18n();

    const countQueries = useQueries({
        queries: imageAlbumPresets.map((preset) => ({
            queryKey: ['images', 'count', buildPresetFilters(preset)],
            queryFn: () => getImageLibraryCount(buildPresetFilters(preset)),
            enabled: isEnabled,
            staleTime: listingStaleTimeMs,
            refetchOnWindowFocus: false,
        })),
    });

    const coverQueries = useQueries({
        queries: imageAlbumPresets.map((preset) => ({
            queryKey: ['images', 'cover', preset.id],
            queryFn: () =>
                getImageLibraryPage({
                    filters: buildPresetFilters(preset),
                    ordering: newestFirst,
                    pageSize: 1,
                }),
            enabled: isEnabled,
            staleTime: listingStaleTimeMs,
            refetchOnWindowFocus: false,
        })),
    });

    return imageAlbumPresets.map((preset, presetIndex) => ({
        id: preset.id,
        title: t(preset.titleKey),
        description: t(preset.descriptionKey),
        imageCount: countQueries[presetIndex]?.data,
        coverImageId: coverQueries[presetIndex]?.data?.items?.[0]?.file_id,
    }));
};

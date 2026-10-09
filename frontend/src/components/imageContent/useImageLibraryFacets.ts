import { useQuery } from '@tanstack/react-query';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getImageCameraFacets, getImageFormatFacets } from '@/service/image';
import type { ImageCameraFacet, ImageFormatFacet, ImageLibraryFilters } from '@/types/imageLibrary';

const noCameraFacets: ImageCameraFacet[] = [];
const noFormatFacets: ImageFormatFacet[] = [];

export const useImageLibraryFacets = (filters: ImageLibraryFilters) => {
    const cameraScope: ImageLibraryFilters = { ...filters, camera: '' };
    const formatScope: ImageLibraryFilters = { ...filters, formats: [] };

    const cameraFacetsQuery = useQuery({
        queryKey: ['images', 'facets', 'cameras', cameraScope],
        queryFn: () => getImageCameraFacets(cameraScope),
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });
    const formatFacetsQuery = useQuery({
        queryKey: ['images', 'facets', 'formats', formatScope],
        queryFn: () => getImageFormatFacets(formatScope),
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    return {
        cameraFacets: cameraFacetsQuery.data ?? noCameraFacets,
        formatFacets: formatFacetsQuery.data ?? noFormatFacets,
    };
};

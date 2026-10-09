import type { ImageGroupBy, IImageData } from '@/types/image';
import type {
    ImageCameraFacet,
    ImageFormatFacet,
    ImageLibraryCount,
    ImageLibraryFolder,
    ImageLibraryFilters,
    ImageLibraryOrdering,
    ImageLibraryPage,
    ImageLibraryNeighbors,
    ImageMetadataSummary,
    ImageTimelineBucket,
} from '@/types/imageLibrary';
import { Pagination } from '@/types/pagination';
import { apiBase } from '.';

export type ImageLibraryPageRequest = {
    filters: ImageLibraryFilters;
    ordering: ImageLibraryOrdering;
    pageSize: number;
    cursor?: string;
    page?: number;
    takenBefore?: string;
    albumId?: number;
};

const searchNameOrContentMatch = 'all';

const repeatedKeysSerializer = { indexes: null };

const emptyToUndefined = (value: string) => (value === '' ? undefined : value);
const listToUndefined = <T>(values: T[]) => (values.length === 0 ? undefined : values);

const buildFilterParams = (filters: ImageLibraryFilters) => ({
    q: emptyToUndefined(filters.nameQuery),
    content: emptyToUndefined(filters.nameQuery),
    match: filters.nameQuery === '' ? undefined : searchNameOrContentMatch,
    category: listToUndefined(filters.categories),
    starred: filters.isStarredOnly ? true : undefined,
    format: listToUndefined(filters.formats),
    camera: emptyToUndefined(filters.camera),
    taken_from: emptyToUndefined(filters.takenFrom),
    taken_to: emptyToUndefined(filters.takenTo),
    folder: emptyToUndefined(filters.folder),
});

export const getImageFiles = async (
    page: number,
    pageSize: number,
    groupBy: ImageGroupBy
): Promise<Pagination<IImageData>> => {
    const response = await apiBase.get<Pagination<IImageData>>('/files/images', {
        params: { page, page_size: pageSize, group_by: groupBy },
    });
    return response.data;
};

export const getImageLibraryPage = async ({
    filters,
    ordering,
    pageSize,
    cursor,
    page,
    takenBefore,
    albumId,
}: ImageLibraryPageRequest): Promise<ImageLibraryPage> => {
    const listingUrl = albumId === undefined ? '/image/library' : `/image/albums/${albumId}/items`;
    const response = await apiBase.get<ImageLibraryPage>(listingUrl, {
        params: {
            ...buildFilterParams(filters),
            sort: ordering.sort,
            order: ordering.order,
            page_size: pageSize,
            cursor: emptyToUndefined(cursor ?? ''),
            page,
            taken_before: emptyToUndefined(takenBefore ?? ''),
        },
        paramsSerializer: repeatedKeysSerializer,
    });
    return response.data;
};

export const defaultImageNeighborsCount = 20;

export const getImageLibraryNeighbors = async (
    fileId: number,
    filters: ImageLibraryFilters,
    count = defaultImageNeighborsCount
): Promise<ImageLibraryNeighbors> => {
    const response = await apiBase.get<ImageLibraryNeighbors>(
        `/image/library/neighbors/${fileId}`,
        {
            params: { ...buildFilterParams(filters), count },
            paramsSerializer: repeatedKeysSerializer,
        }
    );
    return {
        before: response.data?.before ?? [],
        after: response.data?.after ?? [],
    };
};

export const getImageLibraryCount = async (filters: ImageLibraryFilters): Promise<number> => {
    const response = await apiBase.get<ImageLibraryCount>('/image/library/count', {
        params: buildFilterParams(filters),
        paramsSerializer: repeatedKeysSerializer,
    });
    return response.data.total;
};

export const getImageLibraryTimeline = async (
    filters: ImageLibraryFilters
): Promise<ImageTimelineBucket[]> => {
    const response = await apiBase.get<ImageTimelineBucket[]>('/image/library/timeline', {
        params: buildFilterParams(filters),
        paramsSerializer: repeatedKeysSerializer,
    });
    return response.data ?? [];
};

export const getImageCameraFacets = async (
    filters: ImageLibraryFilters
): Promise<ImageCameraFacet[]> => {
    const response = await apiBase.get<ImageCameraFacet[]>('/image/library/facets/cameras', {
        params: buildFilterParams(filters),
        paramsSerializer: repeatedKeysSerializer,
    });
    return Array.isArray(response.data) ? response.data : [];
};

export const getImageFormatFacets = async (
    filters: ImageLibraryFilters
): Promise<ImageFormatFacet[]> => {
    const response = await apiBase.get<ImageFormatFacet[]>('/image/library/facets/formats', {
        params: buildFilterParams(filters),
        paramsSerializer: repeatedKeysSerializer,
    });
    return Array.isArray(response.data) ? response.data : [];
};

export const getImageMetadataSummary = async (fileId: number): Promise<ImageMetadataSummary> => {
    const response = await apiBase.get<ImageMetadataSummary>(`/image/metadata/${fileId}`);
    return response.data;
};

export const getImageLibraryFolders = async (
    parentPath: string,
    page: number,
    pageSize: number
): Promise<Pagination<ImageLibraryFolder>> => {
    const response = await apiBase.get<Pagination<ImageLibraryFolder>>('/image/library/folders', {
        params: { parent: emptyToUndefined(parentPath), page, page_size: pageSize },
    });
    return response.data;
};

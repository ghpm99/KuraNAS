import { getApiV1BaseUrl } from '@/service/apiUrl';

export const GRID_THUMBNAIL_SIZE = 400;
export const FILMSTRIP_THUMBNAIL_SIZE = 160;
export const PREVIEW_IMAGE_SIZE = 1600;

const HIGH_DENSITY_MULTIPLIER = 2;

export const thumbnailUrl = (fileId: number, boxSize: number): string =>
    `${getApiV1BaseUrl()}/files/thumbnail/${fileId}?width=${boxSize}&height=${boxSize}`;

export const originalImageUrl = (fileId: number): string =>
    `${getApiV1BaseUrl()}/files/blob/${fileId}`;

export const thumbnailSrcSet = (fileId: number, boxSize: number): string => {
    const highDensitySize = boxSize * HIGH_DENSITY_MULTIPLIER;
    return `${thumbnailUrl(fileId, boxSize)} ${boxSize}w, ${thumbnailUrl(fileId, highDensitySize)} ${highDensitySize}w`;
};

export const gridThumbnailSizes = `${GRID_THUMBNAIL_SIZE}px`;

export const viewerImageUrl = (fileId: number, zoom: number): string =>
    zoom > 1 ? originalImageUrl(fileId) : thumbnailUrl(fileId, PREVIEW_IMAGE_SIZE);

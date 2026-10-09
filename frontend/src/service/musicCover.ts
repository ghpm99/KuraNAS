import { getApiV1BaseUrl } from './apiUrl';

export const getTrackCoverUrl = (fileId: number, size: number): string =>
    `${getApiV1BaseUrl()}/music/tracks/${fileId}/cover?size=${size}`;

export const getAlbumCoverUrl = (albumKey: string, size: number): string =>
    `${getApiV1BaseUrl()}/music/library/albums/${encodeURIComponent(albumKey)}/cover?size=${size}`;

const MEDIA_SESSION_ARTWORK_SIZES = [96, 256, 512];

export const getTrackCoverArtwork = (fileId: number): MediaImage[] =>
    MEDIA_SESSION_ARTWORK_SIZES.map((size) => ({
        src: getTrackCoverUrl(fileId, size),
        sizes: `${size}x${size}`,
        type: 'image/jpeg',
    }));

import { getApiV1BaseUrl } from '@/service/apiUrl';

export type CanPlayType = (mimeType: string) => string;

const REMUX_CANDIDATE_MIME_TYPE_BY_EXTENSION: Record<string, string> = {
    mkv: 'video/x-matroska',
    avi: 'video/x-msvideo',
    wmv: 'video/x-ms-wmv',
    flv: 'video/x-flv',
    mpg: 'video/mpeg',
    mpeg: 'video/mpeg',
    vob: 'video/mpeg',
    m2ts: 'video/mp2t',
};

export const buildDirectVideoStreamUrl = (videoId: number): string =>
    `${getApiV1BaseUrl()}/files/video-stream/${videoId}`;

export const buildRemuxVideoStreamUrl = (videoId: number, startSeconds: number): string => {
    const baseUrl = `${getApiV1BaseUrl()}/video/stream/${videoId}/remux`;
    return startSeconds > 0 ? `${baseUrl}?start=${startSeconds.toFixed(3)}` : baseUrl;
};

export const isContainerUnplayableByBrowser = (
    format: string | undefined,
    canPlayType: CanPlayType | undefined
): boolean => {
    if (!canPlayType) return false;
    const extension = (format ?? '').trim().replace(/^\./, '').toLowerCase();
    const mimeType = REMUX_CANDIDATE_MIME_TYPE_BY_EXTENSION[extension];
    if (!mimeType) return false;
    return canPlayType(mimeType) === '';
};

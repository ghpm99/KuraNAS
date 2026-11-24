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

const buildReencodedVideoStreamUrl = (
    route: 'remux' | 'transcode',
    videoId: number,
    startSeconds: number,
    audioStreamIndex: number | null
): string => {
    const baseUrl = `${getApiV1BaseUrl()}/video/stream/${videoId}/${route}`;
    const queryParams: string[] = [];
    if (startSeconds > 0) queryParams.push(`start=${startSeconds.toFixed(3)}`);
    if (audioStreamIndex !== null) queryParams.push(`audio_stream_index=${audioStreamIndex}`);
    return queryParams.length > 0 ? `${baseUrl}?${queryParams.join('&')}` : baseUrl;
};

export const buildRemuxVideoStreamUrl = (
    videoId: number,
    startSeconds: number,
    audioStreamIndex: number | null = null
): string => buildReencodedVideoStreamUrl('remux', videoId, startSeconds, audioStreamIndex);

export const buildTranscodeVideoStreamUrl = (
    videoId: number,
    startSeconds: number,
    audioStreamIndex: number | null = null
): string => buildReencodedVideoStreamUrl('transcode', videoId, startSeconds, audioStreamIndex);

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

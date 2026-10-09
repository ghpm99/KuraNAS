import { getApiV1BaseUrl } from '@/service/apiUrl';
import type { IMusicData } from '@/types/music';
import { getTrackDurationSeconds } from '@/utils/music';

export interface TranscodedStream {
    buildUrl: (startSeconds: number) => string;
    fallbackUrl: string;
    durationSeconds: number;
}

export interface TrackStreamSource {
    url: string;
    transcodedStream?: TranscodedStream;
}

export type CanPlayType = (mimeType: string) => string;

const TRANSCODE_TARGET_FORMAT = 'mp3';

const MIME_TYPE_BY_EXTENSION: Record<string, string> = {
    mp3: 'audio/mpeg',
    mpga: 'audio/mpeg',
    aac: 'audio/aac',
    m4a: 'audio/mp4',
    mp4: 'audio/mp4',
    alac: 'audio/mp4; codecs="alac"',
    ogg: 'audio/ogg',
    oga: 'audio/ogg',
    opus: 'audio/ogg; codecs="opus"',
    flac: 'audio/flac',
    wav: 'audio/wav',
    wave: 'audio/wav',
    aiff: 'audio/aiff',
    aif: 'audio/aiff',
    wma: 'audio/x-ms-wma',
    ape: 'audio/x-ape',
    wv: 'audio/x-wavpack',
};

export const buildDirectStreamUrl = (trackId: number): string =>
    `${getApiV1BaseUrl()}/files/stream/${trackId}`;

export const buildTranscodedStreamUrl = (trackId: number, startSeconds: number): string => {
    const baseUrl = `${getApiV1BaseUrl()}/music/tracks/${trackId}/stream?format=${TRANSCODE_TARGET_FORMAT}`;
    return startSeconds > 0 ? `${baseUrl}&start=${startSeconds.toFixed(3)}` : baseUrl;
};

export const getAudioMimeType = (format: string | undefined): string | undefined => {
    const extension = (format ?? '').trim().replace(/^\./, '').toLowerCase();
    return MIME_TYPE_BY_EXTENSION[extension];
};

export const isFormatPlayable = (format: string | undefined, canPlayType: CanPlayType): boolean => {
    const mimeType = getAudioMimeType(format);
    if (!mimeType) return true;
    return canPlayType(mimeType) !== '';
};

export const resolveTrackStreamSource = (
    track: Pick<IMusicData, 'id' | 'format' | 'metadata'>,
    canPlayType: CanPlayType
): TrackStreamSource => {
    const directUrl = buildDirectStreamUrl(track.id);
    if (isFormatPlayable(track.format, canPlayType)) {
        return { url: directUrl };
    }
    return {
        url: buildTranscodedStreamUrl(track.id, 0),
        transcodedStream: {
            buildUrl: (startSeconds) => buildTranscodedStreamUrl(track.id, startSeconds),
            fallbackUrl: directUrl,
            durationSeconds: getTrackDurationSeconds(track.metadata),
        },
    };
};

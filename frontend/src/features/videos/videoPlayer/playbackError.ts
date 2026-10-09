export type PlaybackErrorKind = 'aborted' | 'network' | 'decode' | 'unsupported' | 'unknown';

const MEDIA_ERROR_KINDS: Record<number, PlaybackErrorKind> = {
    1: 'aborted',
    2: 'network',
    3: 'decode',
    4: 'unsupported',
};

export const PLAYBACK_ERROR_MESSAGE_KEYS: Record<PlaybackErrorKind, string> = {
    aborted: 'VIDEO_ERROR_ABORTED',
    network: 'VIDEO_ERROR_NETWORK',
    decode: 'VIDEO_ERROR_DECODE',
    unsupported: 'VIDEO_ERROR_UNSUPPORTED',
    unknown: 'VIDEO_ERROR_UNKNOWN',
};

export const getPlaybackErrorKindFromMediaErrorCode = (
    mediaErrorCode: number | undefined
): PlaybackErrorKind => MEDIA_ERROR_KINDS[mediaErrorCode ?? 0] ?? 'unknown';

const IGNORED_PLAY_REJECTION_NAMES = ['NotAllowedError', 'AbortError'];

export const getPlaybackErrorKindFromPlayRejection = (
    rejection: unknown
): PlaybackErrorKind | null => {
    const rejectionName = (rejection as { name?: string } | null)?.name;
    if (rejectionName && IGNORED_PLAY_REJECTION_NAMES.includes(rejectionName)) {
        return null;
    }
    return rejectionName === 'NotSupportedError' ? 'unsupported' : 'unknown';
};

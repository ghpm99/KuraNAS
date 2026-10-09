import {
    getPlaybackErrorKindFromMediaErrorCode,
    getPlaybackErrorKindFromPlayRejection,
} from './playbackError';

describe('videoPlayer/playbackError', () => {
    it.each([
        [1, 'aborted'],
        [2, 'network'],
        [3, 'decode'],
        [4, 'unsupported'],
        [99, 'unknown'],
        [undefined, 'unknown'],
    ])('maps media error code %s to %s', (mediaErrorCode, expectedKind) => {
        expect(getPlaybackErrorKindFromMediaErrorCode(mediaErrorCode)).toBe(expectedKind);
    });

    it('ignores autoplay and abort rejections', () => {
        expect(getPlaybackErrorKindFromPlayRejection({ name: 'NotAllowedError' })).toBeNull();
        expect(getPlaybackErrorKindFromPlayRejection({ name: 'AbortError' })).toBeNull();
    });

    it('classifies other rejections', () => {
        expect(getPlaybackErrorKindFromPlayRejection({ name: 'NotSupportedError' })).toBe(
            'unsupported'
        );
        expect(getPlaybackErrorKindFromPlayRejection(new Error('x'))).toBe('unknown');
        expect(getPlaybackErrorKindFromPlayRejection(null)).toBe('unknown');
    });
});
